package flux

import (
	"fmt"
	"strings"

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
	GetBinPath() (string, error)
	GetExePath() (string, error)
	CreateSourceGit(name string, repoUrl gitproviders.RepoURL, branch string, secretRef string, namespace string) ([]byte, error)
	CreateSourceHelm(name string, url string, namespace string) ([]byte, error)
	CreateKustomization(name string, source string, path string, namespace string) ([]byte, error)
	CreateHelmReleaseGitRepository(name, source, path, namespace, targetNamespace string) ([]byte, error)
	CreateHelmReleaseHelmRepository(name, chart, namespace, targetNamespace string) ([]byte, error)
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

func (f *FluxClient) CreateSourceGit(name string, repoUrl gitproviders.RepoURL, branch string, secretRef string, namespace string) ([]byte, error) {
	args := []string{
		"create", "source", "git", name,
		"--branch", branch,
		"--namespace", namespace,
		"--interval", "30s",
		"--export",
	}

	if secretRef != "" {
		args = append(args, "--secret-ref", secretRef, "--url", repoUrl.String())
	} else {
		args = append(args, "--url", makePublicUrl(repoUrl))
	}

	out, err := f.runFluxCmd(args...)
	if err != nil {
		return out, fmt.Errorf("failed to create source git: %w", err)
	}

	return out, nil
}

func makePublicUrl(repoUrl gitproviders.RepoURL) string {
	trimmed := ""

	url := repoUrl.String()
	provider := repoUrl.Provider()

	if !strings.HasSuffix(url, ".git") {
		url = url + ".git"
	}

	gitSshPrefix := fmt.Sprintf("git@%scom:", provider)
	sshPrefix := fmt.Sprintf("ssh://git@%s.com/", provider)

	if strings.HasPrefix(url, gitSshPrefix) {
		trimmed = strings.TrimPrefix(url, gitSshPrefix)
	} else if strings.HasPrefix(url, sshPrefix) {
		trimmed = strings.TrimPrefix(url, sshPrefix)
	}

	if trimmed != "" {
		return fmt.Sprintf("https://%s.com/%s", provider, trimmed)
	}

	return url
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
		"--interval", "1m",
		"--export",
	}

	out, err := f.runFluxCmd(args...)
	if err != nil {
		return out, fmt.Errorf("failed to create kustomization: %w", err)
	}

	return out, nil
}

func (f *FluxClient) CreateHelmReleaseGitRepository(name, source, chartPath, namespace, targetNamespace string) ([]byte, error) {
	args := []string{
		"create", "helmrelease", name,
		"--source", "GitRepository/" + source,
		"--chart", chartPath,
		"--namespace", namespace,
		"--interval", "5m",
		"--export",
	}

	if targetNamespace != "" {
		args = append(args, "--target-namespace", targetNamespace)
	}

	out, err := f.runFluxCmd(args...)
	if err != nil {
		return out, fmt.Errorf("failed to create helm release git repo: %w", err)
	}

	return out, nil
}

func (f *FluxClient) CreateHelmReleaseHelmRepository(name, chart, namespace, targetNamespace string) ([]byte, error) {
	args := []string{
		"create", "helmrelease", name,
		"--source", "HelmRepository/" + name,
		"--chart", chart,
		"--namespace", namespace,
		"--interval", "5m",
		"--export",
	}

	if targetNamespace != "" {
		args = append(args, "--target-namespace", targetNamespace)
	}

	out, err := f.runFluxCmd(args...)
	if err != nil {
		return out, fmt.Errorf("failed to create helm release helm repo: %w", err)
	}

	return out, nil
}

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
