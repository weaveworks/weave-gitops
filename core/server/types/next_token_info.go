package types

type PageTokenInfo struct {
	NamespaceIndex int    `json:"namespace_index"`
	Namespace      string `json:"namespace"`
	K8sPageToken   string `json:"k8s_page_token"`
}
