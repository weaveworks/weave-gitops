package templates_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/cmd/gitops/root"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
)

func TestEndpointNotSet(t *testing.T) {
	client := adapters.NewHTTPClient()
	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"get", "templates",
	})

	err := cmd.Execute()
	assert.EqualError(t, err, "the Weave GitOps Enterprise HTTP API endpoint flag (--endpoint) has not been set")
}

func TestProviderIsNotValid(t *testing.T) {
	client := adapters.NewHTTPClient()

	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"get", "template",
		"--provider",
		"--endpoint", "http://localhost:8000",
	})

	err := cmd.Execute()
	assert.EqualError(t, err, "provider \"--endpoint\" is not valid")
}

func TestTemplateNameIsRequired(t *testing.T) {
	client := adapters.NewHTTPClient()

	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"get", "template",
		"--list-parameters",
		"--provider", "aws",
		"--endpoint", "http://localhost:8000",
	})

	err := cmd.Execute()
	assert.EqualError(t, err, "template name is required")
}
