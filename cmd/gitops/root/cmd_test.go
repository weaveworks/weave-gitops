package root_test

import (
	"net/http"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/cmd/gitops/root"
)

func TestInsecureSkipVerifyTrue(t *testing.T) {
	client := resty.New()

	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"add", "cluster",
		"--insecure-skip-tls-verify",
	})

	// Command is incomplete and should raise an error, it helps us short circuit here to quickly
	// test that the client has been set
	err := cmd.Execute()
	assert.Error(t, err)

	transport, ok := client.GetClient().Transport.(*http.Transport)
	assert.True(t, ok)
	assert.True(t, transport.TLSClientConfig.InsecureSkipVerify, "InsecureSkipVerify wasn't set to true")
}

func TestInsecureSkipVerifyFalse(t *testing.T) {
	client := resty.New()

	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"add", "cluster",
	})

	// Command is incomplete and should raise an error, it helps us short circuit here to quickly
	// test that the client has been set
	err := cmd.Execute()
	assert.Error(t, err)

	transport, ok := client.GetClient().Transport.(*http.Transport)
	assert.True(t, ok)
	// Its set to nil and uses whatever the golang defaults are (InsecureSkipVerify: false)
	assert.Nil(t, transport.TLSClientConfig)
}
