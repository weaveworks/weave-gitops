package server

import (
	"fmt"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	v1 "k8s.io/api/core/v1"
)

type GetRawPageData func(namespace string, limit int32, pageToken string) (string, int32, error)

type Pagination struct {
	namespaceList []v1.Namespace
	getRawPage    GetRawPageData
	pageParams    *pb.Pagination
}

func NewPagination(namespaceList []v1.Namespace, pageParams *pb.Pagination, getRawPage GetRawPageData) *Pagination {
	return &Pagination{
		namespaceList: namespaceList,
		pageParams:    pageParams,
		getRawPage:    getRawPage,
	}
}

func (p *Pagination) GetNextPage() (string, error) {

	newPageToken := ""
	currentNamespaceIndex := 0
	currentK8sPageToken := ""
	itemsLeft := p.pageParams.PageSize

	if p.pageParams.PageToken != "" {
		var pageTokenInfo types.PageTokenInfo
		err := decodeFromBase64(&pageTokenInfo, p.pageParams.PageToken)
		if err != nil {
			return "", fmt.Errorf("error decoding next token %w", err)
		}
		currentNamespaceIndex = pageTokenInfo.NamespaceIndex
		currentK8sPageToken = pageTokenInfo.K8sPageToken
	}

	for nsIndex, ns := range p.namespaceList[currentNamespaceIndex:] {
		nextPageToken, pageLen, err := p.getRawPage(ns.Name, itemsLeft, currentK8sPageToken)
		if err != nil {
			continue
		}

		itemsLeft = itemsLeft - pageLen

		globalNamespaceIndex := nsIndex + currentNamespaceIndex

		if nextPageToken != "" {
			newPageToken, err = getPageTokenInfoBase64(globalNamespaceIndex, nextPageToken, ns.Name)
			if err != nil {
				return "", err
			}

			break
		}
		if itemsLeft == 0 {
			newPageToken, err = getPageTokenInfoBase64(globalNamespaceIndex+1, "", p.namespaceList[globalNamespaceIndex+1].Name)
			if err != nil {
				return "", err
			}

			break
		}
	}

	return newPageToken, nil
}
