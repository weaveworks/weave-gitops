package terraform_test

import (
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/root"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

func TestEndpointNotSet(t *testing.T) {
	client := resty.New()
	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"add", "terraform",
		"--from-template=terraform-template",
		"--url=https://github.com/weaveworks/test-repo",
		"--set=CLUSTER_NAME=dev",
		"--set=TEMPLATE_NAME=aurora",
		"--set=NAMESPACE=default",
		"--set=TEMPLATE_PATH=./",
		"--set=GIT_REPO_NAME=test-repo",
	})

	err := cmd.Execute()
	assert.EqualError(t, err, "the Weave GitOps Enterprise HTTP API endpoint flag (--endpoint) has not been set")
}

func TestGitProviderToken(t *testing.T) {
	t.Cleanup(testutils.Setenv("GITHUB_TOKEN", "test-token"))

	client := resty.New()

	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		http.MethodPost,
		"http://localhost:8000/v1/tfcontrollers",
		func(r *http.Request) (*http.Response, error) {
			h, ok := r.Header["Git-Provider-Token"]
			assert.True(t, ok)
			assert.Contains(t, h, "test-token")

			return httpmock.NewJsonResponse(http.StatusOK, httpmock.File("../../../../pkg/adapters/testdata/pull_request_created.json"))
		},
	)

	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"add", "terraform",
		"--from-template=terraform-template",
		"--url=https://github.com/weaveworks/test-repo",
		"--set=CLUSTER_NAME=dev",
		"--set=TEMPLATE_NAME=aurora",
		"--set=NAMESPACE=default",
		"--set=TEMPLATE_PATH=./",
		"--set=GIT_REPO_NAME=test-repo",
		"--endpoint", "http://localhost:8000",
		"--skip-auth",
	})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestGitProviderToken_NoURL(t *testing.T) {
	client := resty.New()

	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"add", "terraform",
		"--from-template=terraform-template",
		"--set=CLUSTER_NAME=dev",
		"--set=TEMPLATE_NAME=aurora",
		"--set=NAMESPACE=default",
		"--set=TEMPLATE_PATH=./",
		"--set=GIT_REPO_NAME=test-repo",
		"--endpoint", "http://localhost:8000",
		"--skip-auth",
	})

	err := cmd.Execute()
	assert.ErrorIs(t, err, cmderrors.ErrNoURL)
}

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
		"--skip-auth",
	})

	err := cmd.Execute()
	assert.EqualError(t, err, "cannot parse url: could not get provider name from URL invalid_url: no git providers found for \"invalid_url\"")
}
