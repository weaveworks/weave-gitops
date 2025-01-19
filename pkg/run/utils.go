package run

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/pkg/kube"
)

func RequestReconciliation(ctx context.Context, kubeClient client.Client, namespacedName types.NamespacedName, gvk schema.GroupVersionKind) (string, error) {
	requestAt := time.Now().Format(time.RFC3339Nano)

	return requestAt, retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		object := &metav1.PartialObjectMetadata{}
		object.SetGroupVersionKind(gvk)
		object.SetName(namespacedName.Name)
		object.SetNamespace(namespacedName.Namespace)
		if err := kubeClient.Get(ctx, namespacedName, object); err != nil {
			return err
		}
		patch := client.MergeFrom(object.DeepCopy())
		if ann := object.GetAnnotations(); ann == nil {
			object.SetAnnotations(map[string]string{
				meta.ReconcileRequestAnnotation: requestAt,
			})
		} else {
			ann[meta.ReconcileRequestAnnotation] = requestAt
			object.SetAnnotations(ann)
		}
		err = kubeClient.Patch(ctx, object, patch)
		return err
	})
}

// IsLocalCluster checks if it's a local cluster. See https://skaffold.dev/docs/environment/local-cluster/
func IsLocalCluster(kubeClient *kube.KubeHTTP) bool {
	const (
		// cluster context prefixes
		kindPrefix = "kind-"
		k3dPrefix  = "k3d-"

		// cluster context names
		minikube         = "minikube"
		dockerForDesktop = "docker-for-desktop"
		dockerDesktop    = "docker-desktop"
	)

	if strings.HasPrefix(kubeClient.ClusterName, kindPrefix) ||
		strings.HasPrefix(kubeClient.ClusterName, k3dPrefix) ||
		kubeClient.ClusterName == minikube ||
		kubeClient.ClusterName == dockerForDesktop ||
		kubeClient.ClusterName == dockerDesktop {
		return true
	} else {
		return false
	}
}

func GetPodFromResourceDescription(ctx context.Context, kubeClient client.Client, namespacedName types.NamespacedName, kind string, podLabels map[string]string) (*corev1.Pod, error) {
	if namespacedName.Name == "" {
		if len(podLabels) == 0 {
			return nil, fmt.Errorf("no pod name or labels provided")
		}

		// list all pods in the provided namespace and return the first pod with matching labels
		podList := &corev1.PodList{}
		if err := kubeClient.List(ctx, podList,
			client.MatchingLabelsSelector{
				Selector: labels.Set(podLabels).AsSelector(),
			},
			client.InNamespace(namespacedName.Namespace),
		); err != nil {
			return nil, err
		}

		if len(podList.Items) == 0 {
			return nil, fmt.Errorf("no pods with labels: %v found in the namespace: %s", podLabels, namespacedName.Namespace)
		}

		return &podList.Items[0], nil
	}

	switch kind {
	case "pod":
		pod := &corev1.Pod{}
		if err := kubeClient.Get(ctx, namespacedName, pod); err != nil {
			return nil, err
		}

		return pod, nil
	case "service":
		svc := &corev1.Service{}
		if err := kubeClient.Get(ctx, namespacedName, svc); err != nil {
			return nil, fmt.Errorf("error getting service: %w, namespaced Name: %v", err, namespacedName)
		}

		// list pods of the service "svc" by selector in a specific namespace using the controller-runtime client
		podList := &corev1.PodList{}
		if err := kubeClient.List(ctx, podList,
			client.MatchingLabelsSelector{
				Selector: labels.Set(svc.Spec.Selector).AsSelector(),
			},
			client.InNamespace(svc.Namespace),
		); err != nil {
			return nil, err
		}

		if len(podList.Items) == 0 {
			return nil, ErrNoPodsForService
		}

		for _, pod := range podList.Items {
			if pod.Status.Phase == corev1.PodRunning {
				return &pod, nil
			}
		}

		return nil, ErrNoRunningPodsForService
	case "deployment":
		deployment := &appsv1.Deployment{}
		if err := kubeClient.Get(ctx, namespacedName, deployment); err != nil {
			return nil, fmt.Errorf("error getting deployment: %w, namespaced Name: %v", err, namespacedName)
		}

		// list pods of the deployment "deployment" by selector in a specific namespace using the controller-runtime client
		podList := &corev1.PodList{}
		if err := kubeClient.List(ctx, podList,
			client.MatchingLabelsSelector{
				Selector: labels.Set(deployment.Spec.Selector.MatchLabels).AsSelector(),
			},
			client.InNamespace(deployment.Namespace),
		); err != nil {
			return nil, err
		}

		if len(podList.Items) == 0 {
			return nil, ErrNoPodsForDeployment
		}

		for _, pod := range podList.Items {
			if pod.Status.Phase == corev1.PodRunning {
				return &pod, nil
			}
		}

		return nil, ErrNoRunningPodsForDeployment
	default:
		return nil, errors.New("unsupported spec kind")
	}
}
