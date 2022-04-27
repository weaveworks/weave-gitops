package server

import (
	"context"
	"errors"

	"github.com/weaveworks/weave-gitops/gitops-server/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/gitops-server/core/server/types"
	pb "github.com/weaveworks/weave-gitops/gitops-server/pkg/api/core"
)

const (
	FluxNamespacePartOf   = "flux"
	fluxNamespaceInstance = "flux-system"
)

var ErrNamespaceNotFound = errors.New("namespace not found")

func (cs *coreServer) GetFluxNamespace(ctx context.Context, msg *pb.GetFluxNamespaceRequest) (*pb.GetFluxNamespaceResponse, error) {
	namespaces, err := cs.namespaces()
	if err != nil {
		return nil, err
	}

	nsList, found := namespaces[clustersmngr.DefaultCluster]
	if !found {
		return nil, defaultClusterNotFound{}
	}

	for _, ns := range nsList {
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

	restCfg, err := client.RestConfig(clustersmngr.DefaultCluster)
	if err != nil {
		return nil, err
	}

	response := &pb.ListNamespacesResponse{
		Namespaces: []*pb.Namespace{},
	}

	namespaces, err := cs.namespaces()
	if err != nil {
		return nil, err
	}

	for clusterName, nsList := range namespaces {
		nsList, err := cs.nsChecker.FilterAccessibleNamespaces(ctx, restCfg, nsList)
		if err != nil {
			cs.logger.Error(err, "filtering namespaces")

			continue
		}

		for _, ns := range nsList {
			response.Namespaces = append(
				response.Namespaces,
				types.NamespaceToProto(ns, clusterName),
			)
		}
	}

	return response, nil
}
