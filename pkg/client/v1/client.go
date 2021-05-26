package client

import (
	"net/http"

	"github.com/weaveworks/weave-gitops/pkg/rpc/gitops"
)

func NewClient(url string, c *http.Client) gitops.GitOps {
	return gitops.NewGitOpsJSONClient(url, c)
}
