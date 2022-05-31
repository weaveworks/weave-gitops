package server

import (
	"context"
	"fmt"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cs *coreServer) ListGitRepositories(ctx context.Context, msg *pb.ListGitRepositoriesRequest) (*pb.ListGitRepositoriesResponse, error) {
	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %s", err)
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.GitRepositoryList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist, true); err != nil {
		return nil, err
	}

	var results []*pb.GitRepository

	for n, lists := range clist.Lists() {
		for _, l := range lists {
			list, ok := l.(*sourcev1.GitRepositoryList)
			if !ok {
				continue
			}

			for _, repository := range list.Items {
				results = append(results, types.GitRepositoryToProto(&repository, n))
			}
		}
	}

	return &pb.ListGitRepositoriesResponse{
		GitRepositories: results,
	}, nil
}

func (cs *coreServer) ListHelmRepositories(ctx context.Context, msg *pb.ListHelmRepositoriesRequest) (*pb.ListHelmRepositoriesResponse, error) {
	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %s", err)
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.HelmRepositoryList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist, true); err != nil {
		return nil, err
	}

	var results []*pb.HelmRepository

	for n, lists := range clist.Lists() {
		for _, l := range lists {
			list, ok := l.(*sourcev1.HelmRepositoryList)
			if !ok {
				continue
			}

			for _, repository := range list.Items {
				results = append(results, types.HelmRepositoryToProto(&repository, n))
			}
		}
	}

	return &pb.ListHelmRepositoriesResponse{
		HelmRepositories: results,
	}, nil
}

func (cs *coreServer) ListHelmCharts(ctx context.Context, msg *pb.ListHelmChartsRequest) (*pb.ListHelmChartsResponse, error) {
	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %s", err)
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.HelmChartList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist, true); err != nil {
		return nil, err
	}

	var results []*pb.HelmChart

	for n, lists := range clist.Lists() {
		for _, l := range lists {
			list, ok := l.(*sourcev1.HelmChartList)
			if !ok {
				continue
			}

			for _, repository := range list.Items {
				results = append(results, types.HelmChartToProto(&repository, n))
			}
		}
	}

	return &pb.ListHelmChartsResponse{
		HelmCharts: results,
	}, nil
}

func (cs *coreServer) ListBuckets(ctx context.Context, msg *pb.ListBucketRequest) (*pb.ListBucketsResponse, error) {
	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %s", err)
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.BucketList{}
	})

	var results []*pb.Bucket

	if err := clustersClient.ClusteredList(ctx, clist, true); err != nil {
		return nil, err
	}

	for n, lists := range clist.Lists() {
		for _, l := range lists {
			list, ok := l.(*sourcev1.BucketList)
			if !ok {
				continue
			}

			for _, repository := range list.Items {
				results = append(results, types.BucketToProto(&repository, n))
			}
		}
	}

	return &pb.ListBucketsResponse{
		Buckets: results,
	}, nil
}
