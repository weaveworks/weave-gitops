package app

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/pkg/errors"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

type DeploymentType string
type SourceType string
type ConfigType string

const (
	DeployTypeKustomize DeploymentType = "kustomize"
	DeployTypeHelm      DeploymentType = "helm"

	SourceTypeGit  SourceType = "git"
	SourceTypeHelm SourceType = "helm"

	ConfigTypeUserRepo ConfigType = ""
	ConfigTypeNone     ConfigType = "NONE"
)

type AddParams struct {
	Dir            string
	Name           string
	Url            string
	Path           string
	Branch         string
	PrivateKey     string
	DeploymentType string
	Chart          string
	SourceType     string
	AppConfigUrl   string
	Namespace      string
	DryRun         bool
	AutoMerge      bool
}

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

func (a *App) Add(params AddParams) error {
	ctx := context.Background()
	fmt.Print("Updating parameters from environment... ")
	params, err := a.updateParametersIfNecessary(params)
	if err != nil {
		return errors.Wrap(err, "could not update parameters")
	}

	fmt.Print("done\n\n")
	fmt.Print("Checking cluster status... ")

	clusterStatus := a.kube.GetClusterStatus(ctx)
	fmt.Printf("%s\n\n", clusterStatus)

	switch clusterStatus {
	case kube.Unmodified:
		return errors.New("WeGO not installed... exiting")
	case kube.Unknown:
		return errors.New("WeGO can not determine cluster status... exiting")
	}

	clusterName, err := a.kube.GetClusterName(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Generating deploy key for repo %s ...\n", params.Url)
	secretRef, err := a.createAndUploadDeployKey(params.Url, SourceType(params.SourceType), clusterName, params.Namespace, params.DryRun)
	if err != nil {
		return errors.Wrap(err, "could not generate deploy key")
	}

	switch strings.ToUpper(params.AppConfigUrl) {
	case string(ConfigTypeNone):
		return a.addAppWithNoConfigRepo(params, clusterName, secretRef)
	case string(ConfigTypeUserRepo):
		return a.addAppWithConfigInAppRepo(params, clusterName, secretRef)
	default:
		return a.addAppWithConfigInExternalRepo(params, clusterName, secretRef)
	}
}

func (a *App) updateParametersIfNecessary(params AddParams) (AddParams, error) {
	params.SourceType = string(SourceTypeGit)

	if params.Chart != "" {
		params.SourceType = string(SourceTypeHelm)
		params.DeploymentType = string(DeployTypeHelm)
		params.Name = params.Chart

		return params, nil
	}

	// Identifying repo url if not set by the user
	if params.Url == "" {
		url, err := a.getGitRemoteUrl(params)
		if err != nil {
			return params, err
		}

		params.Url = url
	} else {
		// making sure url is in the correct format
		params.Url = sanitizeRepoUrl(params.Url)

		// resetting Dir param since Url has priority over it
		params.Dir = ""
	}

	fmt.Printf("using URL: '%s' of origin from git config...\n\n", params.Url)

	if params.Name == "" {
		params.Name = generateResourceName(params.Url)
	}

	return params, nil
}

func (a *App) getGitRemoteUrl(params AddParams) (string, error) {
	repo, err := a.git.Open(params.Dir)
	if err != nil {
		return "", errors.Wrapf(err, "failed to open repository: %s", params.Dir)
	}

	remote, err := repo.Remote("origin")
	if err != nil {
		return "", errors.Wrapf(err, "failed to find the origin remote in the repository")
	}

	urls := remote.Config().URLs
	if len(urls) == 0 {
		return "", errors.Errorf("remote config in %s does not have an url", params.Dir)
	}

	return sanitizeRepoUrl(urls[0]), nil
}

func (a *App) addAppWithNoConfigRepo(params AddParams, clusterName string, secretRef string) error {
	appHash, err := utils.GetAppHash(params.Url, params.Path, params.Branch)
	if err != nil {
		return err
	}
	// if appHash exists as a label in the cluster we fail to create a PR
	if err := a.kube.LabelExistsInCluster(appHash); err != nil {
		return err
	}

	// Returns the source, app spec and kustomization
	source, appGoat, appSpec, err := a.generateAppManifests(params, params.Url, secretRef, clusterName)
	if err != nil {
		return errors.Wrap(err, "could not generate application GitOps Automation manifests")
	}

	fmt.Println("Applying manifests to the cluster...")
	return a.applyToCluster(params, source, appGoat, appSpec)
}

func (a *App) addAppWithConfigInAppRepo(params AddParams, clusterName string, secretRef string) error {
	appHash, err := utils.GetAppHash(params.Url, params.Path, params.Branch)
	if err != nil {
		return err
	}
	// if appHash exists as a label in the cluster we fail to create a PR
	if err := a.kube.LabelExistsInCluster(appHash); err != nil {
		return err
	}

	// Returns the source, app spec and kustomization
	source, appGoat, appSpec, err := a.generateAppManifests(params, params.Url, secretRef, clusterName)
	if err != nil {
		return errors.Wrap(err, "could not generate application GitOps Automation manifests")
	}

	// Kustomization pointing to the repo in .wego directory
	appWegoGoat, err := a.generateAppWegoManifests(params, clusterName)
	if err != nil {
		return errors.Wrap(err, "could not create GitOps automation for .wego directory")
	}

	// a local directory has not been passed, so we clone the repo passed in the --url
	if params.Dir == "" {
		fmt.Printf("Cloning %s...\n", params.Url)
		if err := a.cloneRepo(params.Url, params.Branch, params.DryRun); err != nil {
			return errors.Wrap(err, "failed to clone application repo")
		}
	}

	if !params.DryRun {
		if !params.AutoMerge {
			if err := a.createPullRequestToRepo(params, params.Url, appSpec, appGoat, clusterName); err != nil {
				return err
			}
		} else {
			fmt.Println("Writing manifests to disk...")

			if err := a.writeAppYaml(".wego", params.Name, appSpec); err != nil {
				return errors.Wrap(err, "failed writing app.yaml to disk")
			}

			if err := a.writeAppGoats(".wego", params.Name, clusterName, source, appGoat); err != nil {
				return errors.Wrap(err, "failed writing app.yaml to disk")
			}
		}
	}

	fmt.Println("Applying manifests to the cluster...")
	if err := a.applyToCluster(params, source, appWegoGoat); err != nil {
		return errors.Wrap(err, "could not apply manifests to the cluster")
	}

	return a.commitAndPush(params, func(fname string) bool {
		return strings.Contains(fname, ".wego")
	})
}

func (a *App) addAppWithConfigInExternalRepo(params AddParams, clusterName string, appSecretRef string) error {
	// making sure the url is in good format
	params.AppConfigUrl = sanitizeRepoUrl(params.AppConfigUrl)

	appHash, err := utils.GetAppHash(params.AppConfigUrl, params.Path, params.Branch)
	if err != nil {
		return err
	}
	// if appHash exists as a label in the cluster we fail to create a PR
	if err := a.kube.LabelExistsInCluster(appHash); err != nil {
		return err
	}

	appConfigSecretName, err := a.createAndUploadDeployKey(params.AppConfigUrl, SourceType(params.SourceType), clusterName, params.Namespace, params.DryRun)
	if err != nil {
		return errors.Wrap(err, "could not generate deploy key")
	}

	// Returns the source, app spec and kustomization
	appSource, appGoat, appSpec, err := a.generateAppManifests(params, params.AppConfigUrl, appSecretRef, clusterName)
	if err != nil {
		return errors.Wrap(err, "could not generate application GitOps Automation manifests")
	}

	targetSource, targetGoats, err := a.generateExternalRepoManifests(params, appConfigSecretName, clusterName)
	if err != nil {
		return errors.Wrap(err, "could not generate target GitOps Automation manifests")
	}

	if err := a.cloneRepo(params.AppConfigUrl, params.Branch, params.DryRun); err != nil {
		return errors.Wrap(err, "failed to clone application repo")
	}

	if !params.DryRun {
		if !params.AutoMerge {
			if err := a.createPullRequestToRepo(params, params.AppConfigUrl, appSpec, appGoat, clusterName); err != nil {
				return err
			}
		} else {
			fmt.Println("Writing manifests to disk...")

			if err := a.writeAppYaml(".", params.Name, appSpec); err != nil {
				return errors.Wrap(err, "failed writing app.yaml to disk")
			}

			if err := a.writeAppGoats(".", params.Name, clusterName, appSource, appGoat); err != nil {
				return errors.Wrap(err, "failed writing app.yaml to disk")
			}
		}
	}

	fmt.Println("Applying manifests to the cluster...")
	if err := a.applyToCluster(params, targetSource, targetGoats); err != nil {
		return errors.Wrapf(err, "could not apply manifests to the cluster")
	}

	return a.commitAndPush(params)
}

func (a *App) generateAppManifests(params AddParams, repo string, secretRef string, clusterName string) ([]byte, []byte, []byte, error) {
	var sourceManifest, appManifest, appGoatManifest []byte
	var err error
	fmt.Println("Generating Source manifest...")
	sourceManifest, err = a.generateSource(params, secretRef)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "could not set up GitOps for user repository")
	}

	fmt.Println("Generating GitOps automation manifests...")
	appGoatManifest, err = a.generateApplicationGoat(params)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, fmt.Sprintf("could not create GitOps automation for '%s'", params.Name))
	}

	fmt.Println("Generating Application spec manifest...")
	appManifest, err = generateAppYaml(params, repo)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, fmt.Sprintf("could not create app.yaml for '%s'", params.Name))
	}

	return sourceManifest, appGoatManifest, appManifest, nil
}

