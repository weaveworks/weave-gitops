package server

import (
	"context"
	"errors"
	"fmt"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
)

const (
	FluxNamespacePartOf   = "flux"
	fluxNamespaceInstance = "flux-system"
)

var ErrNamespaceNotFound = errors.New("namespace not found")

func (cs *coreServer) GetFluxNamespace(ctx context.Context, msg *pb.GetFluxNamespaceRequest) (*pb.GetFluxNamespaceResponse, error) {
	for _, ns := range cs.cacheContainer.Namespaces() {
		instanceLabelMatch := ns.Labels[types.InstanceLabel] == fluxNamespaceInstance
		partofLabelMatch := ns.Labels[types.PartOfLabel] == FluxNamespacePartOf

		if instanceLabelMatch && partofLabelMatch {
			return &pb.GetFluxNamespaceResponse{Name: ns.Name}, nil
		}
	}

	return nil, ErrNamespaceNotFound
}

func (cs *coreServer) ListNamespaces(ctx context.Context, msg *pb.ListNamespacesRequest) (*pb.ListNamespacesResponse, error) {
	client := clustersmngr.ClientFromCtx(ctx)

	if client == nil {
		return nil, errors.New("getting clients pool from context: pool was nil")
	}

	namespaces := cs.cacheContainer.Namespaces()

	restCfg, err := client.RestConfig(clustersmngr.DefaultCluster)
	if err != nil {
		return nil, err
	}

	filtered, err := cs.nsChecker.FilterAccessibleNamespaces(ctx, restCfg, namespaces)
	if err != nil {
		return nil, fmt.Errorf("filtering namespaces: %w", err)
	}

	response := &pb.ListNamespacesResponse{
		Namespaces: []*pb.Namespace{},
	}

	for _, ns := range filtered {
		response.Namespaces = append(response.Namespaces, types.NamespaceToProto(ns))
	}

	return response, nil
}
