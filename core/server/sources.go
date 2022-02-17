package server

import (
	"context"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
)

func (as *coreServer) ListGitRepositories(ctx context.Context, msg *pb.ListGitRepositoriesRequest) (*pb.ListGitRepositoriesResponse, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	l := &sourcev1.GitRepositoryList{}

	if err := list(ctx, k8s, temporarilyEmptyAppName, msg.Namespace, l); err != nil {
		return nil, err
	}

	var results []*pb.GitRepository
	for _, repository := range l.Items {
		results = append(results, types.GitRepositoryToProto(&repository))
	}

	return &pb.ListGitRepositoriesResponse{
		GitRepositories: results,
	}, nil
}

func (as *coreServer) ListHelmRepositories(ctx context.Context, msg *pb.ListHelmRepositoriesRequest) (*pb.ListHelmRepositoriesResponse, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	l := &sourcev1.HelmRepositoryList{}

	if err := list(ctx, k8s, temporarilyEmptyAppName, msg.Namespace, l); err != nil {
		return nil, err
	}

	var results []*pb.HelmRepository
	for _, repository := range l.Items {
		results = append(results, types.HelmRepositoryToProto(&repository))
	}

	return &pb.ListHelmRepositoriesResponse{
		HelmRepositories: results,
	}, nil
}

func (as *coreServer) ListHelmCharts(ctx context.Context, msg *pb.ListHelmChartsRequest) (*pb.ListHelmChartsResponse, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	l := &sourcev1.HelmChartList{}

	if err := list(ctx, k8s, temporarilyEmptyAppName, msg.Namespace, l); err != nil {
		return nil, err
	}

	var results []*pb.HelmChart
	for _, repository := range l.Items {
		results = append(results, types.HelmChartToProto(&repository))
	}

	return &pb.ListHelmChartsResponse{
		HelmCharts: results,
	}, nil
}

func (as *coreServer) ListBuckets(ctx context.Context, msg *pb.ListBucketRequest) (*pb.ListBucketsResponse, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	l := &sourcev1.BucketList{}

	if err := list(ctx, k8s, temporarilyEmptyAppName, msg.Namespace, l); err != nil {
		return nil, err
	}

	var results []*pb.Bucket
	for _, repository := range l.Items {
		results = append(results, types.BucketToProto(&repository))
	}

	return &pb.ListBucketsResponse{
		Buckets: results,
	}, nil
}
