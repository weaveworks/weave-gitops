package run

import (
	"context"
	"strings"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func requestReconciliation(ctx context.Context, kubeClient client.Client, namespacedName types.NamespacedName, gvk schema.GroupVersionKind) (string, error) {
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

// isPodStatusConditionPresentAndEqual returns true when conditionType is present and equal to status.
func isPodStatusConditionPresentAndEqual(conditions []corev1.PodCondition, conditionType corev1.PodConditionType, status corev1.ConditionStatus) bool {
	for _, condition := range conditions {
		if condition.Type == conditionType {
			return condition.Status == status
		}
	}

	return false
}
