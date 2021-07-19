package flux

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/version"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

//counterfeiter:generate . Flux
type Flux interface {
	Install(namespace string, export bool) ([]byte, error)
	Uninstall(namespace string, export bool) error
	CreateSourceGit(name string, url string, branch string, secretRef string, namespace string) ([]byte, error)
	CreateSourceHelm(name string, url string, namespace string) ([]byte, error)
	CreateKustomization(name string, source string, path string, namespace string) ([]byte, error)
	CreateHelmReleaseGitRepository(name string, source string, path string, namespace string) ([]byte, error)
	CreateHelmReleaseHelmRepository(name string, chart string, namespace string) ([]byte, error)
	CreateSecretGit(name string, url string, namespace string) ([]byte, error)
	GetVersion() (string, error)
	GetAllResourcesStatus(name string, namespace string) ([]byte, error)
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

func (f *FluxClient) Install(namespace string, export bool) ([]byte, error) {
	args := []string{
		"install",
		"--namespace", namespace,
		"--components-extra", "image-reflector-controller,image-automation-controller",
	}

	if export {
		args = append(args, "--export")

		out, err := f.runFluxCmd(args...)
		if err != nil {
			return out, errors.Wrapf(err, "failed to run flux install: %s", string(out))
		}

		return out, nil
	}

	if _, err := f.runFluxCmdOutputStream(args...); err != nil {
		return []byte{}, errors.Wrap(err, "failed to run flux binary")
	}

	return []byte{}, nil
}

func (f *FluxClient) Uninstall(namespace string, dryRun bool) error {
	args := []string{
		"uninstall", "-s",
		"--namespace", namespace,
	}

	if dryRun {
		args = append(args, "--dry-run")
	}

	if _, err := f.runFluxCmdOutputStream(args...); err != nil {
		return errors.Wrap(err, "failed to run flux binary")
	}

	return nil
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

func (f *FluxClient) GetVersion() (string, error) {
	out, err := f.runFluxCmd("-v")
	if err != nil {
		return "", err
	}
	// change string format to match our versioning standard
	version := strings.ReplaceAll(string(out), "flux version ", "v")
	return version, nil
}

func (f *FluxClient) runFluxCmd(args ...string) ([]byte, error) {
	fluxPath, err := f.fluxPath()
	if err != nil {
		return []byte{}, errors.Wrap(err, "error getting flux binary path")
	}
	out, err := f.runner.Run(fluxPath, args...)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to run flux with output: %s", string(out))
	}

	return out, nil
}

func (f *FluxClient) runFluxCmdOutputStream(args ...string) ([]byte, error) {
	fluxPath, err := f.fluxPath()
	if err != nil {
		return []byte{}, errors.Wrap(err, "error getting flux binary path")
	}
	out, err := f.runner.RunWithOutputStream(fluxPath, args...)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to run flux with output: %s", string(out))
	}

	return out, nil
}

func (f *FluxClient) fluxPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", errors.Wrap(err, "failed getting user home directory")
	}
	path := fmt.Sprintf("%v/.wego/bin", homeDir)
	return fmt.Sprintf("%v/flux-%v", path, version.FluxVersion), nil
}
