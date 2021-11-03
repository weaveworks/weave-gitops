package clusters_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/cmd/gitops/root"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
)

func TestSetSeparateValues(t *testing.T) {
	client := resty.New()

	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		http.MethodPost,
		"http://localhost:8000/v1/templates/cluster-template-eks-fargate/render",
		func(r *http.Request) (*http.Response, error) {
			var vs adapters.TemplateParameterValuesAndCredentials

			err := json.NewDecoder(r.Body).Decode(&vs)
			assert.NoError(t, err)

			assert.Equal(t, "dev", vs.Values["CLUSTER_NAME"])
			assert.Equal(t, "us-east-1", vs.Values["AWS_REGION"])
			assert.Equal(t, "ssh_key", vs.Values["AWS_SSH_KEY_NAME"])
			assert.Equal(t, "1.19", vs.Values["KUBERNETES_VERSION"])

			return httpmock.NewJsonResponse(http.StatusOK, httpmock.File("../../../../pkg/adapters/testdata/rendered_template.json"))
		},
	)

	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"add", "cluster",
		"--from-template=cluster-template-eks-fargate",
		"--set=CLUSTER_NAME=dev",
		"--set=AWS_REGION=us-east-1",
		"--set=AWS_SSH_KEY_NAME=ssh_key",
		"--set=KUBERNETES_VERSION=1.19",
		"--dry-run",
		"--endpoint", "http://localhost:8000",
	})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSetMultipleValues(t *testing.T) {
	client := resty.New()

	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		http.MethodPost,
		"http://localhost:8000/v1/templates/cluster-template-eks-fargate/render",
		func(r *http.Request) (*http.Response, error) {
			var vs adapters.TemplateParameterValuesAndCredentials

			err := json.NewDecoder(r.Body).Decode(&vs)
			assert.NoError(t, err)

			assert.Equal(t, "dev", vs.Values["CLUSTER_NAME"])
			assert.Equal(t, "us-east-1", vs.Values["AWS_REGION"])
			assert.Equal(t, "ssh_key", vs.Values["AWS_SSH_KEY_NAME"])
			assert.Equal(t, "1.19", vs.Values["KUBERNETES_VERSION"])

			return httpmock.NewJsonResponse(http.StatusOK, httpmock.File("../../../../pkg/adapters/testdata/rendered_template.json"))
		},
	)

	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"add", "cluster",
		"--from-template=cluster-template-eks-fargate",
		"--set=CLUSTER_NAME=dev,AWS_REGION=us-east-1,AWS_SSH_KEY_NAME=ssh_key,KUBERNETES_VERSION=1.19",
		"--dry-run",
		"--endpoint", "http://localhost:8000",
	})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestSetMultipleAndSeparateValues(t *testing.T) {
	client := resty.New()

	httpmock.ActivateNonDefault(client.GetClient())
	defer httpmock.DeactivateAndReset()
	httpmock.RegisterResponder(
		http.MethodPost,
		"http://localhost:8000/v1/templates/cluster-template-eks-fargate/render",
		func(r *http.Request) (*http.Response, error) {
			var vs adapters.TemplateParameterValuesAndCredentials

			err := json.NewDecoder(r.Body).Decode(&vs)
			assert.NoError(t, err)

			assert.Equal(t, "dev", vs.Values["CLUSTER_NAME"])
			assert.Equal(t, "us-east-1", vs.Values["AWS_REGION"])
			assert.Equal(t, "ssh_key", vs.Values["AWS_SSH_KEY_NAME"])
			assert.Equal(t, "1.19", vs.Values["KUBERNETES_VERSION"])

			return httpmock.NewJsonResponse(http.StatusOK, httpmock.File("../../../../pkg/adapters/testdata/rendered_template.json"))
		},
	)

	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"add", "cluster",
		"--from-template=cluster-template-eks-fargate",
		"--set=CLUSTER_NAME=dev,AWS_REGION=us-east-1",
		"--set=AWS_SSH_KEY_NAME=ssh_key",
		"--set=KUBERNETES_VERSION=1.19",
		"--dry-run",
		"--endpoint", "http://localhost:8000",
	})

	err := cmd.Execute()
	assert.NoError(t, err)
}

func TestEndpointNotSet(t *testing.T) {
	client := resty.New()
	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"add", "cluster",
		"--from-template=cluster-template-eks-fargate",
		"--set=CLUSTER_NAME=dev",
		"--set=AWS_REGION=us-east-1",
		"--set=AWS_SSH_KEY_NAME=ssh_key",
		"--set=KUBERNETES_VERSION=1.19",
		"--dry-run",
	})

	err := cmd.Execute()
	assert.EqualError(t, err, "the Weave GitOps Enterprise HTTP API endpoint flag (--endpoint) has not been set")
}
