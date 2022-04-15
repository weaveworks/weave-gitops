package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	v1 "k8s.io/api/core/v1"
)

type pageTokenInfo struct {
	NamespaceIndex int    `json:"namespace_index"`
	Namespace      string `json:"namespace"`
	K8sPageToken   string `json:"k8s_page_token"`
}

type getRawPageData func(namespace string, limit int32, pageToken string) (string, int32, error)

type pagination struct {
	namespaceList []v1.Namespace
	getRawPage    getRawPageData
	pageParams    *pb.Pagination
}

func NewPagination(namespaceList []v1.Namespace, pageParams *pb.Pagination, getRawPage getRawPageData) *pagination {
	return &pagination{
		namespaceList: namespaceList,
		pageParams:    pageParams,
		getRawPage:    getRawPage,
	}
}

func (p *pagination) GetNextPage() (string, error) {
	newPageToken := ""
	currentNamespaceIndex := 0
	currentK8sPageToken := ""
	itemsLeft := p.pageParams.PageSize

	if p.pageParams.PageToken != "" {
		var pageTokenInfo pageTokenInfo

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

func getPageTokenInfoBase64(namespaceIndex int, k8sNextToken string, namespace string) (string, error) {
	nextTokenInfo := pageTokenInfo{
		NamespaceIndex: namespaceIndex,
		K8sPageToken:   k8sNextToken,
		Namespace:      namespace,
	}

	nextToken, err := encodeToBase64(nextTokenInfo)
	if err != nil {
		return "", err
	}

	return nextToken, nil
}

func decodeFromBase64(v interface{}, enc string) error {
	return json.NewDecoder(base64.NewDecoder(base64.StdEncoding, strings.NewReader(enc))).Decode(v)
}

func encodeToBase64(v interface{}) (string, error) {
	var buf bytes.Buffer
	encoder := base64.NewEncoder(base64.StdEncoding, &buf)

	err := json.NewEncoder(encoder).Encode(v)
	if err != nil {
		return "", err
	}

	encoder.Close()

	return buf.String(), nil
}
