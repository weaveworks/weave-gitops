package credentials_test

import (
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/gitops/cmd/gitops/root"
)

func TestEndpointNotSet(t *testing.T) {
	client := resty.New()
	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"get", "credentials",
	})

	err := cmd.Execute()
	assert.EqualError(t, err, "the Weave GitOps Enterprise HTTP API endpoint flag (--endpoint) has not been set")
}
