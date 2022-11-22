package watch

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/run"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	RunDevBucketName   = "run-dev-bucket"
	RunDevKsName       = "run-dev-ks"
	GitOpsRunNamespace = "gitops-run"
)

var (
	// The variables below are to be set by flags passed to `go build`.
	// Examples: -X run.DevBucketContainerImage=xxxxx

	DevBucketContainerImage = "ghcr.io/weaveworks/gitops-bucket-server@sha256:8fbb7534e772e14ea598d287a4b54a3f556416cac6621095ce45f78346fda78a"
)

// InstallDevBucketServer installs the dev bucket server, open port forwarding, and returns a function that can be used to the port forwarding.
func InstallDevBucketServer(ctx context.Context, log logger.Logger, kubeClient client.Client, config *rest.Config, devBucketPort int32) (func(), error) {
	var (
		err                error
		devBucketAppLabels = map[string]string{
			"app": RunDevBucketName,
		}
	)

	// create namespace
	devBucketNamespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: GitOpsRunNamespace,
		},
	}

	log.Actionf("Checking namespace %s ...", GitOpsRunNamespace)

	err = kubeClient.Get(ctx,
		client.ObjectKeyFromObject(&devBucketNamespace),
		&devBucketNamespace)

	if err != nil && apierrors.IsNotFound(err) {
		if err := kubeClient.Create(ctx, &devBucketNamespace); err != nil {
			log.Failuref("Error creating namespace %s: %v", GitOpsRunNamespace, err.Error())
			return nil, err
		} else {
			log.Successf("Created namespace %s", GitOpsRunNamespace)
		}
	} else if err == nil {
		log.Successf("Namespace %s already existed", GitOpsRunNamespace)
	}

	// create service
	devBucketService := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RunDevBucketName,
			Namespace: GitOpsRunNamespace,
			Labels:    devBucketAppLabels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name: RunDevBucketName,
					Port: devBucketPort,
				},
			},
			Selector: devBucketAppLabels,
		},
	}

	log.Actionf("Checking service %s/%s ...", GitOpsRunNamespace, RunDevBucketName)

	err = kubeClient.Get(ctx,
		client.ObjectKeyFromObject(&devBucketService),
		&devBucketService)

	if err != nil && apierrors.IsNotFound(err) {
		if err := kubeClient.Create(ctx, &devBucketService); err != nil {
			log.Failuref("Error creating service %s/%s: %v", GitOpsRunNamespace, RunDevBucketName, err.Error())
			return nil, err
		} else {
			log.Successf("Created service %s/%s", GitOpsRunNamespace, RunDevBucketName)
		}
	} else if err == nil {
		log.Successf("Service %s/%s already existed", GitOpsRunNamespace, RunDevBucketName)
	}

	// create deployment
	replicas := int32(1)
	devBucketDeployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      RunDevBucketName,
			Namespace: GitOpsRunNamespace,
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
							Name:  RunDevBucketName,
							Image: DevBucketContainerImage,
							Env: []corev1.EnvVar{
								{Name: "MINIO_ROOT_USER", Value: "user"},
								{Name: "MINIO_ROOT_PASSWORD", Value: "doesn't matter"},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: devBucketPort,
									HostPort:      devBucketPort,
								},
							},
							Args: []string{strconv.Itoa(int(devBucketPort))},
						},
					},
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
		},
	}

	log.Actionf("Checking deployment %s/%s ...", GitOpsRunNamespace, RunDevBucketName)

	err = kubeClient.Get(ctx,
		client.ObjectKeyFromObject(&devBucketDeployment),
		&devBucketDeployment)

	if err != nil && apierrors.IsNotFound(err) {
		if err := kubeClient.Create(ctx, &devBucketDeployment); err != nil {
			log.Failuref("Error creating deployment %s/%s: %v", GitOpsRunNamespace, RunDevBucketName, err.Error())
			return nil, err
		} else {
			log.Successf("Created deployment %s/%s", GitOpsRunNamespace, RunDevBucketName)
		}
	} else if err == nil {
		log.Successf("Deployment %s/%s already existed", GitOpsRunNamespace, RunDevBucketName)
	}

	log.Actionf("Waiting for deployment %s to be ready ...", RunDevBucketName)

	if err := wait.ExponentialBackoff(wait.Backoff{
		Duration: 1 * time.Second,
		Factor:   2,
		Jitter:   1,
		Steps:    10,
	}, func() (done bool, err error) {
		d := devBucketDeployment.DeepCopy()
		if err := kubeClient.Get(ctx, client.ObjectKeyFromObject(d), d); err != nil {
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

	specMap := &PortForwardSpec{
		Name:          RunDevBucketName,
		Namespace:     GitOpsRunNamespace,
		Kind:          "service",
		HostPort:      strconv.Itoa(int(devBucketPort)),
		ContainerPort: strconv.Itoa(int(devBucketPort)),
	}
	// get pod from specMap
	namespacedName := types.NamespacedName{Namespace: specMap.Namespace, Name: specMap.Name}

	pod, err := run.GetPodFromResourceDescription(ctx, namespacedName, specMap.Kind, kubeClient)
	if err != nil {
		log.Failuref("Error getting pod from specMap: %v", err)
	}

	if pod != nil {
		waitFwd := make(chan struct{}, 1)
		readyChannel := make(chan struct{})
		cancelPortFwd := func() {
			close(waitFwd)
		}

		log.Actionf("Port forwarding to pod %s/%s ...", pod.Namespace, pod.Name)

		go func() {
			if err := ForwardPort(log.L(), pod, config, specMap, waitFwd, readyChannel); err != nil {
				log.Failuref("Error forwarding port: %v", err)
			}
		}()
		<-readyChannel

		log.Successf("Port forwarding for %s is ready.", RunDevBucketName)

		return cancelPortFwd, nil
	}

	return nil, fmt.Errorf("pod not found")
}

// UninstallDevBucketServer deletes the dev-bucket namespace.
func UninstallDevBucketServer(ctx context.Context, log logger.Logger, kubeClient client.Client) error {
	// create namespace
	devBucketNamespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: GitOpsRunNamespace,
		},
	}

	log.Actionf("Removing namespace %s ...", GitOpsRunNamespace)

	if err := kubeClient.Delete(ctx, &devBucketNamespace); err != nil {
		log.Failuref("Cannot remove namespace %s", GitOpsRunNamespace)
		return err
	}

	log.Actionf("Waiting for namespace %s to be terminated ...", GitOpsRunNamespace)

	if err := wait.ExponentialBackoff(wait.Backoff{
		Duration: 1 * time.Second,
		Factor:   2,
		Jitter:   1,
		Steps:    10,
	}, func() (done bool, err error) {
		ns := devBucketNamespace.DeepCopy()
		if err := kubeClient.Get(ctx, client.ObjectKeyFromObject(ns), ns); err != nil {
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

	log.Successf("Namespace %s terminated", GitOpsRunNamespace)

	return nil
}
