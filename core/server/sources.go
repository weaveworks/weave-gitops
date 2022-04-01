package server

import (
	"context"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cs *coreServer) ListGitRepositories(ctx context.Context, msg *pb.ListGitRepositoriesRequest) (*pb.ListGitRepositoriesResponse, error) {
	clustersClient := clustersmngr.ClientFromCtx(ctx)

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.GitRepositoryList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist, client.InNamespace(msg.Namespace)); err != nil {
		return nil, err
	}

	var results []*pb.GitRepository

	for n, l := range clist.Lists() {
		list, ok := l.(*sourcev1.GitRepositoryList)
		if !ok {
			continue
		}

		for _, repository := range list.Items {
			results = append(results, types.GitRepositoryToProto(&repository, n))
		}
	}

	return &pb.ListGitRepositoriesResponse{
		GitRepositories: results,
	}, nil
}

func (cs *coreServer) ListHelmRepositories(ctx context.Context, msg *pb.ListHelmRepositoriesRequest) (*pb.ListHelmRepositoriesResponse, error) {
	clustersClient := clustersmngr.ClientFromCtx(ctx)

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.HelmRepositoryList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist, client.InNamespace(msg.Namespace)); err != nil {
		return nil, err
	}

	var results []*pb.HelmRepository

	for n, l := range clist.Lists() {
		list, ok := l.(*sourcev1.HelmRepositoryList)
		if !ok {
			continue
		}

		for _, repository := range list.Items {
			results = append(results, types.HelmRepositoryToProto(&repository, n))
		}
	}

	return &pb.ListHelmRepositoriesResponse{
		HelmRepositories: results,
	}, nil
}

func (cs *coreServer) ListHelmCharts(ctx context.Context, msg *pb.ListHelmChartsRequest) (*pb.ListHelmChartsResponse, error) {
	clustersClient := clustersmngr.ClientFromCtx(ctx)

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.HelmChartList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist, client.InNamespace(msg.Namespace)); err != nil {
		return nil, err
	}

	var results []*pb.HelmChart

	for n, l := range clist.Lists() {
		list, ok := l.(*sourcev1.HelmChartList)
		if !ok {
			continue
		}

		for _, repository := range list.Items {
			results = append(results, types.HelmChartToProto(&repository, n))
		}
	}

	return &pb.ListHelmChartsResponse{
		HelmCharts: results,
	}, nil
}

func (cs *coreServer) ListBuckets(ctx context.Context, msg *pb.ListBucketRequest) (*pb.ListBucketsResponse, error) {
	clustersClient := clustersmngr.ClientFromCtx(ctx)

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.BucketList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist, client.InNamespace(msg.Namespace)); err != nil {
		return nil, err
	}

	var results []*pb.Bucket

	for n, l := range clist.Lists() {
		list, ok := l.(*sourcev1.BucketList)
		if !ok {
			continue
		}

		for _, bucket := range list.Items {
			results = append(results, types.BucketToProto(&bucket, n))
		}
	}

	return &pb.ListBucketsResponse{
		Buckets: results,
	}, nil
}
