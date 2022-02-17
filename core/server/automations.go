package server

import (
	"context"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
)

func (as *coreServer) ListKustomizations(ctx context.Context, msg *pb.ListKustomizationsRequest) (*pb.ListKustomizationsResponse, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	l := &kustomizev1.KustomizationList{}

	if err := list(ctx, k8s, temporarilyEmptyAppName, msg.Namespace, l); err != nil {
		return nil, err
	}

	var results []*pb.Kustomization
	for _, kustomization := range l.Items {
		results = append(results, types.KustomizationToProto(&kustomization))
	}

	return &pb.ListKustomizationsResponse{
		Kustomizations: results,
	}, nil
}

func (as *coreServer) ListHelmReleases(ctx context.Context, msg *pb.ListHelmReleasesRequest) (*pb.ListHelmReleasesResponse, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	l := &helmv2.HelmReleaseList{}

	if err := list(ctx, k8s, temporarilyEmptyAppName, msg.Namespace, l); err != nil {
		return nil, err
	}

	var results []*pb.HelmRelease
	for _, repository := range l.Items {
		results = append(results, types.HelmReleaseToProto(&repository))
	}

	return &pb.ListHelmReleasesResponse{
		HelmReleases: results,
	}, nil
}
