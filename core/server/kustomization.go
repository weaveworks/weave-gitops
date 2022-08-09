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

	clustersClient, err := cs.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
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

	clusterUserNamespaces := cs.clustersManager.GetUserNamespaces(auth.Principal(ctx))

	for n, lists := range clist.Lists() {
		for _, l := range lists {
			list, ok := l.(*kustomizev1.KustomizationList)
			if !ok {
				continue
			}

			for _, kustomization := range list.Items {
				tenant := GetTenant(kustomization.Namespace, n, clusterUserNamespaces)

				k, err := types.KustomizationToProto(&kustomization, n, tenant)
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
