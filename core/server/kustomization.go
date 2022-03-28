package server

import (
	"context"
	"errors"
	"fmt"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
)

func (cs *coreServer) ListKustomizations(ctx context.Context, msg *pb.ListKustomizationsRequest) (*pb.ListKustomizationsResponse, error) {
	clientsPool := clustersmngr.ClientsPoolFromCtx(ctx)
	if clientsPool == nil {
		return &pb.ListKustomizationsResponse{
			Kustomizations: []*pb.Kustomization{},
		}, errors.New("no clients pool present in context")
	}

	var results []*pb.Kustomization

	//TODO: handle failures and parallelize
	for _, c := range clientsPool.Clients() {
		l := &kustomizev1.KustomizationList{}
		if err := list(ctx, c, temporarilyEmptyAppName, msg.Namespace, l); err != nil {
			return nil, err
		}

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
