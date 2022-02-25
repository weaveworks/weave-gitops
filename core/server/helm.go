package server

import (
	"context"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
)

func (cs *coreServer) ListHelmReleases(ctx context.Context, msg *pb.ListHelmReleasesRequest) (*pb.ListHelmReleasesResponse, error) {
	k8s, err := cs.k8s.Client(ctx)
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

func (cs *coreServer) GetHelmRelease(ctx context.Context, msg *pb.GetHelmReleaseRequest) (*pb.GetHelmReleaseResponse, error) {
	k8s, err := cs.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	helmRelease := helmv2.HelmRelease{}

	if err = get(ctx, k8s, msg.Name, msg.Namespace, &helmRelease); err != nil {
		return nil, err
	}

	return &pb.GetHelmReleaseResponse{
		HelmRelease: types.HelmReleaseToProto(&helmRelease),
	}, nil
}
