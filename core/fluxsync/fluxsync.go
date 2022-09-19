package fluxsync

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fluxcd/pkg/apis/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var k8sPollInterval = 2 * time.Second
var k8sTimeout = 1 * time.Minute

// RequestReconciliation sets the annotations of an object so that the flux controller(s) will force a reconciliation.
// Take straight from the flux CLI source:
// https://github.com/fluxcd/flux2/blob/cb53243fc11de81de3a34616d14322d66573aa65/cmd/flux/reconcile.go#L155
func RequestReconciliation(ctx context.Context, k client.Client, name client.ObjectKey, gvk schema.GroupVersionKind) error {
	return retry.RetryOnConflict(retry.DefaultBackoff, func() (err error) {
		object := &metav1.PartialObjectMetadata{}
		object.SetGroupVersionKind(gvk)
		object.SetName(name.Name)
		object.SetNamespace(name.Namespace)
		if err := k.Get(ctx, name, object); err != nil {
			return err
		}
		patch := client.MergeFrom(object.DeepCopy())
		if ann := object.GetAnnotations(); ann == nil {
			object.SetAnnotations(map[string]string{
				meta.ReconcileRequestAnnotation: time.Now().Format(time.RFC3339Nano),
			})
		} else {
			ann[meta.ReconcileRequestAnnotation] = time.Now().Format(time.RFC3339Nano)
			object.SetAnnotations(ann)
		}
		return k.Patch(ctx, object, patch)
	})
}

// WaitForSync polls the k8s API until the resources is sync'd, and times out eventually.
func WaitForSync(ctx context.Context, c client.Client, key client.ObjectKey, obj Reconcilable) error {
	if err := wait.PollImmediate(
		k8sPollInterval,
		k8sTimeout,
		checkResourceSync(ctx, c, key, obj, obj.GetLastHandledReconcileRequest()),
	); err != nil {
		if errors.Is(err, wait.ErrWaitTimeout) {
			return errors.New("Sync request timed out. The sync operation may still be in progress.")
		}

		return fmt.Errorf("syncing resource: %w", err)
	}

	return nil
}

func checkResourceSync(ctx context.Context, c client.Client, name client.ObjectKey, obj Reconcilable, lastReconcile string) func() (bool, error) {
	return func() (bool, error) {
		err := c.Get(ctx, name, obj.AsClientObject())
		if err != nil {
			return false, err
		}

		return obj.GetLastHandledReconcileRequest() != lastReconcile, nil
	}
}
