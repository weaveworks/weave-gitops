package terraform_test

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/cmd/gitops/root"
)

func TestGitProviderToken_InvalidURL(t *testing.T) {
	client := resty.New()

	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"add", "terraform",
		"--from-template=terraform-template",
		"--url=invalid_url",
		"--set=CLUSTER_NAME=dev",
		"--set=TEMPLATE_NAME=aurora",
		"--set=NAMESPACE=default",
		"--set=TEMPLATE_PATH=./",
		"--set=GIT_REPO_NAME=test-repo",
		"--endpoint", "http://localhost:8000",
	})

	err := cmd.Execute()
	assert.EqualError(t, err, "cannot parse url: could not get provider name from URL invalid_url: no git providers found for \"invalid_url\"")
}
