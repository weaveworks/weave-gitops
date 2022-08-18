package server

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/core/server/internal"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cs *coreServer) ToggleSuspendResource(ctx context.Context, msg *pb.ToggleSuspendResourceRequest) (*pb.ToggleSuspendResourceResponse, error) {
	principal := auth.Principal(ctx)

	clustersClient, err := cs.clientsFactory.GetImpersonatedClientForCluster(ctx, principal, msg.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	c, err := clustersClient.Scoped(msg.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("getting cluster client: %w", err)
	}

	key := client.ObjectKey{
		Name:      msg.Name,
		Namespace: msg.Namespace,
	}

	obj, err := getReconcilableObject(msg.Kind)
	if err != nil {
		return nil, fmt.Errorf("converting to reconcilable source: %w", err)
	}

	log := cs.logger.WithValues(
		"user", principal.ID,
		"kind", obj.GroupVersionKind().Kind,
		"name", msg.Name,
		"namespace", msg.Namespace,
	)

	if err := c.Get(ctx, key, obj.AsClientObject()); err != nil {
		return nil, fmt.Errorf("getting reconcilable object: %w", err)
	}

	patch := client.MergeFrom(obj.DeepCopyClientObject())

	obj.SetSuspended(msg.Suspend)

	if msg.Suspend {
		log.Info("Suspending resource")
	} else {
		log.Info("Resuming resource")
	}

	if err := c.Patch(ctx, obj.AsClientObject(), patch); err != nil {
		return nil, fmt.Errorf("patching object: %w", err)
	}

	return &pb.ToggleSuspendResourceResponse{}, nil
}

func getReconcilableObject(kind pb.FluxObjectKind) (internal.Reconcilable, error) {
	_, s, err := internal.ToReconcileable(kind)

	return s, err
}
