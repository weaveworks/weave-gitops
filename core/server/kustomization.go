package server

import (
	"context"
	"errors"
	"fmt"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/hashicorp/go-multierror"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cs *coreServer) ListKustomizations(ctx context.Context, msg *pb.ListKustomizationsRequest) (*pb.ListKustomizationsResponse, error) {
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
		return &kustomizev1.KustomizationList{}
	})

	opts := []client.ListOption{}
	if msg.Pagination != nil {
		opts = append(opts, client.Limit(msg.Pagination.PageSize))
		opts = append(opts, client.Continue(msg.Pagination.PageToken))
	}

	if err := clustersClient.ClusteredList(ctx, clist, true, opts...); err != nil {
		var errs clustersmngr.ClusteredListError
		if !errors.As(err, &errs) {
			return nil, err
		}

		for _, e := range errs.Errors {
			respErrors = append(respErrors, &pb.ListError{ClusterName: e.Cluster, Namespace: e.Namespace, Message: e.Err.Error()})
		}
	}

	var results []*pb.Kustomization

	for n, lists := range clist.Lists() {
		for _, l := range lists {
			list, ok := l.(*kustomizev1.KustomizationList)
			if !ok {
				continue
			}

			for _, kustomization := range list.Items {
				k, err := types.KustomizationToProto(&kustomization, n)
				if err != nil {
					return nil, fmt.Errorf("converting items: %w", err)
				}

				results = append(results, k)
			}
		}
	}

	return &pb.ListKustomizationsResponse{
		Kustomizations: results,
		NextPageToken:  clist.GetContinue(),
		Errors:         respErrors,
	}, nil
}

func (cs *coreServer) GetKustomization(ctx context.Context, msg *pb.GetKustomizationRequest) (*pb.GetKustomizationResponse, error) {
	clustersClient, err := cs.clientsFactory.GetImpersonatedClientForCluster(ctx, auth.Principal(ctx), msg.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	apiVersion := kustomizev1.GroupVersion.String()
	k := &kustomizev1.Kustomization{}
	key := client.ObjectKey{
		Name:      msg.Name,
		Namespace: msg.Namespace,
	}

	if err := clustersClient.Get(ctx, msg.ClusterName, key, k); err != nil {
		return nil, err
	}

	res, err := types.KustomizationToProto(k, msg.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("converting kustomization to proto: %w", err)
	}

	res.ApiVersion = apiVersion

	return &pb.GetKustomizationResponse{Kustomization: res}, nil
}
