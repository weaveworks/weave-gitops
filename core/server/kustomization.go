package server

import (
	"context"
	"fmt"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cs *coreServer) ListKustomizations(ctx context.Context, msg *pb.ListKustomizationsRequest) (*pb.ListKustomizationsResponse, error) {
	clustersClient := clustersmngr.ClientFromCtx(ctx)

	clist := &clustersmngr.ClusteredKustomizationList{}

	if err := clustersClient.List(ctx, clist, client.InNamespace(msg.Namespace)); err != nil {
		return nil, err
	}

	var results []*pb.Kustomization

	for _, l := range clist.Lists() {
		for _, kustomization := range l.Items {
			k, err := types.KustomizationToProto(&kustomization)
			if err != nil {
				return nil, fmt.Errorf("converting items: %w", err)
			}

			results = append(results, k)
		}
	}

	return &pb.ListKustomizationsResponse{
		Kustomizations: results,
	}, nil
}

func (cs *coreServer) GetKustomization(ctx context.Context, msg *pb.GetKustomizationRequest) (*pb.GetKustomizationResponse, error) {
	k8s, err := cs.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	k := &kustomizev1.Kustomization{}

	if err := get(ctx, k8s, msg.Name, msg.Namespace, k); err != nil {
		return nil, err
	}

	res, err := types.KustomizationToProto(k)
	if err != nil {
		return nil, fmt.Errorf("converting kustomization to proto: %w", err)
	}

	return &pb.GetKustomizationResponse{Kustomization: res}, nil
}