func (a *App) generateAppWegoManifests(params AddParams, clusterName string) ([]byte, error) {
	wegoPath := ".wego"

	appsDirManifest, err := a.flux.CreateKustomization(params.Name+"-wego-apps-dir", params.Name, filepath.Join(wegoPath, "apps", params.Name), params.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("could not create kustomization for '%s' .wego/apps", params.Name))
	}

	targetDirManifest, err := a.flux.CreateKustomization(fmt.Sprintf("%s-%s", clusterName, params.Name), params.Name, filepath.Join(wegoPath, "targets", clusterName), params.Namespace)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("could not create kustomization for '%s' .wego/apps", params.Name))
	}

	manifests := bytes.Join([][]byte{appsDirManifest, targetDirManifest}, []byte(""))

	return bytes.ReplaceAll(manifests, []byte("path: ./wego"), []byte("path: .wego")), nil
}

func (a *App) generateExternalRepoManifests(params AddParams, secretRef string, clusterName string) ([]byte, []byte, error) {
	repoName := generateResourceName(params.AppConfigUrl)

	targetSource, err := a.flux.CreateSourceGit(repoName, params.AppConfigUrl, params.Branch, secretRef, params.Namespace)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not generate target source manifests")
	}

	appGoat, err := a.flux.CreateKustomization(params.Name, repoName, filepath.Join(".", "apps", params.Name), params.Namespace)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not generate target goat manifests")
	}

	targetPath := filepath.Join(".", "targets", clusterName)
	targetGoat, err := a.flux.CreateKustomization(fmt.Sprintf("weave-gitops-%s", clusterName), repoName, targetPath, params.Namespace)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not generate target goat manifests")
	}

	manifests := bytes.Join([][]byte{targetGoat, appGoat}, []byte(""))

	return targetSource, manifests, nil
}

