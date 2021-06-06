package cmdimpl

// Implementation of the 'wego add' command

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/shims"
	"github.com/weaveworks/weave-gitops/pkg/status"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

type DeploymentType string
type SourceType string
type ConfigType string

const appYamlTemplate = `apiVersion: wego.weave.works/v1alpha1
kind: Application
metadata:
  name: {{ .AppName }}
spec:
  path: {{ .AppPath }}
  url: {{ .AppURL }}
`

const (
	DeployTypeKustomize DeploymentType = "kustomize"
	DeployTypeHelm      DeploymentType = "helm"

	SourceTypeGit  SourceType = "git"
	SourceTypeHelm SourceType = "helm"

	ConfigTypeUserRepo ConfigType = ""
	ConfigTypeNone     ConfigType = "NONE"
)

type AddParamSet struct {
	Dir            string
	Name           string
	Owner          string
	Url            string
	Path           string
	Branch         string
	PrivateKey     string
	PrivateKeyPass string
	DeploymentType string
	Chart          string
	SourceType     string
	AppConfigUrl   string
	Namespace      string
	DryRun         bool
}

var (
	params AddParamSet
)

type AddDependencies struct {
	GitClient git.Git
}

func sanitizePath(path string) string {
	trimmed := strings.TrimSuffix(path, "/")
	return strings.ReplaceAll(strings.ReplaceAll(trimmed, "/", "-"), ".", "dot")
}

func generateAppName(dirOrUrl string) string {
	base := strings.ReplaceAll(filepath.Base(dirOrUrl), "_", "-")
	if params.Path != "./" {
		return base + "-" + sanitizePath(params.Path)
	}
	return base
}

func updateParametersIfNecessary(gitClient git.Git) error {
	if params.Name == "" {
		if params.Url != "" {
			params.Name = generateAppName(strings.TrimSuffix(params.Url, ".git"))
		} else {
			repoPath, err := filepath.Abs(params.Dir)

			if err != nil {
				return wrapError(err, "could not get directory")
			}

			params.Name = generateAppName(repoPath)
		}
	}

	if params.Url == "" && !params.DryRun {
		urls, err := getUrls(gitClient, "origin")
		if err != nil {
			return err
		}

		if len(urls) == 0 {
			return fmt.Errorf("remote config in %s does not have an url", params.Dir)
		}

		params.Url = urls[0]

	}

	sshPrefix := "git@github.com:"
	if strings.HasPrefix(params.Url, sshPrefix) {
		trimmed := strings.TrimPrefix(params.Url, sshPrefix)
		params.Url = "ssh://git@github.com/" + trimmed
	}

	fmt.Printf("using URL: '%s' of origin from git config...\n\n", params.Url)

	return nil
}

func generateSourceManifestGit() ([]byte, error) {
	secretName := params.Name

	cmd := fmt.Sprintf(`create secret git "%s" \
            --url="%s" \
            --private-key-file="%s" \
            --namespace=%s`,
		secretName,
		params.Url,
		params.PrivateKey,
		params.Namespace)
	if params.DryRun {
		fmt.Printf(cmd + "\n")
	} else {

		_, err := fluxops.CallFlux(cmd)

		if err != nil {
			return nil, wrapError(err, "could not create git secret")
		}
	}

	cmd = fmt.Sprintf(`create source git "%s" \
            --url="%s" \
            --branch="%s" \
            --secret-ref="%s" \
            --interval=30s \
            --export \
            --namespace=%s `,
		params.Name,
		params.Url,
		params.Branch,
		secretName,
		params.Namespace)
	sourceManifest, err := fluxops.CallFlux(cmd)
	if err != nil {
		return nil, wrapError(err, "could not create git source")
	}
	return sourceManifest, nil
}

