package server

import (
	"context"
	"errors"

	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	corev1 "k8s.io/api/core/v1"
)

const (
	fluxNamespacePartOf   = "flux"
	fluxNamespaceInstance = "flux-system"
)

var ErrNamespaceNotFound = errors.New("namespace not found")

func (as *coreServer) GetFluxNamespace(ctx context.Context, msg *pb.GetFluxNamespaceRequest) (*pb.GetFluxNamespaceResponse, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	nsList := corev1.NamespaceList{}
	options := matchLabel(
		withPartOfLabel(fluxNamespacePartOf),
		withInstanceLabel(fluxNamespaceInstance),
	)

	if err = k8s.List(ctx, &nsList, &options); err != nil {
		return nil, doClientError(err)
	}

	if len(nsList.Items) == 0 {
		return nil, ErrNamespaceNotFound
	}

	return &pb.GetFluxNamespaceResponse{Name: nsList.Items[0].Name}, nil
}

func (as *coreServer) ListNamespaces(ctx context.Context, msg *pb.ListNamespacesRequest) (*pb.ListNamespacesResponse, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	nsList := corev1.NamespaceList{}

	if err = k8s.List(ctx, &nsList); err != nil {
		return nil, doClientError(err)
	}

	response := &pb.ListNamespacesResponse{
		Namespaces: []*pb.Namespace{},
	}

	for _, ns := range nsList.Items {
		response.Namespaces = append(response.Namespaces, types.NamespaceToProto(ns))
	}

	return response, nil
}
