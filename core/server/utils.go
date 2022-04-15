package server

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"github.com/weaveworks/weave-gitops/core/server/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"strings"
)

func getMatchingLabels(appName string) client.MatchingLabels {
	return matchLabel(withPartOfLabel(appName))
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

func getPageTokenInfoBase64(namespaceIndex int, k8sNextToken string, namespace string) (string, error) {

	nextTokenInfo := types.PageTokenInfo{
		NamespaceIndex: namespaceIndex,
		K8sPageToken:   k8sNextToken,
		Namespace:      namespace,
	}

	nextToken, err := encodeToBase64(nextTokenInfo)
	if err != nil {
		return "", err
	}

	//	fmt.Printf("NextToken(decoded) %+v\n", nextTokenInfo)
	//	fmt.Println("NextToken(encoded)", nextToken)

	return nextToken, nil
}
