package flux

import (
	"bytes"
	"fmt"

	"github.com/pkg/errors"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/shims"
	"github.com/weaveworks/weave-gitops/pkg/version"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Flux
type Flux interface {
	CreateSourceGit(name string, url string, branch string, secretRef string, namespace string) ([]byte, error)
	CreateSourceHelm(name string, url string, namespace string) ([]byte, error)
	CreateKustomization(name string, source string, path string, namespace string) ([]byte, error)
	CreateHelmReleaseGitRepository(name string, source string, path string, namespace string) ([]byte, error)
	CreateHelmReleaseHelmRepository(name string, chart string, namespace string) ([]byte, error)
	CreateSecretGit(name string, url string, namespace string) ([]byte, error)
}

type FluxClient struct {
	runner runner.Runner
}

func New() Flux {
	return &FluxClient{
		runner: &runner.CLIRunner{},
	}
}

func (f *FluxClient) CreateSourceGit(name string, url string, branch string, secretRef string, namespace string) ([]byte, error) {
	args := []string{
		"create", "source", "git", name,
		"--url", url,
		"--branch", branch,
		"--secret-ref", secretRef,
		"--namespace", namespace,
		"--interval", "30s",
		"--export",
	}

	out, err := f.runFluxCmd(args...)
	if err != nil {
		return out, fmt.Errorf("failed to create source git: %w", err)
	}

	return out, nil
}

func (f *FluxClient) CreateSourceHelm(name string, url string, namespace string) ([]byte, error) {
	args := []string{
		"create", "source", "helm", name,
		"--url", url,
		"--namespace", namespace,
		"--interval", "30s",
		"--export",
	}

	out, err := f.runFluxCmd(args...)
	if err != nil {
		return out, fmt.Errorf("failed to create source helm: %w", err)
	}

	return out, nil
}

func (f *FluxClient) CreateKustomization(name string, source string, path string, namespace string) ([]byte, error) {
	args := []string{
		"create", "kustomization", name,
		"--path", path,
		"--source", source,
		"--namespace", namespace,
		"--prune", "true",
		"--validation", "client",
		"--interval", "1m",
		"--export",
	}

	out, err := f.runFluxCmd(args...)
	if err != nil {
		return out, fmt.Errorf("failed to create kustomization: %w", err)
	}

	return out, nil
}

func (f *FluxClient) CreateHelmReleaseGitRepository(name string, source string, chartPath string, namespace string) ([]byte, error) {
	args := []string{
		"create", "helmrelease", name,
		"--source", "GitRepository/" + source,
		"--chart", chartPath,
		"--namespace", namespace,
		"--interval", "5m",
		"--export",
	}

	out, err := f.runFluxCmd(args...)
	if err != nil {
		return out, fmt.Errorf("failed to create helm release git repo: %w", err)
	}

	return out, nil
}

func (f *FluxClient) CreateHelmReleaseHelmRepository(name string, chart string, namespace string) ([]byte, error) {
	args := []string{
		"create", "helmrelease", name,
		"--source", "HelmRepository/" + name,
		"--chart", chart,
		"--namespace", namespace,
		"--interval", "5m",
		"--export",
	}

	out, err := f.runFluxCmd(args...)
	if err != nil {
		return out, fmt.Errorf("failed to create helm release helm repo: %w", err)
	}

	return out, nil
}

// CreatSecretGit Creates a Git secret returns the deploy key
func (f *FluxClient) CreateSecretGit(name string, url string, namespace string) ([]byte, error) {
	args := []string{
		"create", "secret", "git", name,
		"--url", url,
		"--namespace", namespace,
	}

	out, err := f.runFluxCmd(args...)
	if err != nil {
		return out, fmt.Errorf("failed to create secret git: %w", err)
	}

	deployKeyBody := bytes.TrimPrefix(out, []byte("✚ deploy key: "))
	deployKeyLines := bytes.Split(deployKeyBody, []byte("\n"))
	if len(deployKeyBody) == 0 {
		return nil, fmt.Errorf("error getting deploy key from flux output: %s", string(out))
	}

	return deployKeyLines[0], nil
}

func (f *FluxClient) runFluxCmd(args ...string) ([]byte, error) {
	fluxPath, err := f.fluxPath()
	if err != nil {
		return []byte{}, errors.Wrap(err, "error getting flux binary path")
	}
	out, err := f.runner.Run(fluxPath, args...)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to run git with output: %s", string(out))
	}

	return out, nil
}

func (f *FluxClient) fluxPath() (string, error) {
	homeDir, err := shims.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("%v/.wego/bin", homeDir)
	return fmt.Sprintf("%v/flux-%v", path, version.FluxVersion), nil
}
