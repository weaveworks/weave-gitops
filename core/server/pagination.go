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

	// If there is information in pageToken we need to decode it to know the exact
	// place we are going to start querying from. We will get the namespace and the
	// k8s continue token, meaning "which namespace should we fetch from next and from which page"
	if pageToken != "" {
		var pageTokenInfo PageTokenInfo

		err := decodeFromBase64(&pageTokenInfo, pageToken)
		if err != nil {
			return "", fmt.Errorf("error decoding next token %w", err)
		}

		currentNamespaceIndex = pageTokenInfo.NamespaceIndex
		currentK8sPageToken = pageTokenInfo.K8sPageToken
	}

	// We don't need to iterate over namespaces from the same place all the time
	// It all will depend on the namespace index coming in pageToken parameter.
	for nsIndex, ns := range namespaceList[currentNamespaceIndex:] {
		nextPageToken, pageLen, err := getRawPage(ns.Name, itemsLeft, currentK8sPageToken)
		if err != nil {
			continue
		}

		itemsLeft = itemsLeft - pageLen
		globalNamespaceIndex := nsIndex + currentNamespaceIndex

		// We will return here every time k8s api responds with a NOT empty string in nextPageToken.
		// This is because the k8s api responds that way when we have fetched the resources we wanted,
		// in this case itemsLeft.
		// When will it fall in this condition?:
		// 1.- We have fetched the itemLeft number of resources AND there are more items to fetch in this particular namespace
		if nextPageToken != "" {
			newPageToken, err = getPageTokenInfoBase64(globalNamespaceIndex, nextPageToken, ns.Name)
			if err != nil {
				return "", err
			}

			break
		}

		// There might be cases where we requested exactly the number of resources in the namespace.
		// In this case nextPageToken will be empty forcing us to verify in a different way
		// if we already have fetched the resources requested. That is why we check for itemsLeft here.
		// But only as long as there are more namespaces to fetch, otherwise we need to return an empty string
		// Which will happen in the next attempt of the namespace loop in the iteration.
		// When will it fall in this condition?:
		//  1.- We have fetched the requested number of resources and there is no more items to fetch in the current namespace.
		// But there are more namespaces to fetch from in `namespaceList`
		if itemsLeft == 0 && globalNamespaceIndex+1 < len(namespaceList) {
			newPageToken, err = getPageTokenInfoBase64(globalNamespaceIndex+1, "", namespaceList[globalNamespaceIndex+1].Name)
			if err != nil {
				return "", err
			}

			break
		}

		// If we haven't fetched the items requested in PageSize we need to keep fetching from the next namespace
	}
	// At this point it doesn't matter the value of newPageToken
	// If it is empty it means there are no more namespaces to fetch from
	// and we have fetched all the resources we could.

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
