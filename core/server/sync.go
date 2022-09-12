package server

import (
	"context"
	"errors"
	"fmt"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/hashicorp/go-multierror"
	"github.com/weaveworks/weave-gitops/core/server/internal"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var k8sPollInterval = 2 * time.Second
var k8sTimeout = 1 * time.Minute

func (cs *coreServer) SyncFluxObject(ctx context.Context, msg *pb.SyncFluxObjectRequest) (*pb.SyncFluxObjectResponse, error) {
	principal := auth.Principal(ctx)
	respErrors := multierror.Error{}

	for _, sync := range msg.Objects {
		clustersClient, err := cs.clustersManager.GetImpersonatedClientForCluster(ctx, principal, sync.ClusterName)
		if err != nil {
			respErrors = *multierror.Append(fmt.Errorf("error getting impersonating client: %w", err), respErrors.Errors...)
			continue
		}

		c, err := clustersClient.Scoped(sync.ClusterName)
		if err != nil {
			respErrors = *multierror.Append(fmt.Errorf("getting cluster client: %w", err), respErrors.Errors...)
			continue
		}

		key := client.ObjectKey{
			Name:      sync.Name,
			Namespace: sync.Namespace,
		}

		obj, err := getFluxObject(sync.Kind)
		if err != nil {
			respErrors = *multierror.Append(fmt.Errorf("error converting to object: %w", err), respErrors.Errors...)
			continue
		}

		if err := c.Get(ctx, key, obj.AsClientObject()); err != nil {
			respErrors = *multierror.Append(fmt.Errorf("error getting object: %w", err), respErrors.Errors...)
			continue
		}

		automation, isAutomation := obj.(internal.Automation)
		if msg.WithSource && isAutomation {
			sourceRef := automation.SourceRef()

			_, sourceObj, err := internal.ToReconcileable(kindToSourceType(sourceRef.Kind()))

			if err != nil {
				respErrors = *multierror.Append(fmt.Errorf("getting source type for %q: %w", sourceRef.Kind(), err), respErrors.Errors...)
				continue
			}

			sourceNs := sourceRef.Namespace()

			// sourceRef.Namespace is an optional field in flux
			// From the flux type reference:
			// "Namespace of the referent, defaults to the namespace of the Kubernetes resource object that contains the reference."
			// https://github.com/fluxcd/kustomize-controller/blob/4da17e1ffb9c2b9e057ff3440f66500394a4f765/api/v1beta2/reference_types.go#L37
			if sourceNs == "" {
				sourceNs = sync.Namespace
			}

			sourceKey := client.ObjectKey{
				Name:      sourceRef.Name(),
				Namespace: sourceNs,
			}

			sourceGvk := sourceObj.GroupVersionKind()

			log := cs.logger.WithValues(
				"user", principal.ID,
				"kind", sourceRef.Kind(),
				"name", sourceRef.Name(),
				"namespace", sourceNs,
			)
			log.Info("Syncing resource")

			if err := requestReconciliation(ctx, c, sourceKey, sourceGvk); err != nil {
				respErrors = *multierror.Append(fmt.Errorf("requesting source reconciliation: %w", err), respErrors.Errors...)
				continue
			}

			if err := waitForSync(ctx, c, sourceKey, sourceObj); err != nil {
				respErrors = *multierror.Append(fmt.Errorf("syncing source: %w", err), respErrors.Errors...)
				continue
			}
		}

		log := cs.logger.WithValues(
			"user", principal.ID,
			"kind", obj.GroupVersionKind().Kind,
			"name", key.Name,
			"namespace", key.Namespace,
		)
		log.Info("Syncing resource")

		gvk := obj.GroupVersionKind()
		if err := requestReconciliation(ctx, c, key, gvk); err != nil {
			respErrors = *multierror.Append(fmt.Errorf("requesting reconciliation: %w", err), respErrors.Errors...)
			continue
		}

		if err := waitForSync(ctx, c, key, obj); err != nil {
			respErrors = *multierror.Append(fmt.Errorf("syncing automation: %w", err), respErrors.Errors...)
			continue
		}
	}

	return &pb.SyncFluxObjectResponse{}, respErrors.ErrorOrNil()
}

func getFluxObject(kind pb.FluxObjectKind) (internal.Reconcilable, error) {
	switch kind {
	case pb.FluxObjectKind_KindKustomization:
		return &internal.KustomizationAdapter{Kustomization: &kustomizev1.Kustomization{}}, nil
	case pb.FluxObjectKind_KindHelmRelease:
		return &internal.HelmReleaseAdapter{HelmRelease: &helmv2.HelmRelease{}}, nil

	case pb.FluxObjectKind_KindGitRepository:
		return &internal.GitRepositoryAdapter{GitRepository: &sourcev1.GitRepository{}}, nil
	case pb.FluxObjectKind_KindBucket:
		return &internal.BucketAdapter{Bucket: &sourcev1.Bucket{}}, nil
	case pb.FluxObjectKind_KindHelmRepository:
		return &internal.HelmRepositoryAdapter{HelmRepository: &sourcev1.HelmRepository{}}, nil
	case pb.FluxObjectKind_KindOCIRepository:
		return &internal.OCIRepositoryAdapter{OCIRepository: &sourcev1.OCIRepository{}}, nil
	}

	return nil, fmt.Errorf("not supported kind: %s", kind.String())
}

func kindToSourceType(kind string) pb.FluxObjectKind {
	switch kind {
	case "GitRepository":
		return pb.FluxObjectKind_KindGitRepository
	case "Bucket":
		return pb.FluxObjectKind_KindBucket
	case "HelmRepository":
		return pb.FluxObjectKind_KindHelmRepository
	case "OCIRepository":
		return pb.FluxObjectKind_KindOCIRepository
	case "HelmChart":
		return pb.FluxObjectKind_KindHelmChart
	}

	return -1
}

// requestReconciliation sets the annotations of an object so that the flux controller(s) will force a reconciliation.
// Take straight from the flux CLI source:
// https://github.com/fluxcd/flux2/blob/cb53243fc11de81de3a34616d14322d66573aa65/cmd/flux/reconcile.go#L155
func requestReconciliation(ctx context.Context, k client.Client, name client.ObjectKey, gvk schema.GroupVersionKind) error {
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

func checkResourceSync(ctx context.Context, c client.Client, name client.ObjectKey, obj internal.Reconcilable, lastReconcile string) func() (bool, error) {
	return func() (bool, error) {
		err := c.Get(ctx, name, obj.AsClientObject())
		if err != nil {
			return false, err
		}

		return obj.GetLastHandledReconcileRequest() != lastReconcile, nil
	}
}

func waitForSync(ctx context.Context, c client.Client, key client.ObjectKey, obj internal.Reconcilable) error {
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
