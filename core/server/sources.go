package server

import (
	"context"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/hashicorp/go-multierror"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cs *coreServer) ListGitRepositories(ctx context.Context, msg *pb.ListGitRepositoriesRequest) (*pb.ListGitRepositoriesResponse, error) {
	respErrors := []*pb.ListError{}

	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			for _, err := range merr.Errors {
				if cerr, ok := err.(*clustersmngr.ClientError); ok {
					respErrors = append(respErrors, &pb.ListError{ClusterName: cerr.ClusterName, Message: cerr.Error()})
				}
			}
		}
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.GitRepositoryList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist, true); err != nil {
		return nil, err
	}

	var results []*pb.GitRepository

	clusterUserNamespaces := cs.clientsFactory.GetUserNamespaces(auth.Principal(ctx))

	for n, lists := range clist.Lists() {
		for _, l := range lists {
			list, ok := l.(*sourcev1.GitRepositoryList)
			if !ok {
				continue
			}

			for _, repository := range list.Items {
				tenant := GetTenant(repository.Namespace, n, clusterUserNamespaces)

				results = append(results, types.GitRepositoryToProto(&repository, n, tenant))
			}
		}
	}

	return &pb.ListGitRepositoriesResponse{
		GitRepositories: results,
		Errors:          respErrors,
	}, nil
}

func (cs *coreServer) ListHelmRepositories(ctx context.Context, msg *pb.ListHelmRepositoriesRequest) (*pb.ListHelmRepositoriesResponse, error) {
	respErrors := []*pb.ListError{}

	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			for _, err := range merr.Errors {
				if cerr, ok := err.(*clustersmngr.ClientError); ok {
					respErrors = append(respErrors, &pb.ListError{ClusterName: cerr.ClusterName, Message: cerr.Error()})
				}
			}
		}
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.HelmRepositoryList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist, true); err != nil {
		return nil, err
	}

	var results []*pb.HelmRepository

	clusterUserNamespaces := cs.clientsFactory.GetUserNamespaces(auth.Principal(ctx))

	for n, lists := range clist.Lists() {
		for _, l := range lists {
			list, ok := l.(*sourcev1.HelmRepositoryList)
			if !ok {
				continue
			}

			for _, repository := range list.Items {
				tenant := GetTenant(repository.Namespace, n, clusterUserNamespaces)

				results = append(results, types.HelmRepositoryToProto(&repository, n, tenant))
			}
		}
	}

	return &pb.ListHelmRepositoriesResponse{
		HelmRepositories: results,
		Errors:           respErrors,
	}, nil
}

func (cs *coreServer) ListHelmCharts(ctx context.Context, msg *pb.ListHelmChartsRequest) (*pb.ListHelmChartsResponse, error) {
	respErrors := []*pb.ListError{}

	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			for _, err := range merr.Errors {
				if cerr, ok := err.(*clustersmngr.ClientError); ok {
					respErrors = append(respErrors, &pb.ListError{ClusterName: cerr.ClusterName, Message: cerr.Error()})
				}
			}
		}
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.HelmChartList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist, true); err != nil {
		return nil, err
	}

	var results []*pb.HelmChart

	clusterUserNamespaces := cs.clientsFactory.GetUserNamespaces(auth.Principal(ctx))

	for n, lists := range clist.Lists() {
		for _, l := range lists {
			list, ok := l.(*sourcev1.HelmChartList)
			if !ok {
				continue
			}

			for _, repository := range list.Items {
				tenant := GetTenant(repository.Namespace, n, clusterUserNamespaces)

				results = append(results, types.HelmChartToProto(&repository, n, tenant))
			}
		}
	}

	return &pb.ListHelmChartsResponse{
		HelmCharts: results,
		Errors:     respErrors,
	}, nil
}

func (cs *coreServer) ListBuckets(ctx context.Context, msg *pb.ListBucketRequest) (*pb.ListBucketsResponse, error) {
	respErrors := []*pb.ListError{}

	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			for _, err := range merr.Errors {
				if cerr, ok := err.(*clustersmngr.ClientError); ok {
					respErrors = append(respErrors, &pb.ListError{ClusterName: cerr.ClusterName, Message: cerr.Error()})
				}
			}
		}
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.BucketList{}
	})

	var results []*pb.Bucket

	clusterUserNamespaces := cs.clientsFactory.GetUserNamespaces(auth.Principal(ctx))

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
				tenant := GetTenant(repository.Namespace, n, clusterUserNamespaces)

				results = append(results, types.BucketToProto(&repository, n, tenant))
			}
		}
	}

	return &pb.ListBucketsResponse{
		Buckets: results,
		Errors:  respErrors,
	}, nil
}

func (cs *coreServer) ListOCIRepositories(ctx context.Context, msg *pb.ListOCIRepositoriesRequest) (*pb.ListOCIRepositoriesResponse, error) {
	respErrors := []*pb.ListError{}

	if featureflags.Get("WEAVE_GITOPS_FEATURE_OCI_REPOSITORIES") == "" {
		return &pb.ListOCIRepositoriesResponse{
			OciRepositories: []*pb.OCIRepository{},
			Errors:          respErrors,
		}, nil
	}

	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		if merr, ok := err.(*multierror.Error); ok {
			for _, err := range merr.Errors {
				if cerr, ok := err.(*clustersmngr.ClientError); ok {
					respErrors = append(respErrors, &pb.ListError{ClusterName: cerr.ClusterName, Message: cerr.Error()})
				}
			}
		}
	}

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &sourcev1.OCIRepositoryList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist, true); err != nil {
		return nil, err
	}

	var results []*pb.OCIRepository

	clusterUserNamespaces := cs.clientsFactory.GetUserNamespaces(auth.Principal(ctx))

	for n, lists := range clist.Lists() {
		for _, l := range lists {
			list, ok := l.(*sourcev1.OCIRepositoryList)
			if !ok {
				continue
			}

			for _, repository := range list.Items {
				tenant := GetTenant(repository.Namespace, n, clusterUserNamespaces)

				results = append(results, types.OCIRepositoryToProto(&repository, n, tenant))
			}
		}
	}

	return &pb.ListOCIRepositoriesResponse{
		OciRepositories: results,
		Errors:          respErrors,
	}, nil
}