func generateSourceManifestHelm() ([]byte, error) {
	cmd := fmt.Sprintf(`create source helm %s \
            --url="%s" \
            --interval=30s \
            --export \
            --namespace=%s `,
		params.Name,
		params.Url,
		params.Namespace)

	sourceManifest, err := fluxops.CallFlux(cmd)
	if err != nil {
		return nil, wrapError(err, "could not create git source")
	}
	return sourceManifest, nil
}

func getUrls(client git.Git, remote string) ([]string, error) {
	if _, ok := client.(*git.GoGit); ok {
		repo, err := client.Open(params.Dir)
		if err != nil {
			return nil, wrapError(err, fmt.Sprintf("failed to open repository: %s", params.Dir))
		}

		remote, err := repo.Remote("origin")
		if err != nil {
			return nil, err
		}

		return remote.Config().URLs, nil
	}
	return []string{"ssh://git@github.com/auser/arepo"}, nil
}

func generateKustomizeManifest(sourceName, path string) ([]byte, error) {
	kustName := sourceName + "-" + sanitizePath(path)

	cmd := fmt.Sprintf(`create kustomization "%s" \
                --path="%s" \
                --source="%s" \
                --prune=true \
                --validation=client \
                --interval=5m \
                --export \
                --namespace=%s`,
		kustName,
		path,
		sourceName,
		params.Namespace)
	kustomizeManifest, err := fluxops.CallFlux(cmd)
	if err != nil {
		return nil, wrapError(err, "could not create kustomization manifest")
	}

	return bytes.ReplaceAll(kustomizeManifest, []byte("path: ./wego"), []byte("path: .wego")), nil
}

func generateHelmManifestGit(sourceName, path string) ([]byte, error) {
	helmName := sourceName + "-" + sanitizePath(path)

	cmd := fmt.Sprintf(`create helmrelease %s \
            --source="GitRepository/%s" \
            --chart="%s" \
            --interval=5m \
            --export \
            --namespace=%s`,
		helmName,
		sourceName,
		path,
		params.Namespace,
	)
	helmManifest, err := fluxops.CallFlux(cmd)
	if err != nil {
		return nil, wrapError(err, "could not create helm manifest")
	}

	return bytes.ReplaceAll(helmManifest, []byte("path: ./wego"), []byte("path: .wego")), nil
}

func generateHelmManifestHelm() ([]byte, error) {
	cmd := fmt.Sprintf(`create helmrelease %s \
            --source="HelmRepository/%s" \
            --chart="%s" \
            --interval=5m \
            --export \
            --namespace=%s`,
		params.Name,
		params.Name,
		params.Chart,
		params.Namespace,
	)

	return fluxops.CallFlux(cmd)
}

func commitAndPush(ctx context.Context, gitClient git.Git) error {
	fmt.Fprintf(shims.Stdout(), "Commiting and pushing wego resources for application...\n")
	if params.DryRun {
		return nil
	}
	_, err := gitClient.Commit(git.Commit{
		Author:  git.Author{Name: "Weave Gitops", Email: "weave-gitops@weave.works"},
		Message: "Add App manifests",
	})
	if err != nil && err != git.ErrNoStagedFiles {
		return fmt.Errorf("failed to commit sync manifests: %w", err)
	}
	if err == nil {
		fmt.Println("Pushing app manifests to repository")
		if err = gitClient.Push(ctx); err != nil {
			return fmt.Errorf("failed to push manifests: %w", err)
		}
	} else {
		fmt.Println("App manifests are up to date")
	}

	return nil
}

func wrapError(err error, msg string) error {
	return fmt.Errorf("%s: %w", msg, err)
}

// Add provides the implementation for the wego add command