func (a *App) commitAndPush(params AddParams, filters ...func(string) bool) error {
	fmt.Println("Commiting and pushing wego resources for application...")
	if params.DryRun || !params.AutoMerge {
		return nil
	}

	_, err := a.git.Commit(git.Commit{
		Author:  git.Author{Name: "Weave Gitops", Email: "weave-gitops@weave.works"},
		Message: "Add App manifests",
	}, filters...)
	if err != nil && err != git.ErrNoStagedFiles {
		return fmt.Errorf("failed to commit sync manifests: %w", err)
	}

	if err == nil {
		fmt.Println("Pushing app manifests to repository...")
		if err = a.git.Push(context.Background()); err != nil {
			return fmt.Errorf("failed to push manifests: %w", err)
		}
	} else {
		fmt.Println("App manifests are up to date")
	}

	return nil
}

func (a *App) createAndUploadDeployKey(repoUrl string, sourceType SourceType, clusterName string, namespace string, dryRun bool) (string, error) {
	if SourceType(sourceType) == SourceTypeHelm {
		return "", nil
	}

	if repoUrl == "" {
		return "", nil
	}

	repoName := urlToRepoName(repoUrl)

	secretRefName := fmt.Sprintf("weave-gitops-%s-%s", clusterName, repoName)
	if dryRun {
		return secretRefName, nil
	}

	repoUrl = sanitizeRepoUrl(repoUrl)

	owner, err := getOwnerFromUrl(repoUrl)
	if err != nil {
		return "", err
	}

	deployKeyExists, err := a.gitProviders.DeployKeyExists(owner, repoName)
	if err != nil {
		return "", errors.Wrap(err, "could not check for existing deploy key")
	}

	secretPresent, err := a.kube.SecretPresent(context.Background(), secretRefName, namespace)
	if err != nil {
		return "", errors.Wrap(err, "could not check for existing secret")
	}

	if !deployKeyExists || !secretPresent {
		fmt.Printf("Generating deploy key for repo %s ...\n", repoUrl)
		deployKey, err := a.flux.CreateSecretGit(secretRefName, repoUrl, namespace)
		if err != nil {
			return "", errors.Wrap(err, "could not create git secret")
		}

		if err := a.gitProviders.UploadDeployKey(owner, repoName, deployKey); err != nil {
			return "", errors.Wrap(err, "error uploading deploy key")
		}
	}

	return secretRefName, nil
}

