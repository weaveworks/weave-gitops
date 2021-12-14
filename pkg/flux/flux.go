package flux

import (
	"fmt"
	"strings"

	"github.com/pkg/errors"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
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
	Install(namespace string, export bool) ([]byte, error)
	Uninstall(namespace string, export bool) error
	CreateSourceGit(name string, repoUrl gitproviders.RepoURL, branch, secretRef, namespace string, creds *HTTPSCreds) ([]byte, error)
	CreateSourceHelm(name, url, namespace string) ([]byte, error)
	CreateKustomization(name string, source string, path string, namespace string) ([]byte, error)
	CreateHelmReleaseGitRepository(name, source, path, namespace, targetNamespace string) ([]byte, error)
	CreateHelmReleaseHelmRepository(name, chart, namespace, targetNamespace string) ([]byte, error)
	CreateSecretGit(name string, repoUrl gitproviders.RepoURL, namespace string, creds *HTTPSCreds) ([]byte, error)
	GetVersion() (string, error)
	GetAllResourcesStatus(name string, namespace string) ([]byte, error)
	SuspendOrResumeApp(pause wego.SuspendActionType, name, namespace, deploymentType string) ([]byte, error)
	GetLatestStatusAllNamespaces() ([]string, error)
}

const fluxBinaryPathEnvVar = "WEAVE_GITOPS_FLUX_BIN_PATH"

// HTTPSCreds is an optional username/password use to authenticate flux
// source operations for https based git servers.
type HTTPSCreds struct {
	Username string
	Password string
}

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

func (f *FluxClient) CreateSourceGit(name string, repoUrl gitproviders.RepoURL, branch, secretRef, namespace string, creds *HTTPSCreds) ([]byte, error) {
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

	args = append(args, argsFromCreds(creds)...)
	out, err := f.runFluxCmd(args...)
	if err != nil {
		return out, fmt.Errorf("failed to create source git: %w", err)
	}

	return out, nil
}

func argsFromCreds(creds *HTTPSCreds) []string {
	args := []string{}
	if creds != nil {
		if creds.Username != "" {
			args = append(args, "--username", creds.Username)
		}
		if creds.Password != "" {
			args = append(args, "--password", creds.Password)
		}
	}
	return args
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

// CreateSecretGit Creates a Git secret returns the deploy key.
func (f *FluxClient) CreateSecretGit(name string, repoUrl gitproviders.RepoURL, namespace string, creds *HTTPSCreds) ([]byte, error) {
	args := []string{
		"create", "secret", "git", name,
		"--url", repoUrl.String(),
		"--namespace", namespace,
		"--export",
	}
	args = append(args, argsFromCreds(creds)...)

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
		return []byte{}, fmt.Errorf("failed to run flux %v with output: %s and error: %w", args, out, err)
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

func (f *FluxClient) SuspendOrResumeApp(suspend wego.SuspendActionType, name, namespace string, deploymentType string) ([]byte, error) {
	args := []string{
		string(suspend), deploymentType, name, fmt.Sprintf("--namespace=%s", namespace),
	}

	return f.runFluxCmd(args...)
}
