package flux

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Flux
type Flux interface {
	CreateSecretGit(name string, repoUrl gitproviders.RepoURL, namespace string) ([]byte, error)
}

type FluxClient struct {
	runner runner.Runner
}

func New(cliRunner runner.Runner) *FluxClient {
	return &FluxClient{
		runner: cliRunner,
	}
}

var _ Flux = &FluxClient{}

// CreatSecretGit Creates a Git secret returns the deploy key
func (f *FluxClient) CreateSecretGit(name string, repoUrl gitproviders.RepoURL, namespace string) ([]byte, error) {
	args := []string{
		"create", "secret", "git", name,
		"--url", repoUrl.String(),
		"--namespace", namespace,
		"--export",
	}

	out, err := f.runFluxCmd(args...)
	if err != nil {
		return out, fmt.Errorf("failed to create secret git: %w", err)
	}

	return out, nil
}

func (f *FluxClient) runFluxCmd(args ...string) ([]byte, error) {
	out, err := f.runner.Run("flux", args...)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to run flux with output: %s and error: %w", string(out), err)
	}

	return out, nil
}
