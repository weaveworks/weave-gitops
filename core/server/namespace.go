package server

import (
	"context"
	"errors"

	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
)

const (
	fluxNamespacePartOf   = "flux"
	fluxNamespaceInstance = "flux-system"
)

var ErrNamespaceNotFound = errors.New("namespace not found")

func (as *coreServer) GetFluxNamespace(ctx context.Context, msg *pb.GetFluxNamespaceRequest) (*pb.GetFluxNamespaceResponse, error) {
	for _, ns := range as.cacheContainer.Namespaces() {
		instanceLabelMatch := ns.Labels[types.InstanceLabel] == fluxNamespaceInstance
		partofLabelMatch := ns.Labels[types.PartOfLabel] == fluxNamespacePartOf

		if instanceLabelMatch && partofLabelMatch {
			return &pb.GetFluxNamespaceResponse{Name: ns.Name}, nil
		}
	}

	return nil, ErrNamespaceNotFound

}

func (as *coreServer) ListNamespaces(ctx context.Context, msg *pb.ListNamespacesRequest) (*pb.ListNamespacesResponse, error) {
	response := &pb.ListNamespacesResponse{
		Namespaces: []*pb.Namespace{},
	}

	for _, ns := range as.cacheContainer.Namespaces() {
		response.Namespaces = append(response.Namespaces, types.NamespaceToProto(ns))
	}

	return response, nil
}
