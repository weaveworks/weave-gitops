package server

import (
	"fmt"

	v1 "k8s.io/api/core/v1"
)

type PageTokenInfo struct {
	NamespaceIndex int    `json:"namespace_index"`
	Namespace      string `json:"namespace"`
	K8sPageToken   string `json:"k8s_page_token"`
}

type getRawPageData func(namespace string, limit int32, pageToken string) (string, int32, error)

func GetNextPage(namespaceList []v1.Namespace, pageSize int32, pageToken string, getRawPage getRawPageData) (string, error) {
	newPageToken := ""
	currentNamespaceIndex := 0
	currentK8sPageToken := ""
	itemsLeft := pageSize

	if pageToken != "" {
		var pageTokenInfo PageTokenInfo

		err := decodeFromBase64(&pageTokenInfo, pageToken)
		if err != nil {
			return "", fmt.Errorf("error decoding next token %w", err)
		}

		currentNamespaceIndex = pageTokenInfo.NamespaceIndex
		currentK8sPageToken = pageTokenInfo.K8sPageToken
	}

	for nsIndex, ns := range namespaceList[currentNamespaceIndex:] {
		nextPageToken, pageLen, err := getRawPage(ns.Name, itemsLeft, currentK8sPageToken)
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

		if itemsLeft == 0 && globalNamespaceIndex+1 < len(namespaceList) {
			newPageToken, err = getPageTokenInfoBase64(globalNamespaceIndex+1, "", namespaceList[globalNamespaceIndex+1].Name)
			if err != nil {
				return "", err
			}

			break
		}
	}

	return newPageToken, nil
}

func getPageTokenInfoBase64(namespaceIndex int, k8sNextToken string, namespace string) (string, error) {
	nextTokenInfo := PageTokenInfo{
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
