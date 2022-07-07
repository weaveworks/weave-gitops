package templates_test

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/cmd/gitops/root"
)

func TestEndpointNotSet(t *testing.T) {
	client := resty.New()
	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"get", "templates",
	})

	err := cmd.Execute()
	assert.EqualError(t, err, "the Weave GitOps Enterprise HTTP API endpoint flag (--endpoint) has not been set")
}

func TestProviderIsNotValid(t *testing.T) {
	client := resty.New()

	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"get", "template",
		"--provider",
		"--endpoint", "http://localhost:8000",
		"--skip-auth",
	})

	err := cmd.Execute()
	assert.EqualError(t, err, "provider \"--endpoint\" is not valid")
}

func TestTemplateNameIsRequired(t *testing.T) {
	client := resty.New()

	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"get", "template",
		"--list-parameters",
		"--provider", "aws",
		"--endpoint", "http://localhost:8000",
		"--skip-auth",
	})

	err := cmd.Execute()
	assert.EqualError(t, err, "template name is required")
}
