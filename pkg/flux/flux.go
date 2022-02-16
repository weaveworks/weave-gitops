package flux

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/version"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Flux
type Flux interface {
	SetupBin()
	CreateSecretGit(name string, repoUrl gitproviders.RepoURL, namespace string) ([]byte, error)
	GetAllResourcesStatus(name string, namespace string) ([]byte, error)
	PreCheck() (string, error)
}

const (
	PartOfLabelKey   = "app.kubernetes.io/part-of"
	PartOfLabelValue = "flux"
	VersionLabelKey  = "app.kubernetes.io/version"
)

const fluxBinaryPathEnvVar = "WEAVE_GITOPS_FLUX_BIN_PATH"

type FluxClient struct {
	osys   osys.Osys
	runner runner.Runner
}

func New(osysClient osys.Osys, cliRunner runner.Runner) *FluxClient {
	return &FluxClient{
		osys:   osysClient,
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

func (f *FluxClient) GetAllResourcesStatus(name string, namespace string) ([]byte, error) {
	args := []string{
		"get", "all", "--namespace", namespace, name,
	}

	out, err := f.runFluxCmd(args...)
	if err != nil {
		return out, fmt.Errorf("failed to get flux resources status: %w", err)
	}

	return out, nil
}

func (f *FluxClient) runFluxCmd(args ...string) ([]byte, error) {
	fluxPath, err := f.fluxPath()
	if err != nil {
		return []byte{}, errors.Wrap(err, "error getting flux binary path")
	}

	out, err := f.runner.Run(fluxPath, args...)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to run flux with output: %s and error: %w", string(out), err)
	}

	return out, nil
}

func (f *FluxClient) fluxPath() (string, error) {
	homeDir, err := f.osys.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "failed getting user home directory")
	}

	path := fmt.Sprintf("%v/.wego/bin", homeDir)

	return fmt.Sprintf("%v/flux-%v", path, version.FluxVersion), nil
}

func (f *FluxClient) PreCheck() (string, error) {
	args := []string{
		"check",
		"--pre",
	}

	output, err := f.runFluxCmd(args...)
	if err != nil {
		return "", fmt.Errorf("failed running flux pre check %w", err)
	}

	return string(output), nil
}
