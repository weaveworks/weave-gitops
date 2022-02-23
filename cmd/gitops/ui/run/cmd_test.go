package run_test

import (
	"os"
	"testing"

	"github.com/go-resty/resty/v2"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/root"
)

func TestNoClientID(t *testing.T) {
	os.Setenv("WEAVE_GITOPS_AUTH_ENABLED", "true")
	defer os.Unsetenv("WEAVE_GITOPS_AUTH_ENABLED")

	client := resty.New()
	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"ui", "run",
		"--oidc-issuer-url=http://weave.works",
	})

	err := cmd.Execute()
	assert.ErrorIs(t, err, cmderrors.ErrNoClientID)
}

func TestNoClientSecret(t *testing.T) {
	os.Setenv("WEAVE_GITOPS_AUTH_ENABLED", "true")
	defer os.Unsetenv("WEAVE_GITOPS_AUTH_ENABLED")

	client := resty.New()
	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"ui", "run",
		"--oidc-issuer-url=http://weave.works",
		"--oidc-client-id=client-id",
	})

	err := cmd.Execute()
	assert.ErrorIs(t, err, cmderrors.ErrNoClientSecret)
}

func TestNoRedirectURL(t *testing.T) {
	os.Setenv("WEAVE_GITOPS_AUTH_ENABLED", "true")
	defer os.Unsetenv("WEAVE_GITOPS_AUTH_ENABLED")

	client := resty.New()
	cmd := root.RootCmd(client)
	cmd.SetArgs([]string{
		"ui", "run",
		"--oidc-issuer-url=http://weave.works",
		"--oidc-client-id=client-id",
		"--oidc-client-secret=client-secret",
	})

	err := cmd.Execute()
	assert.ErrorIs(t, err, cmderrors.ErrNoRedirectURL)
}
