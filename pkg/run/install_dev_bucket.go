package run

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/run/forwarder"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	devBucket = "dev-bucket"
	port      = 9000
)

func InstallBucketServer(log logger.Logger, kubeClient *kube.KubeHTTP) error {
	var (
		err                error
		devBucketAppLabels = map[string]string{
			"app": devBucket,
		}
	)

	// create namespace
	devBucketNamespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: devBucket,
		},
	}

	log.Actionf("Checking namespace %s ...", devBucket)

	err = kubeClient.Get(context.Background(),
		client.ObjectKeyFromObject(&devBucketNamespace),
		&devBucketNamespace)

	if err != nil && apierrors.IsNotFound(err) {
		if err := kubeClient.Create(context.Background(), &devBucketNamespace); err != nil {
			log.Failuref("Error creating namespace %s: %v", devBucket, err.Error())
			return err
		} else {
			log.Successf("Created namespace %s", devBucket)
		}
	} else if err == nil {
		log.Successf("Namespace %s already existed", devBucket)
	}

	// create service
	devBucketService := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      devBucket,
			Namespace: devBucket,
			Labels:    devBucketAppLabels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{Name: devBucket, Port: port},
			},
			Selector: devBucketAppLabels,
		},
	}

	log.Actionf("Checking service %s/%s ...", devBucket, devBucket)

	err = kubeClient.Get(context.Background(),
		client.ObjectKeyFromObject(&devBucketService),
		&devBucketService)

	if err != nil && apierrors.IsNotFound(err) {
		if err := kubeClient.Create(context.Background(), &devBucketService); err != nil {
			log.Failuref("Error creating service %s/%s: %v", devBucket, devBucket, err.Error())
			return err
		} else {
			log.Successf("Created service %s/%s", devBucket, devBucket)
		}
	} else if err == nil {
		log.Successf("Service %s/%s already existed", devBucket, devBucket)
	}

	// create deployment
	replicas := int32(1)
	devBucketDeployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      devBucket,
			Namespace: devBucket,
			Labels:    devBucketAppLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: devBucketAppLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: devBucketAppLabels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  devBucket,
							Image: "ghcr.io/weaveworks/gitops-bucket-server:dev",
							Env: []corev1.EnvVar{
								{Name: "MINIO_ROOT_USER", Value: "user"},
								{Name: "MINIO_ROOT_PASSWORD", Value: "doesn't matter"},
							},
							Ports: []corev1.ContainerPort{
								{HostPort: port, ContainerPort: port},
							},
						},
					},
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
		},
	}

	log.Actionf("Checking deployment %s/%s ...", devBucket, devBucket)

	err = kubeClient.Get(context.Background(),
		client.ObjectKeyFromObject(&devBucketDeployment),
		&devBucketDeployment)

	if err != nil && apierrors.IsNotFound(err) {
		if err := kubeClient.Create(context.Background(), &devBucketDeployment); err != nil {
			log.Failuref("Error creating deployment %s/%s: %v", devBucket, devBucket, err.Error())
			return err
		} else {
			log.Successf("Created deployment %s/%s", devBucket, devBucket)
		}
	} else if err == nil {
		log.Successf("Deployment %s/%s already existed", devBucket, devBucket)
	}

	log.Actionf("Waiting for deployment %s to be ready ...", devBucket)

	if err := wait.ExponentialBackoff(wait.Backoff{
		Duration: 1 * time.Second,
		Factor:   2,
		Jitter:   1,
		Steps:    10,
	}, func() (done bool, err error) {
		d := devBucketDeployment.DeepCopy()
		if err := kubeClient.Get(context.Background(), client.ObjectKeyFromObject(d), d); err != nil {
			return false, err
		}
		// Confirm the state we are observing is for the current generation
		if d.Generation != d.Status.ObservedGeneration {
			return false, nil
		}

		if d.Status.ReadyReplicas == 1 {
			return true, nil
		}

		return false, nil
	}); err != nil {
		log.Failuref("Max retry exceeded waiting for deployment to be ready")
	}

	options := []*forwarder.Option{
		{
			LocalPort:   port,
			RemotePort:  port,
			ServiceName: devBucket,
			Namespace:   devBucket,
		},
	}

	log.Actionf("Forwarding dev-bucket port :%d to localhost ...", port)

	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	ret, err := forwarder.WithForwarders(ctx, options)
	if err != nil {
		return err
	}
	// defer to close the forwarding after the function ends
	defer ret.Close()

	// wait forwarding ready
	// the remote and local ports are listed
	log.Actionf("Waiting for dev-bucket port forwarding to be ready ...")

	if _, err := ret.Ready(); err != nil {
		return err
	}

	log.Successf("Port forwarding for dev-bucket is ready.")

	// wait for ctrl+C
	log.Waitingf("Press Ctrl+C to stop GitOps Run ...")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	<-sigs
	cancel()
	fmt.Println()

	log.Actionf("Removing namespace %s ...", devBucket)

	if err := kubeClient.Delete(context.Background(), &devBucketNamespace); err != nil {
		log.Failuref("Cannot remove namespace %s", devBucket)
		return err
	}

	log.Actionf("Waiting for namespace %s to be terminated ...", devBucket)

	if err := wait.ExponentialBackoff(wait.Backoff{
		Duration: 1 * time.Second,
		Factor:   2,
		Jitter:   1,
		Steps:    10,
	}, func() (done bool, err error) {
		ns := devBucketNamespace.DeepCopy()
		if err := kubeClient.Get(context.Background(), client.ObjectKeyFromObject(ns), ns); err != nil {
			if apierrors.IsNotFound(err) {
				return true, nil
			} else {
				return false, err
			}
		}
		return false, nil
	}); err != nil {
		log.Failuref("Max retry exceeded waiting for namespace to be deleted")
	}

	log.Successf("Namespace %s terminated", devBucket)

	return nil
}
