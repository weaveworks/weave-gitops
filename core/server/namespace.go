package server

import (
	"context"
	"errors"

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
	response := &pb.ListNamespacesResponse{
		Namespaces: []*pb.Namespace{},
	}

	for _, ns := range cs.cacheContainer.Namespaces() {
		response.Namespaces = append(response.Namespaces, types.NamespaceToProto(ns))
	}

	return response, nil
}
