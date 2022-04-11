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
	res, err := cs.listObjects(ctx, msg.Namespace, listGitRepositoriesInNamespace)
	if err != nil {
		return nil, err
	}

	var results []*pb.GitRepository

	for _, object := range res {
		obj, ok := object.(*pb.GitRepository)
		if !ok {
			return nil, nil
		}

		results = append(results, obj)
	}

	return &pb.ListGitRepositoriesResponse{
		GitRepositories: results,
	}, nil
}

func (cs *coreServer) ListHelmRepositories(ctx context.Context, msg *pb.ListHelmRepositoriesRequest) (*pb.ListHelmRepositoriesResponse, error) {
	res, err := cs.listObjects(ctx, msg.Namespace, listHelmRepositoriesInNamespace)
	if err != nil {
		return nil, err
	}

	var results []*pb.HelmRepository

	for _, object := range res {
		obj, ok := object.(*pb.HelmRepository)
		if !ok {
			return nil, nil
		}

		results = append(results, obj)
	}

	return &pb.ListHelmRepositoriesResponse{
		HelmRepositories: results,
	}, nil
}

func (cs *coreServer) ListHelmCharts(ctx context.Context, msg *pb.ListHelmChartsRequest) (*pb.ListHelmChartsResponse, error) {
	res, err := cs.listObjects(ctx, msg.Namespace, listHelmChartsInNamespace)
	if err != nil {
		return nil, err
	}

	var results []*pb.HelmChart

	for _, object := range res {
		obj, ok := object.(*pb.HelmChart)
		if !ok {
			return nil, nil
		}

		results = append(results, obj)
	}

	return &pb.ListHelmChartsResponse{
		HelmCharts: results,
	}, nil
}

func (cs *coreServer) ListBuckets(ctx context.Context, msg *pb.ListBucketRequest) (*pb.ListBucketsResponse, error) {
	res, err := cs.listObjects(ctx, msg.Namespace, listBucketsInNamespace)
	if err != nil {
		return nil, err
	}

	var results []*pb.Bucket

	for _, object := range res {
		obj, ok := object.(*pb.Bucket)
		if !ok {
			return nil, nil
		}

		results = append(results, obj)
	}

	return &pb.ListBucketsResponse{
		Buckets: results,
	}, nil
}

func listGitRepositoriesInNamespace(
	ctx context.Context,
	clustersClient clustersmngr.Client,
	namespace string,
) ([]interface{}, error) {
	results := []interface{}{}
	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.GitRepositoryList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist, client.InNamespace(namespace)); err != nil {
		return results, err
	}

	for n, l := range clist.Lists() {
		list, ok := l.(*sourcev1.GitRepositoryList)
		if !ok {
			continue
		}

		for _, repository := range list.Items {
			results = append(results, types.GitRepositoryToProto(&repository, n))
		}
	}

	return results, nil
}

func listHelmRepositoriesInNamespace(
	ctx context.Context,
	clustersClient clustersmngr.Client,
	namespace string,
) ([]interface{}, error) {
	results := []interface{}{}
	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.HelmRepositoryList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist, client.InNamespace(namespace)); err != nil {
		return results, err
	}

	for n, l := range clist.Lists() {
		list, ok := l.(*sourcev1.HelmRepositoryList)
		if !ok {
			continue
		}

		for _, repository := range list.Items {
			results = append(results, types.HelmRepositoryToProto(&repository, n))
		}
	}

	return results, nil
}

func listHelmChartsInNamespace(
	ctx context.Context,
	clustersClient clustersmngr.Client,
	namespace string,
) ([]interface{}, error) {
	results := []interface{}{}
	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.HelmChartList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist, client.InNamespace(namespace)); err != nil {
		return results, err
	}

	for n, l := range clist.Lists() {
		list, ok := l.(*sourcev1.HelmChartList)
		if !ok {
			continue
		}

		for _, repository := range list.Items {
			results = append(results, types.HelmChartToProto(&repository, n))
		}
	}

	return results, nil
}

func listBucketsInNamespace(
	ctx context.Context,
	clustersClient clustersmngr.Client,
	namespace string,
) ([]interface{}, error) {
	results := []interface{}{}
	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.BucketList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist, client.InNamespace(namespace)); err != nil {
		return results, err
	}

	for n, l := range clist.Lists() {
		list, ok := l.(*sourcev1.BucketList)
		if !ok {
			continue
		}

		for _, repository := range list.Items {
			results = append(results, types.BucketToProto(&repository, n))
		}
	}

	return results, nil
}
