package server

import (
	"context"
	"errors"
	"fmt"

	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/core/fluxsync"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func (cs *coreServer) SyncFluxObject(ctx context.Context, msg *pb.SyncFluxObjectRequest) (*pb.SyncFluxObjectResponse, error) {
	principal := auth.Principal(ctx)
	var syncErr error

	for _, sync := range msg.Objects {
		clustersClient, err := cs.clustersManager.GetImpersonatedClient(ctx, principal)
		if err != nil {
			syncErr = errors.Join(syncErr, fmt.Errorf("error getting impersonating client: %w", err))
			continue
		}

		c, err := clustersClient.Scoped(sync.ClusterName)
		if err != nil {
			syncErr = errors.Join(syncErr, fmt.Errorf("getting cluster client: %w", err))
			continue
		}

		key := client.ObjectKey{
			Name:      sync.Name,
			Namespace: sync.Namespace,
		}

		gvk, err := cs.primaryKinds.Lookup(sync.Kind)
		if err != nil {
			syncErr = errors.Join(syncErr, fmt.Errorf("looking up GVK for %q: %w", sync.Kind, err))
			continue
		}

		obj := fluxsync.ToReconcileable(*gvk)
		if err := c.Get(ctx, key, obj.AsClientObject()); err != nil {
			syncErr = errors.Join(syncErr, fmt.Errorf("error getting object: %w", err))
			continue
		}

		automation, isAutomation := obj.(fluxsync.Automation)
		if msg.WithSource && isAutomation {
			sourceRef := automation.SourceRef()

			sourceGVK, err := cs.primaryKinds.Lookup(sourceRef.Kind())
			if err != nil {
				return nil, err
			}

			sourceObj := fluxsync.ToReconcileable(*sourceGVK)
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

			if err := fluxsync.RequestReconciliation(ctx, c, sourceKey, sourceGvk); err != nil {
				syncErr = errors.Join(syncErr, fmt.Errorf("requesting source reconciliation: %w", err))
				continue
			}

			if err := fluxsync.WaitForSync(ctx, c, sourceKey, sourceObj); err != nil {
				syncErr = errors.Join(syncErr, fmt.Errorf("syncing source: %w", err))
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

		if err := fluxsync.RequestReconciliation(ctx, c, key, *gvk); err != nil {
			syncErr = errors.Join(syncErr, fmt.Errorf("requesting reconciliation: %w", err))
			continue
		}

		if err := fluxsync.WaitForSync(ctx, c, key, obj); err != nil {
			syncErr = errors.Join(syncErr, fmt.Errorf("syncing automation: %w", err))
			continue
		}
	}

	return &pb.SyncFluxObjectResponse{}, syncErr
}
