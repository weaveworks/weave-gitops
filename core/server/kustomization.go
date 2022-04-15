package server

import (
	"context"
	"fmt"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
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

	//for nsID, ns := range nsList {
	//	fmt.Println("nsID", nsID, "NS", ns.Name)
	//}
	//fmt.Println("PaginationObject", msg.Pagination)
	//if msg.Pagination != nil {
	//	fmt.Println("Request PageSize", msg.Pagination.PageSize)
	//	fmt.Println("Request PageToken", msg.Pagination.PageToken)
	//}

	newPageToken := ""

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
		var namespaceStartIndex = 0
		k8sPageToken := ""

		itemsLeft := msg.Pagination.PageSize

		if msg.Pagination.PageToken != "" {
			var pageTokenInfo types.PageTokenInfo
			err := decodeFromBase64(&pageTokenInfo, msg.Pagination.PageToken)
			if err != nil {
				return nil, fmt.Errorf("error decoding next token %w", err)
			}
			//fmt.Printf("NextToken received encoded => %+v\n", msg.Pagination.PageToken)
			//fmt.Printf("NextToken received decoded => %+v\n", pageTokenInfo)
			namespaceStartIndex = pageTokenInfo.NamespaceIndex
			k8sPageToken = pageTokenInfo.K8sPageToken
		}

		for cNsIndex, ns := range nsList[namespaceStartIndex:] {
			//fmt.Println("Querying NS", ns.Name, "NSindex", cNsIndex, "Nk8spageTOKEN", k8sPageToken)
			nsResult, nextPageToken, err := listKustomizationsInNamespace(ctx, clustersClient, ns.Name, itemsLeft, k8sPageToken)
			if err != nil {
				cs.logger.Error(err, fmt.Sprintf("unable to list kustomizations in namespace: %s", ns.Name))

				continue
			}
			//fmt.Println("Next k8s Page Token", nextPageToken, "len", len(nsResult))

			//for _, ns := range nsResult {
			//	fmt.Println("\tNS", ns.Name)
			//}

			results = append(results, nsResult...)

			itemsLeft = itemsLeft - int32(len(nsResult))

			if nextPageToken != "" {
				//fmt.Println("CASE", "if nextPageToken != \"\" {")
				//data, err := base64.RawStdEncoding.DecodeString(nextPageToken)
				//if err != nil {
				//	return nil, err
				//}
				//fmt.Println("K8S next token encoded", nextPageToken)
				//fmt.Println("K8S next token decoded", string(data))
				newPageToken, err = getPageTokenInfoBase64(cNsIndex+namespaceStartIndex, nextPageToken, ns.Name)
				if err != nil {
					return nil, err
				}

				break
			}
			if itemsLeft == 0 {
				//fmt.Println("if itemsLeft == 0 {")
				newPageToken, err = getPageTokenInfoBase64(cNsIndex+namespaceStartIndex+1, "", nsList[cNsIndex+namespaceStartIndex+1].Name)
				if err != nil {
					return nil, err
				}

				break
			}
		}

	}

	return &pb.ListKustomizationsResponse{
		Kustomizations: results,
		NextPageToken:  newPageToken,
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
		fmt.Println("list.GetContinue()", list.GetContinue())
	}

	return results, nextPageToken, nil
}
