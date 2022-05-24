package server

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/server/internal"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cs *coreServer) ToggleSuspendResource(ctx context.Context, msg *pb.ToggleSuspendResourceRequest) (*pb.ToggleSuspendResourceResponse, error) {
	c, err := clustersmngr.ClientFromCtx(ctx).Scoped(msg.ClusterName)
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

	if err := c.Get(ctx, key, obj.AsClientObject()); err != nil {
		return nil, fmt.Errorf("getting reconcilable object: %w", err)
	}

	patch := client.MergeFrom(obj.DeepCopyClientObject())

	obj.SetSuspended(msg.Suspend)

	if err := c.Patch(ctx, obj.AsClientObject(), patch); err != nil {
		return nil, fmt.Errorf("patching object: %w", err)
	}

	return &pb.ToggleSuspendResourceResponse{}, nil
}

func getReconcilableObject(kind pb.FluxObjectKind) (internal.Reconcilable, error) {
	_, s, err := internal.ToReconcileable(kind)

	return s, err
}