// Three models:
// --app-config-url=none
//
// - Source created for user repo (GitRepository or HelmRepository)
// - app.yaml created for app
// - HelmRelease or Kustomize created for app dir within user repo
// - app.yaml, Source, Helm Release or Kustomize applied directly to cluster
//
// --app-config-url=<URL>
//
// - Separate GOAT repo
// - Source created for GOAT repo
// - Kustomize created for targets/<target name> directory in GOAT repo
// - Kustomize created for apps/<app name> directory within GOAT repo
// - Source, Kustomizes applied directly to cluster
// - app.yaml created for app
// - app.yaml placed in apps/<app name>/app.yaml in GOAT repo
// - Source created for user repo (GitRepository or HelmRepository)
// - User repo Source placed in targets/<target name>/<app-name>/<app name>-gitops-runtime.yaml in GOAT repo
// - HelmRelease or Kustomize referencing user repo source created for user app dir within user repo
// - User app dir HelmRelease or Kustomize placed in targets/<target name>/<app name>/<app name>-gitops-runtime.yaml in GOAT repo
// - PR created or commit directly pushed for GOAT repo
//
// --app-config-url="" (default)
//
// - Source created for user repo (GitRepository only)
// - Kustomize created for .wego/targets/<target name> directory in user repo
// - Kustomize created for .wego/apps/<app name> directory within user repo
// - Source, Kustomizes applied directly to cluster
// - app.yaml created for app
// - app.yaml placed in apps/<app name>/app.yaml in .wego directory within user repo
// - HelmRelease or Kustomize referencing user repo source created for app dir within user repo
// - User app dir HelmRelease or Kustomize placed in targets/<target name>/<app name>/<app name>-gitops-runtime.yaml in .wego
//   directory within user repo
// - PR created or commit directly pushed for user repo

func Add(args []string, allParams AddParamSet, deps *AddDependencies) error {
	ctx := context.Background()

	if allParams.Url == "" {
		if len(args) == 0 {
			return errors.New("no app --url or app location specified")
		} else {
			allParams.Dir = args[0]
		}
	}

	params = allParams
	fmt.Print("Updating parameters from environment... ")
	if err := updateParametersIfNecessary(deps.GitClient); err != nil {
		return wrapError(err, "could not update parameters")
	}
	fmt.Print("done\n\n")
	fmt.Print("Checking cluster status... ")
	clusterStatus := status.GetClusterStatus()
	fmt.Printf("%s\n\n", clusterStatus)

	switch clusterStatus {
	case status.Unmodified:
		return errors.New("WeGO not installed... exiting")
	case status.Unknown:
		return errors.New("WeGO can not determine cluster status... exiting")
	}

	switch strings.ToUpper(params.AppConfigUrl) {
	case string(ConfigTypeNone):
		return addAppWithNoConfigRepo()
	case string(ConfigTypeUserRepo):
		return addAppWithConfigInUserRepo(ctx, deps.GitClient)
	default:
		return addAppWithConfigInExternalRepo(ctx, deps.GitClient)
	}
}

func addAppWithNoConfigRepo() error {
	// Source covers entire user repo
	userSourceName := getUserRepoName()
	userRepoSource, err := generateSource(userSourceName, params.Url, SourceTypeGit)
	if err != nil {
		return wrapError(err, "could not set up GitOps for user repository")
	}
	appYaml, err := generateAppYaml()
	if err != nil {
		return wrapError(err, fmt.Sprintf("could not create app.yaml for '%s'", params.Name))
	}
	// kustomize or helm referencing single user repo source
	applicationGOAT, err := generateApplicationGoat(userSourceName)
	if err != nil {
		return wrapError(err, fmt.Sprintf("could not create GitOps automation for '%s'", params.Name))
	}
	return applyToCluster(userRepoSource, appYaml, applicationGOAT)
}