func (a *App) generateSource(params AddParams, secretRef string) ([]byte, error) {
	switch SourceType(params.SourceType) {
	case SourceTypeGit:
		sourceManifest, err := a.flux.CreateSourceGit(params.Name, params.Url, params.Branch, secretRef, params.Namespace)
		if err != nil {
			return nil, errors.Wrap(err, "could not create git source")
		}

		return sourceManifest, nil
	case SourceTypeHelm:
		return a.flux.CreateSourceHelm(params.Name, params.Url, params.Namespace)
	default:
		return nil, fmt.Errorf("unknown source type: %v", params.SourceType)
	}
}

func (a *App) generateApplicationGoat(params AddParams) ([]byte, error) {
	switch params.DeploymentType {
	case string(DeployTypeKustomize):
		return a.flux.CreateKustomization(params.Name, params.Name, params.Path, params.Namespace)
	case string(DeployTypeHelm):
		switch params.SourceType {
		case string(SourceTypeHelm):
			return a.flux.CreateHelmReleaseHelmRepository(params.Name, params.Chart, params.Namespace)
		case string(SourceTypeGit):
			return a.flux.CreateHelmReleaseGitRepository(params.Name, params.Name, params.Path, params.Namespace)
		default:
			return nil, fmt.Errorf("invalid source type: %v", params.SourceType)
		}
	default:
		return nil, fmt.Errorf("invalid deployment type: %v", params.DeploymentType)
	}
}

func (a *App) applyToCluster(params AddParams, manifests ...[]byte) error {
	if params.DryRun {
		for _, manifest := range manifests {
			fmt.Printf("%s\n", manifest)
		}
		return nil
	}

	for _, manifest := range manifests {
		if out, err := a.kube.Apply(manifest, params.Namespace); err != nil {
			return errors.Wrap(err, fmt.Sprintf("could not apply manifest: %s", string(out)))
		}
	}

	return nil
}

func (a *App) cloneRepo(url string, branch string, dryRun bool) error {
	if dryRun {
		return nil
	}

	url = sanitizeRepoUrl(url)

	repoDir, err := ioutil.TempDir("", "user-repo-")
	if err != nil {
		return errors.Wrap(err, "failed creating temp. directory to clone repo")
	}

	_, err = a.git.Clone(context.Background(), repoDir, url, branch)
	if err != nil {
		return errors.Wrapf(err, "failed cloning user repo: %s", url)
	}

	return nil
}

func (a *App) writeAppYaml(basePath string, name string, manifest []byte) error {
	manifestPath := filepath.Join(basePath, "apps", name, "app.yaml")

	return a.git.Write(manifestPath, manifest)
}

func (a *App) writeAppGoats(basePath string, name string, clusterName string, manifests ...[]byte) error {
	goatPath := filepath.Join(basePath, "targets", clusterName, name, fmt.Sprintf("%s-gitops-runtime.yaml", name))

	goat := bytes.Join(manifests, []byte(""))
	return a.git.Write(goatPath, goat)
}

