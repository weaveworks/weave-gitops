package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/server/internal"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var k8sPollInterval = 2 * time.Second
var k8sTimeout = 1 * time.Minute

func (cs *coreServer) SyncAutomation(ctx context.Context, msg *pb.SyncAutomationRequest) (*pb.SyncAutomationResponse, error) {
	c, err := clustersmngr.ClientFromCtx(ctx).Scoped(msg.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("getting cluster client: %w", err)
	}

	key := client.ObjectKey{
		Name:      msg.Name,
		Namespace: msg.Namespace,
	}

	obj := getAutomation(msg.Kind)

	if err := c.Get(ctx, key, obj.AsClientObject()); err != nil {
		return nil, err
	}

	if msg.WithSource {
		sourceRef := obj.SourceRef()

		_, sourceObj, err := internal.ToReconcileableSource(kindToSourceType(sourceRef.Kind()))

		if err != nil {
			return nil, fmt.Errorf("getting source type for %q: %w", sourceRef.Kind(), err)
		}

		ns := sourceRef.Namespace()

		// sourceRef.Namespace is an optional field in flux
		// From the flux type reference:
		// "Namespace of the referent, defaults to the namespace of the Kubernetes resource object that contains the reference."
		// https://github.com/fluxcd/kustomize-controller/blob/4da17e1ffb9c2b9e057ff3440f66500394a4f765/api/v1beta2/reference_types.go#L37
		if ns == "" {
			ns = msg.Namespace
		}

		name := types.NamespacedName{
			Name:      sourceRef.Name(),
			Namespace: ns,
		}

		src := sourceObj.AsClientObject()

		if err := c.Get(ctx, name, src); err != nil {
			return nil, fmt.Errorf("getting source: %w", err)
		}

		sourceGvk := sourceObj.GroupVersionKind()

		if err := requestReconciliation(ctx, c, name, sourceGvk); err != nil {
			return nil, fmt.Errorf("request source reconciliation: %w", err)
		}

		if err := waitForSync(ctx, c, key, sourceObj); err != nil {
			return nil, fmt.Errorf("syncing source; %w", err)
		}
	}

	gvk := obj.GroupVersionKind()

	if err := requestReconciliation(ctx, c, key, gvk); err != nil {
		return nil, fmt.Errorf("requesting reconciliation: %w", err)
	}

	if err := waitForSync(ctx, c, key, obj); err != nil {
		return nil, fmt.Errorf("syncing automation; %w", err)
	}

	return &pb.SyncAutomationResponse{}, nil
}

func getAutomation(kind pb.AutomationKind) internal.Automation {
	switch kind {
	case pb.AutomationKind_KustomizationAutomation:
		return &internal.KustomizationAdapter{Kustomization: &kustomizev1.Kustomization{}}
	case pb.AutomationKind_HelmReleaseAutomation:
		return &internal.HelmReleaseAdapter{HelmRelease: &helmv2.HelmRelease{}}
	}

	return nil
}

func kindToSourceType(kind string) pb.SourceRef_SourceKind {
	switch kind {
	case "GitRepository":
		return pb.SourceRef_GitRepository
	case "Bucket":
		return pb.SourceRef_Bucket

	case "HelmRepository":
		return pb.SourceRef_HelmRepository

	case "HelmChart":
		return pb.SourceRef_HelmChart
	}

	return -1
}

// requestReconciliation sets the annotations of an object so that the flux controller(s) will force a reconciliation.
// Take straight from the flux CLI source:
// https://github.com/fluxcd/flux2/blob/cb53243fc11de81de3a34616d14322d66573aa65/cmd/flux/reconcile.go#L155
func requestReconciliation(ctx context.Context, k client.Client, name types.NamespacedName, gvk schema.GroupVersionKind) error {
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

func checkResourceSync(ctx context.Context, c client.Client, name types.NamespacedName, obj internal.Reconcilable, lastReconcile string) func() (bool, error) {
	return func() (bool, error) {
		err := c.Get(ctx, name, obj.AsClientObject())
		if err != nil {
			return false, err
		}

		return obj.GetLastHandledReconcileRequest() != lastReconcile, nil
	}
}

func waitForSync(ctx context.Context, c client.Client, key types.NamespacedName, obj internal.Reconcilable) error {
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
