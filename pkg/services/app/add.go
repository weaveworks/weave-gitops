package app

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/pkg/errors"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/utils"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

type DeploymentType string
type SourceType string
type ConfigType string

const (
	DeployTypeKustomize DeploymentType = "kustomize"
	DeployTypeHelm      DeploymentType = "helm"

	SourceTypeGit  SourceType = "git"
	SourceTypeHelm SourceType = "helm"

	ConfigTypeUserRepo        ConfigType = ""
	ConfigTypeNone            ConfigType = "NONE"
	WeGOAppIdentifierLabelKey            = "weave-gitops.weave.works/app-identifier"
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
	params, err := a.updateParametersIfNecessary(params)
	if err != nil {
		return errors.Wrap(err, "could not update parameters")
	}

	a.printAddSummary(params)

	a.logger.Waitingf("Checking cluster status")
	clusterStatus := a.kube.GetClusterStatus(ctx)
	a.logger.Successf(clusterStatus.String())

	switch clusterStatus {
	case kube.Unmodified:
		return errors.New("Wego not installed... exiting")
	case kube.Unknown:
		return errors.New("Wego can not determine cluster status... exiting")
	}

	clusterName, err := a.kube.GetClusterName(ctx)
	if err != nil {
		return err
	}

	secretRef, err := a.createAndUploadDeployKey(params.Url, SourceType(params.SourceType), clusterName, params.Namespace, params.DryRun)
	if err != nil {
		return errors.Wrap(err, "could not generate deploy key")
	}

	if err = a.setAppHashAndValidateIfExistsInCluster(ctx, params); err != nil {
		return err
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

func (a *App) setAppHashAndValidateIfExistsInCluster(ctx context.Context, params AddParams) error {

	appHash, err := utils.GetAppHash(params.Url, params.Path, params.Branch)
	if err != nil {
		return err
	}

	// if appHash exists as a label in the cluster we fail to create a PR
	if err = a.kube.LabelExistsInCluster(ctx, a.hash); err != nil {
		return err
	}

	a.hash = appHash

	return nil
}

func (a *App) printAddSummary(params AddParams) {
	a.logger.Println("Adding application:\n")
	a.logger.Println("Name: %s", params.Name)
	a.logger.Println("URL: %s", params.Url)
	a.logger.Println("Path: %s", params.Path)
	a.logger.Println("Branch: %s", params.Branch)
	a.logger.Println("Type: %s", params.DeploymentType)

	if params.Chart != "" {
		a.logger.Println("Chart: %s", params.Url)
	}

	a.logger.Println("")
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
	// Returns the source, app spec and kustomization
	source, appGoat, appSpec, err := a.generateAppManifests(params, params.Url, secretRef, clusterName)
	if err != nil {
		return errors.Wrap(err, "could not generate application GitOps Automation manifests")
	}

	a.logger.Actionf("Applying manifests to the cluster")
	return a.applyToCluster(params, source, appGoat, appSpec)
}

func (a *App) addAppWithConfigInAppRepo(params AddParams, clusterName string, secretRef string) error {

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
		a.logger.Actionf("Cloning %s", params.Url)
		if err := a.cloneRepo(params.Url, params.Branch, params.DryRun); err != nil {
			return errors.Wrap(err, "failed to clone application repo")
		}
	}

	if !params.DryRun {
		if !params.AutoMerge {
			if err := a.createPullRequestToRepo(params, ".wego", params.Url, clusterName, appSpec, appGoat); err != nil {
				return err
			}
		} else {
			a.logger.Actionf("Writing manifests to disk")

			if err := a.writeAppYaml(".wego", params.Name, appSpec); err != nil {
				return errors.Wrap(err, "failed writing app.yaml to disk")
			}

			if err := a.writeAppGoats(".wego", params.Name, clusterName, source, appGoat); err != nil {
				return errors.Wrap(err, "failed writing app.yaml to disk")
			}
		}
	}

	a.logger.Actionf("Applying manifests to the cluster")
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

	appConfigSecretName, err := a.createAndUploadDeployKey(params.AppConfigUrl, SourceTypeGit, clusterName, params.Namespace, params.DryRun)
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
			if err := a.createPullRequestToRepo(params, ".", params.AppConfigUrl, clusterName, appSpec, appGoat, appSource); err != nil {
				return err
			}
		} else {
			a.logger.Actionf("Writing manifests to disk")

			if err := a.writeAppYaml(".", params.Name, appSpec); err != nil {
				return errors.Wrap(err, "failed writing app.yaml to disk")
			}

			if err := a.writeAppGoats(".", params.Name, clusterName, appSource, appGoat); err != nil {
				return errors.Wrap(err, "failed writing app.yaml to disk")
			}
		}
	}

	a.logger.Actionf("Applying manifests to the cluster")
	if err := a.applyToCluster(params, targetSource, targetGoats); err != nil {
		return errors.Wrapf(err, "could not apply manifests to the cluster")
	}

	return a.commitAndPush(params)
}

