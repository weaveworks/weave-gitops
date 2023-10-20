package server

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/weaveworks/weave-gitops/core/fluxsync"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cs *coreServer) SyncFluxObject(ctx context.Context, msg *pb.SyncFluxObjectRequest) (*pb.SyncFluxObjectResponse, error) {
	principal := auth.Principal(ctx)
	respErrors := multierror.Error{}

	for _, sync := range msg.Objects {
		clustersClient, err := cs.clustersManager.GetImpersonatedClient(ctx, principal)
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

		gvk, err := cs.primaryKinds.Lookup(sync.Kind)
		if err != nil {
			respErrors = *multierror.Append(fmt.Errorf("looking up GVK for %q: %w", sync.Kind, err), respErrors.Errors...)
			continue
		}

		_, obj, err := fluxsync.ToReconcileable(*gvk)
		if err != nil {
			respErrors = *multierror.Append(fmt.Errorf("error converting to object: %w", err), respErrors.Errors...)
			continue
		}

		if err := c.Get(ctx, key, obj.AsClientObject()); err != nil {
			respErrors = *multierror.Append(fmt.Errorf("error getting object: %w", err), respErrors.Errors...)
			continue
		}

		automation, isAutomation := obj.(fluxsync.Automation)
		if msg.WithSource && isAutomation {
			sourceRef := automation.SourceRef()

			sourceGVK, err := cs.primaryKinds.Lookup(sourceRef.Kind())
			if err != nil {
				return nil, err
			}

			_, sourceObj, err := fluxsync.ToReconcileable(*sourceGVK)
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

			if err := fluxsync.RequestReconciliation(ctx, c, sourceKey, sourceGvk); err != nil {
				respErrors = *multierror.Append(fmt.Errorf("requesting source reconciliation: %w", err), respErrors.Errors...)
				continue
			}

			if err := fluxsync.WaitForSync(ctx, c, sourceKey, sourceObj); err != nil {
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

		if err := fluxsync.RequestReconciliation(ctx, c, key, *gvk); err != nil {
			respErrors = *multierror.Append(fmt.Errorf("requesting reconciliation: %w", err), respErrors.Errors...)
			continue
		}

		if err := fluxsync.WaitForSync(ctx, c, key, obj); err != nil {
			respErrors = *multierror.Append(fmt.Errorf("syncing automation: %w", err), respErrors.Errors...)
			continue
		}
	}

	return &pb.SyncFluxObjectResponse{}, respErrors.ErrorOrNil()
}
