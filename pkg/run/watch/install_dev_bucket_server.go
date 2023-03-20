package watch

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/run"
	"github.com/weaveworks/weave-gitops/pkg/run/constants"
	"github.com/weaveworks/weave-gitops/pkg/tls"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	// The variables below are to be set by flags passed to `go build`.
	// Examples: -X run.DevBucketContainerImage=xxxxx

	DevBucketContainerImage string
)

// InstallDevBucketServer installs the dev bucket server, open port forwarding, and returns a function that can be used to the port forwarding.
func InstallDevBucketServer(
	ctx context.Context,
	log logger.Logger,
	kubeClient client.Client,
	config *rest.Config,
	httpPort,
	httpsPort int32,
	accessKey,
	secretKey []byte) (func(), []byte, error) {
	var (
		err                error
		devBucketAppLabels = map[string]string{
			"app": constants.RunDevBucketName,
		}
	)

	// create namespace
	devBucketNamespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.GitOpsRunNamespace,
		},
	}

	log.Actionf("Checking namespace %s ...", constants.GitOpsRunNamespace)

	err = kubeClient.Get(ctx,
		client.ObjectKeyFromObject(&devBucketNamespace),
		&devBucketNamespace)

	if err != nil && apierrors.IsNotFound(err) {
		if err := kubeClient.Create(ctx, &devBucketNamespace); err != nil {
			log.Failuref("Error creating namespace %s: %v", constants.GitOpsRunNamespace, err.Error())
			return nil, nil, err
		} else {
			log.Successf("Created namespace %s", constants.GitOpsRunNamespace)
		}
	} else if err == nil {
		log.Successf("Namespace %s already existed", constants.GitOpsRunNamespace)
	}

	// create service
	devBucketService := corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunDevBucketName,
			Namespace: constants.GitOpsRunNamespace,
			Labels:    devBucketAppLabels,
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeClusterIP,
			Ports: []corev1.ServicePort{
				{
					Name: fmt.Sprintf("%s-http", constants.RunDevBucketName),
					Port: httpPort,
				},
				{
					Name: fmt.Sprintf("%s-https", constants.RunDevBucketName),
					Port: httpsPort,
				},
			},
			Selector: devBucketAppLabels,
		},
	}

	log.Actionf("Checking service %s/%s ...", constants.GitOpsRunNamespace, constants.RunDevBucketName)

	err = kubeClient.Get(ctx,
		client.ObjectKeyFromObject(&devBucketService),
		&devBucketService)

	if err != nil && apierrors.IsNotFound(err) {
		if err := kubeClient.Create(ctx, &devBucketService); err != nil {
			log.Failuref("Error creating service %s/%s: %v", constants.GitOpsRunNamespace, constants.RunDevBucketName, err.Error())
			return nil, nil, err
		} else {
			log.Successf("Created service %s/%s", constants.GitOpsRunNamespace, constants.RunDevBucketName)
		}
	} else if err == nil {
		log.Successf("Service %s/%s already existed", constants.GitOpsRunNamespace, constants.RunDevBucketName)
	}

	credentialsSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: constants.GitOpsRunNamespace,
			Name:      constants.RunDevBucketCredentials,
		},
		Data: map[string][]byte{
			"accesskey": accessKey,
			"secretkey": secretKey,
		},
	}
	if err := kubeClient.Create(ctx, &credentialsSecret); err != nil {
		log.Failuref("Error creating credentials secret: %s", err.Error())
		return nil, nil, fmt.Errorf("failed creating credentials secret: %w", err)
	}

	cert, err := tls.GenerateSelfSignedCertificate("localhost", fmt.Sprintf("%s.%s.svc.cluster.local", devBucketService.Name, devBucketService.Namespace))
	if err != nil {
		err = fmt.Errorf("failed generating self-signed certificate for dev bucket server: %w", err)
		log.Failuref(err.Error())

		return nil, nil, err
	}

	certsSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "dev-bucket-server-certs",
			Namespace: constants.GitOpsRunNamespace,
			Labels:    devBucketAppLabels,
		},
		Data: map[string][]byte{
			"cert.pem": cert.Cert,
			"cert.key": cert.Key,
		},
	}
	if err := kubeClient.Create(ctx, certsSecret); err != nil {
		log.Failuref("Error creating Secret %s/%s: %v", certsSecret.Namespace, certsSecret.Name, err.Error())
		return nil, nil, err
	}

	// create deployment
	replicas := int32(1)
	devBucketDeployment := appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      constants.RunDevBucketName,
			Namespace: constants.GitOpsRunNamespace,
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
					Volumes: []corev1.Volume{{
						Name: "certs",
						VolumeSource: corev1.VolumeSource{
							Secret: &corev1.SecretVolumeSource{
								SecretName: "dev-bucket-server-certs",
							},
						},
					}},
					Containers: []corev1.Container{
						{
							Name:            constants.RunDevBucketName,
							Image:           DevBucketContainerImage,
							ImagePullPolicy: corev1.PullIfNotPresent,
							Env: []corev1.EnvVar{
								{Name: "MINIO_ROOT_USER", ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{Name: credentialsSecret.Name},
										Key:                  "accesskey",
									},
								}},
								{Name: "MINIO_ROOT_PASSWORD", ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{Name: credentialsSecret.Name},
										Key:                  "secretkey",
									},
								}},
							},
							Ports: []corev1.ContainerPort{
								{
									ContainerPort: httpPort,
									HostPort:      httpPort,
								},
								{
									ContainerPort: httpsPort,
									HostPort:      httpsPort,
								},
							},
							Args: []string{
								fmt.Sprintf("--http-port=%d", httpPort),
								fmt.Sprintf("--https-port=%d", httpsPort),
								"--cert-file=/tmp/certs/cert.pem",
								"--key-file=/tmp/certs/cert.key",
							},
							VolumeMounts: []corev1.VolumeMount{{
								Name:      "certs",
								MountPath: "/tmp/certs",
							}},
						},
					},
					RestartPolicy: corev1.RestartPolicyAlways,
				},
			},
		},
	}

	log.Actionf("Checking deployment %s/%s ...", constants.GitOpsRunNamespace, constants.RunDevBucketName)

	err = kubeClient.Get(ctx,
		client.ObjectKeyFromObject(&devBucketDeployment),
		&devBucketDeployment)

	if err != nil && apierrors.IsNotFound(err) {
		if err := kubeClient.Create(ctx, &devBucketDeployment); err != nil {
			log.Failuref("Error creating deployment %s/%s: %v", constants.GitOpsRunNamespace, constants.RunDevBucketName, err.Error())
			return nil, nil, err
		} else {
			log.Successf("Created deployment %s/%s", constants.GitOpsRunNamespace, constants.RunDevBucketName)
		}
	} else if err == nil {
		log.Successf("Deployment %s/%s already existed", constants.GitOpsRunNamespace, constants.RunDevBucketName)
	}

	log.Actionf("Waiting for deployment %s to be ready ...", constants.RunDevBucketName)

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
		Name:          constants.RunDevBucketName,
		Namespace:     constants.GitOpsRunNamespace,
		Kind:          "service",
		HostPort:      strconv.Itoa(int(httpsPort)),
		ContainerPort: strconv.Itoa(int(httpsPort)),
	}
	// get pod from specMap
	namespacedName := types.NamespacedName{Namespace: specMap.Namespace, Name: specMap.Name}

	pod, err := run.GetPodFromResourceDescription(ctx, kubeClient, namespacedName, specMap.Kind, nil)
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

		log.Successf("Port forwarding for %s is ready.", constants.RunDevBucketName)

		return cancelPortFwd, cert.Cert, nil
	}

	return nil, nil, fmt.Errorf("pod not found")
}

// UninstallDevBucketServer deletes the dev-bucket namespace.
func UninstallDevBucketServer(ctx context.Context, log logger.Logger, kubeClient client.Client) error {
	// create namespace
	devBucketNamespace := corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: constants.GitOpsRunNamespace,
		},
	}

	log.Actionf("Removing namespace %s ...", constants.GitOpsRunNamespace)

	if err := kubeClient.Delete(ctx, &devBucketNamespace); err != nil {
		log.Failuref("Cannot remove namespace %s", constants.GitOpsRunNamespace)
		return err
	}

	log.Actionf("Waiting for namespace %s to be terminated ...", constants.GitOpsRunNamespace)

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

	log.Successf("Namespace %s terminated", constants.GitOpsRunNamespace)

	return nil
}
