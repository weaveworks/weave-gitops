package server

import (
	"context"
	"fmt"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/gitops-server/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/gitops-server/core/server/types"
	pb "github.com/weaveworks/weave-gitops/gitops-server/pkg/api/core"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type defaultClusterNotFound struct{}

func (e defaultClusterNotFound) Error() string {
	return "default cluster not found"
}

func (cs *coreServer) ListKustomizations(ctx context.Context, msg *pb.ListKustomizationsRequest) (*pb.ListKustomizationsResponse, error) {
	clustersClient := clustersmngr.ClientFromCtx(ctx)

	// TODO: Implement pagination for cases when the filter namespace is not empty
	if msg.Namespace != "" && msg.Pagination == nil {
		res, _, err := listKustomizationsInNamespace(ctx, clustersClient, msg.Namespace, 0, "")

		return &pb.ListKustomizationsResponse{
			Kustomizations: res,
		}, err
	}

	var results []*pb.Kustomization

	namespaces, err := cs.namespaces()
	if err != nil {
		return nil, err
	}

	nsList, found := namespaces[clustersmngr.DefaultCluster]
	if !found {
		return nil, defaultClusterNotFound{}
	}

	newNextPageToken := ""

	// TODO: Once the UI handles pagination we can remove this if block.
	// It was left like that so the UI doesn't brake with backend pagination changes
	if msg.Pagination == nil {
		for _, ns := range nsList {
			nsResult, _, err := listKustomizationsInNamespace(ctx, clustersClient, ns.Name, 0, "")
			if err != nil {
				cs.logger.Error(err, fmt.Sprintf("unable to list kustomizations in namespace: %s", ns.Name))

				continue
			}

			results = append(results, nsResult...)
		}
	} else {

		newNextPageToken, err = GetNextPage(
			nsList,
			msg.Pagination.PageSize,
			msg.Pagination.PageToken,
			func(namespace string, limit int32, pageToken string) (string, int32, error) {
				nsResult, nextK8sPageToken, err := listKustomizationsInNamespace(ctx, clustersClient, namespace, limit, pageToken)
				if err != nil {
					cs.logger.Error(err, fmt.Sprintf("unable to list kustomizations in namespace: %s", namespace))
					return "", 0, err
				}

				results = append(results, nsResult...)

				return nextK8sPageToken, int32(len(nsResult)), nil
			})
		if err != nil {
			return nil, err
		}

	}

	return &pb.ListKustomizationsResponse{
		Kustomizations: results,
		NextPageToken:  newNextPageToken,
	}, nil
}

func (cs *coreServer) GetKustomization(ctx context.Context, msg *pb.GetKustomizationRequest) (*pb.GetKustomizationResponse, error) {
	clustersClient := clustersmngr.ClientFromCtx(ctx)

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

	return &pb.GetKustomizationResponse{Kustomization: res}, nil
}

func listKustomizationsInNamespace(
	ctx context.Context,
	clustersClient clustersmngr.Client,
	namespace string,
	limit int32,
	pageToken string,
) ([]*pb.Kustomization, string, error) {
	var results []*pb.Kustomization

	clist := clustersmngr.NewClusteredList(func() client.ObjectList {
		return &kustomizev1.KustomizationList{}
	})

	if err := clustersClient.ClusteredList(ctx, clist,
		client.InNamespace(namespace),
		client.Limit(limit),
		client.Continue(pageToken),
	); err != nil {
		return results, "", err
	}

	nextPageToken := ""

	for n, l := range clist.Lists() {
		list, ok := l.(*kustomizev1.KustomizationList)
		if !ok {
			continue
		}

		for _, kustomization := range list.Items {
			k, err := types.KustomizationToProto(&kustomization, n)
			if err != nil {
				return results, "", fmt.Errorf("converting items: %w", err)
			}

			results = append(results, k)
		}

		nextPageToken = list.GetContinue()
	}

	return results, nextPageToken, nil
}