func addAppWithConfigInUserRepo(ctx context.Context, gitClient git.Git) error {
	// Source covers entire user repo
	userSourceName := getUserRepoName()
	userRepoSource, err := generateSource(userSourceName, params.Url, SourceTypeGit)
	if err != nil {
		return err
	}
	targetKustomize, err := generateTargetKustomize(userSourceName, ".wego")
	if err != nil {
		return err
	}
	appKustomize, err := generateAppKustomize(userSourceName, ".wego")
	if err != nil {
		return err
	}
	if err := applyToCluster(userRepoSource, targetKustomize, appKustomize); err != nil {
		return err
	}
	appYaml, err := generateAppYaml()
	if err != nil {
		return wrapError(err, fmt.Sprintf("could not create app.yaml for '%s'", params.Name))
	}
	applicationGoat, err := generateApplicationGoat(userSourceName)
	if err != nil {
		return wrapError(err, fmt.Sprintf("could not create GitOps automation for '%s'", params.Name))
	}
	if err := writeAppYaml(gitClient, appYaml, ".wego"); err != nil {
		return err
	}
	if err := writeGoats(gitClient, ".wego", applicationGoat); err != nil {
		return err
	}
	return commitAndPush(ctx, gitClient)
}

func addAppWithConfigInExternalRepo(ctx context.Context, gitClient git.Git) error {
	// Source covers entire GOAT repo
	goatSourceName := getGoatRepoName()
	tmpDir, err := cloneToTempDir(ctx, gitClient) // leaves gitClient referencing cloned repo
	if err != nil {
		return err
	}
	defer os.RemoveAll(tmpDir)

	goatRepoSource, err := generateSource(goatSourceName, params.AppConfigUrl, SourceTypeGit)
	if err != nil {
		return err
	}
	goatTargetKustomize, err := generateTargetKustomize(goatSourceName, ".")
	if err != nil {
		return err
	}
	goatAppKustomize, err := generateAppKustomize(goatSourceName, ".")
	if err != nil {
		return err
	}
	if err := applyToCluster(goatRepoSource, goatTargetKustomize, goatAppKustomize); err != nil {
		return err
	}
	// Source covers entire user repo
	userSourceName := getUserRepoName()
	userRepoSource, err := generateSource(userSourceName, params.Url, SourceType(params.SourceType))
	if err != nil {
		return err
	}
	userTargetKustomize, err := generateTargetKustomize(userSourceName, ".")
	if err != nil {
		return err
	}
	userAppKustomize, err := generateAppKustomize(userSourceName, ".")
	if err != nil {
		return err
	}
	appYaml, err := generateAppYaml()
	if err != nil {
		return wrapError(err, fmt.Sprintf("could not create app.yaml for '%s'", params.Name))
	}
	applicationGoat, err := generateApplicationGoat(userSourceName)
	if err != nil {
		return wrapError(err, fmt.Sprintf("could not create GitOps automation for '%s'", params.Name))
	}
	if err := writeAppYaml(gitClient, appYaml, "."); err != nil {
		return err
	}
	if err := writeGoats(gitClient, "", userRepoSource, userTargetKustomize, userAppKustomize, applicationGoat); err != nil {
		return err
	}
	return commitAndPush(ctx, gitClient)
}

func generateSource(repoName, repoUrl string, sourceType SourceType) ([]byte, error) {
	secretName := repoName

	cmd := fmt.Sprintf(`create secret git "%s" \
            --url="%s" \
            --private-key-file="%s" \
            --namespace="%s"`,
		secretName,
		repoUrl,
		params.PrivateKey,
		params.Namespace)
	if params.DryRun {
		fmt.Printf(cmd + "\n")
	} else {
		_, err := fluxops.CallFlux(cmd)

		if err != nil {
			return nil, wrapError(err, "could not create git secret")
		}
	}

	cmd = fmt.Sprintf(`create source git "%s" \
            --url="%s" \
            --branch="%s" \
            --secret-ref="%s" \
            --interval=30s \
            --export \
            --namespace="%s"`,
		repoName,
		repoUrl,
		params.Branch,
		secretName,
		params.Namespace)
	sourceManifest, err := fluxops.CallFlux(cmd)
	if err != nil {
		return nil, wrapError(err, "could not create git source")
	}
	return sourceManifest, nil
}