func (a *App) generateAppManifests(params AddParams, repo string, secretRef string, clusterName string) ([]byte, []byte, []byte, error) {
	var sourceManifest, appManifest, appGoatManifest []byte
	var err error
	a.logger.Generatef("Generating Source manifest")
	sourceManifest, err = a.generateSource(params, secretRef)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, "could not set up GitOps for user repository")
	}

	a.logger.Generatef("Generating GitOps automation manifests")
	appGoatManifest, err = a.generateApplicationGoat(params)
	if err != nil {
		return nil, nil, nil, errors.Wrap(err, fmt.Sprintf("could not create GitOps automation for '%s'", params.Name))
	}

	a.logger.Generatef("Generating Application spec manifest")
	appManifest, err = a.generateAppYaml(params)
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
	if params.DryRun || !params.AutoMerge {
		return nil
	}
	a.logger.Actionf("Committing and pushing wego resources for application")

	_, err := a.git.Commit(git.Commit{
		Author:  git.Author{Name: "Weave Gitops", Email: "weave-gitops@weave.works"},
		Message: "Add App manifests",
	}, filters...)
	if err != nil && err != git.ErrNoStagedFiles {
		return fmt.Errorf("failed to commit sync manifests: %w", err)
	}

	if err == nil {
		a.logger.Actionf("Pushing app manifests to repository")
		if err = a.git.Push(context.Background()); err != nil {
			return fmt.Errorf("failed to push manifests: %w", err)
		}
	} else {
		a.logger.Successf("App manifests are up to date")
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
		a.logger.Generatef("Generating deploy key for repo %s", repoUrl)
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

func makeWegoApplication(params AddParams) wego.Application {
	gvk := wego.GroupVersion.WithKind(wego.ApplicationKind)
	app := wego.Application{
		TypeMeta: metav1.TypeMeta{
			Kind:       gvk.Kind,
			APIVersion: gvk.GroupVersion().String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      params.Name,
			Namespace: params.Namespace,
		},
		Spec: wego.ApplicationSpec{
			URL:  params.Url,
			Path: params.Path,
		},
	}

	return app
}

func (a *App) generateAppYaml(params AddParams) ([]byte, error) {
	app := makeWegoApplication(params)

	app.ObjectMeta.Labels = map[string]string{
		WeGOAppIdentifierLabelKey: a.hash,
	}

	b, err := yaml.Marshal(&app)
	if err != nil {
		return nil, fmt.Errorf("could not marshal yaml: %w", err)
	}

	return sanitizeK8sYaml(b), nil
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

func (a *App) createPullRequestToRepo(params AddParams, basePath string, repo string, clusterName string, appYaml []byte, goatManifests ...[]byte) error {
	repoName := generateResourceName(repo)

	appPath := filepath.Join(basePath, "apps", params.Name, "app.yaml")

	goatPath := filepath.Join(basePath, "targets", clusterName, params.Name, fmt.Sprintf("%s-gitops-runtime.yaml", params.Name))

	goat := bytes.Join(goatManifests, []byte(""))

	if params.DryRun {
		fmt.Printf("Writing GitOps Automation to '%s'\n", goatPath)
		return nil
	}

	appcontent := string(appYaml)
	goatContent := string(goat)
	files := []gitprovider.CommitFile{
		{
			Path:    &appPath,
			Content: &appcontent,
		},
		{
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

	if accountType == gitproviders.AccountTypeOrg {
		org, err := fluxops.GetOwnerFromEnv()
		if err != nil {
			return nil
		}

		orgRepoRef := gitproviders.NewOrgRepositoryRef(github.DefaultDomain, org, repoName)
		prLink, err := a.gitProviders.CreatePullRequestToOrgRepo(orgRepoRef, params.Branch, a.hash, files, utils.GetCommitMessage(), fmt.Sprintf("wego add %s", params.Name), fmt.Sprintf("Added yamls for %s", params.Name))
		if err != nil {
			return fmt.Errorf("unable to create pull request: %s", err)
		}
		a.logger.Println("Pull Request created: %s\n", prLink.Get().WebURL)
		return nil
	}

	userRepoRef := gitproviders.NewUserRepositoryRef(github.DefaultDomain, owner, repoName)
	prLink, err := a.gitProviders.CreatePullRequestToUserRepo(userRepoRef, params.Branch, a.hash, files, utils.GetCommitMessage(), fmt.Sprintf("wego add %s", params.Name), fmt.Sprintf("Added yamls for %s", params.Name))
	if err != nil {
		return fmt.Errorf("unable to create pull request: %s", err)
	}
	a.logger.Println("Pull Request created: %s\n", prLink.Get().WebURL)
	return nil
}

// NOTE: ready to save the targets automation in phase 2
// func (a *App) writeTargetGoats(basePath string, name string, manifests ...[]byte) error {
//  goatPath := filepath.Join(basePath, "targets", fmt.Sprintf("%s-gitops-runtime.yaml", name))

//  goat := bytes.Join(manifests, []byte(""))
//  return a.git.Write(goatPath, goat)
// }

// Remove some problematic fields before saving the yaml files.
// K8s/reconcilers will populate these fields after creation.
// https://github.com/fluxcd/flux2/blob/0ae39d5a0a5220c177b29e71fc8824babd1e0d7c/cmd/flux/export.go#L111
func sanitizeK8sYaml(data []byte) []byte {
	out := []byte("---\n")
	data = bytes.Replace(data, []byte("  creationTimestamp: null\n"), []byte(""), 1)
	data = bytes.Replace(data, []byte("status: {}\n"), []byte(""), 1)
	return append(out, data...)
}