func generateAppYaml(params AddParams, repo string) ([]byte, error) {
	const appYamlTemplate = `---
apiVersion: wego.weave.works/v1alpha1
kind: Application
metadata:
  name: {{ .AppName }}
  namespace: {{ .Namespace }}
  labels:
    weave-gitops.weave.works/app-identifier: {{ .AppHash }}
spec:
  path: {{ .AppPath }}
  url: {{ .AppURL }}
`
	// Create app.yaml
	t, err := template.New("appYaml").Parse(appYamlTemplate)
	if err != nil {
		return nil, errors.Wrap(err, "could not parse app yaml template")
	}

	appHash, err := utils.GetAppHash(repo, params.Path, params.Branch)
	if err != nil {
		return nil, err
	}

	var populated bytes.Buffer
	err = t.Execute(&populated, struct {
		AppName   string
		Namespace string
		AppHash   string
		AppPath   string
		AppURL    string
	}{params.Name, params.Namespace, appHash, params.Path, params.Url})
	if err != nil {
		return nil, errors.Wrap(err, "could not execute populated template")
	}
	return populated.Bytes(), nil
}

func generateResourceName(url string) string {
	return strings.ReplaceAll(urlToRepoName(url), "_", "-")
}

func getOwnerFromUrl(url string) (string, error) {
	parts := strings.Split(url, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("could not get owner from url %s", url)
	}
	return parts[len(parts)-2], nil
}

func urlToRepoName(url string) string {
	return strings.TrimSuffix(filepath.Base(url), ".git")
}

func sanitizeRepoUrl(url string) string {
	trimmed := ""

	if !strings.HasSuffix(url, ".git") {
		url = url + ".git"
	}

	sshPrefix := "git@github.com:"
	if strings.HasPrefix(url, sshPrefix) {
		trimmed = strings.TrimPrefix(url, sshPrefix)
	}

	httpsPrefix := "https://github.com/"
	if strings.HasPrefix(url, httpsPrefix) {
		trimmed = strings.TrimPrefix(url, httpsPrefix)
	}

	if trimmed != "" {
		return "ssh://git@github.com/" + trimmed
	}

	return url
}

func (a *App) createPullRequestToRepo(params AddParams, repo string, appYaml, applicationGoatYaml []byte, clusterName string) error {
	repoName := generateResourceName(repo)

	appPath := filepath.Join(".wego", "apps", params.Name, "app.yaml")

	goatPath := filepath.Join(".wego", "targets", clusterName, params.Name, fmt.Sprintf("%s-gitops-runtime.yaml", params.Name))
	if params.DryRun {
		fmt.Printf("Writing GitOps Automation to '%s'\n", goatPath)
		return nil
	}

	appcontent := string(appYaml)
	goatContent := string(applicationGoatYaml)
	files := []gitprovider.CommitFile{
		gitprovider.CommitFile{
			Path:    &appPath,
			Content: &appcontent,
		},
		gitprovider.CommitFile{
			Path:    &goatPath,
			Content: &goatContent,
		},
	}

	owner, err := getOwnerFromUrl(repo)
	if err != nil {
		return nil
	}

	accountType, err := a.gitProviders.GetAccountType(owner)
	if err != nil {
		return nil
	}

	appHash, err := utils.GetAppHash(repo, params.Path, params.Branch)
	if err != nil {
		return err
	}

	if accountType == gitproviders.AccountTypeOrg {
		org, err := fluxops.GetOwnerFromEnv()
		if err != nil {
			return nil
		}

		orgRepoRef := gitproviders.NewOrgRepositoryRef(github.DefaultDomain, org, repoName)
		return a.gitProviders.CreatePullRequestToOrgRepo(orgRepoRef, params.Branch, appHash, files, utils.GetCommitMessage(), fmt.Sprintf("wego add %s", params.Name), fmt.Sprintf("Added yamls for %s", params.Name))
	}

	userRepoRef := gitproviders.NewUserRepositoryRef(github.DefaultDomain, owner, repoName)
	return a.gitProviders.CreatePullRequestToUserRepo(userRepoRef, params.Branch, appHash, files, utils.GetCommitMessage(), fmt.Sprintf("wego add %s", params.Name), fmt.Sprintf("Added yamls for %s", params.Name))
}

// NOTE: ready to save the targets automation in phase 2
// func (a *App) writeTargetGoats(basePath string, name string, manifests ...[]byte) error {
//  goatPath := filepath.Join(basePath, "targets", fmt.Sprintf("%s-gitops-runtime.yaml", name))

//  goat := bytes.Join(manifests, []byte(""))
//  return a.git.Write(goatPath, goat)
// }