func generateAppYaml() ([]byte, error) {
	// Create app.yaml
	t, err := template.New("appYaml").Parse(appYamlTemplate)
	if err != nil {
		return nil, wrapError(err, "could not parse app yaml template")
	}

	var populated bytes.Buffer
	err = t.Execute(&populated, struct {
		AppName string
		AppPath string
		AppURL  string
	}{params.Name, params.Path, params.Url})
	if err != nil {
		return nil, wrapError(err, "could not execute populated template")
	}
	return populated.Bytes(), nil
}

func generateApplicationGoat(sourceName string) ([]byte, error) {
	switch params.DeploymentType {
	case string(DeployTypeKustomize):
		return generateKustomizeManifest(sourceName, params.Path)
	case string(DeployTypeHelm):
		return generateHelmManifest(sourceName, params.Path)
	default:
		return nil, fmt.Errorf("Invalid deployment type: %v", params.DeploymentType)
	}
}

func generateTargetKustomize(sourceName, basePath string) ([]byte, error) {
	clusterName, err := status.GetClusterName()
	if err != nil {
		return nil, err
	}
	return generateKustomizeManifest(sourceName, filepath.Join(basePath, "targets", clusterName))
}

func generateAppKustomize(sourceName, basePath string) ([]byte, error) {
	return generateKustomizeManifest(sourceName, filepath.Join(basePath, "apps", params.Name))
}

func applyToCluster(manifests ...[]byte) error {
	if params.DryRun {
		fmt.Printf("Applying:\n\n")
		for _, manifest := range manifests {
			fmt.Printf("%s\n", manifest)
		}
		return nil
	}

	kubectlApply := fmt.Sprintf("kubectl apply --namespace=%s -f -", params.Namespace)

	for _, manifest := range manifests {
		if err := utils.CallCommandForEffectWithInputPipe(kubectlApply, string(manifest)); err != nil {
			return wrapError(err, fmt.Sprintf("could not apply manifest: %s", manifest))
		}
	}

	return nil
}

func writeAppYaml(gitClient git.Git, appYaml []byte, basePath string) error {
	appYamlPath := filepath.Join(basePath, "apps", params.Name, "app.yaml")
	if params.DryRun {
		fmt.Printf("Writing app.yaml to '%s'\n", appYamlPath)
		return nil
	}
	return gitClient.Write(appYamlPath, appYaml)
}

func writeGoats(gitClient git.Git, basePath string, manifests ...[]byte) error {
	clusterName, err := status.GetClusterName()
	if err != nil {
		return err
	}
	goatPath := filepath.Join(basePath, "targets", clusterName, params.Name, fmt.Sprintf("%s-gitops-runtime.yaml", params.Name))
	if params.DryRun {
		fmt.Printf("Writing GitOps Automation to '%s'\n", goatPath)
		return nil
	}
	goat := bytes.Join(manifests, []byte(""))
	return gitClient.Write(goatPath, goat)
}

func cloneToTempDir(ctx context.Context, gitClient git.Git) (string, error) {
	gitDir, err := ioutil.TempDir("", "git-")
	if err != nil {
		return "", wrapError(err, "TempDir")
	}
	if params.DryRun {
		fmt.Printf("Cloning '%s' to temporary directory\n", params.AppConfigUrl)
		return gitDir, nil
	}
	if _, err := gitClient.Clone(ctx, gitDir, params.AppConfigUrl, params.Branch); err != nil {
		return "", err
	}
	return gitDir, nil
}

func getUserRepoName() string {
	return urlToRepoName(strings.TrimSuffix(params.Url, ".git"))
}

func getGoatRepoName() string {
	return urlToRepoName(strings.TrimSuffix(params.AppConfigUrl, ".git"))
}

func urlToRepoName(url string) string {
	return strings.ReplaceAll(filepath.Base(url), "_", "-")
}
