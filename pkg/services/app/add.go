package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"github.com/weaveworks/weave-gitops/pkg/utils"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

const (
	WeGOAppIdentifierLabelKey = "wego.weave.works/app-identifier"
)

type AddParams struct {
	Dir                        string
	Name                       string
	Url                        string
	Path                       string
	Branch                     string
	DeploymentType             string
	Chart                      string
	SourceType                 wego.SourceType
	AppConfigUrl               string
	Namespace                  string
	DryRun                     bool
	AutoMerge                  bool
	GitProviderToken           string
	HelmReleaseTargetNamespace string
	MigrateToNewDirStructure   func(string) string
}

func (a AddParams) IsHelmRepository() bool {
	return a.Chart != ""
}

var defaultMigrateToNewDirStructure func(string) string = func(s string) string { return s }

const (
	DefaultPath           = "./"
	DefaultBranch         = "main"
	DefaultDeploymentType = "kustomize"
	AddCommitMessage      = "Add App manifests"
)

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

func (a *App) Add(configGit git.Git, gitProvider gitproviders.GitProvider, params AddParams) error {
	ctx := context.Background()

	params, err := a.updateParametersIfNecessary(ctx, gitProvider, params)
	if err != nil {
		return fmt.Errorf("could not update parameters: %w", err)
	}

	a.printAddSummary(params)

	if err := kube.IsClusterReady(a.Logger, a.Kube); err != nil {
		return err
	}

	clusterName, err := a.Kube.GetClusterName(ctx)
	if err != nil {
		return err
	}

	app, err := models.NewApplication(makeWegoApplication(params))
	if err != nil {
		return err
	}

	appHash := app.GetAppHash()

	wegoapps, err := a.Kube.GetApplications(ctx, params.Namespace)
	if err != nil {
		return err
	}

	for _, wegoapp := range wegoapps {
		app, err := models.NewApplication(wegoapp)
		if err != nil {
			return err
		}

		if appHash == app.GetAppHash() {
			return fmt.Errorf("unable to create resource, resource already exists in cluster")
		}
	}

	// secretRef := ""

	// if params.SourceType != wego.SourceTypeHelm {
	//  visibility, visibilityErr := a.GitProvider.GetRepoVisibility(ctx, app.AppRepoUrl)
	//  if visibilityErr != nil {
	//      return visibilityErr
	//  }

	//  if *visibility != gitprovider.RepositoryVisibilityPublic {
	//      secretRef = app.RepoSecretName(app.AppRepoUrl).String()
	//  }
	// }

	switch strings.ToUpper(info.Spec.ConfigURL) {
	case string(models.ConfigTypeNone):
		return a.addAppWithNoConfigRepo(info, params.DryRun, secretRef, appHash)
	case string(models.ConfigTypeUserRepo):
		return a.addAppWithConfigInAppRepo(ctx, configGit, gitProvider, info, params, secretRef, appHash)
	default:
		return a.addAppWithConfigInExternalRepo(ctx, configGit, gitProvider, info, params, secretRef, appHash)
	}
}

func (a *AppSvc) printAddSummary(params AddParams) {
	a.Logger.Println("Adding application:\n")
	a.Logger.Println("Name: %s", params.Name)
	a.Logger.Println("URL: %s", params.Url)
	a.Logger.Println("Path: %s", params.Path)
	a.Logger.Println("Branch: %s", params.Branch)
	a.Logger.Println("Type: %s", params.DeploymentType)

	if params.Chart != "" {
		a.Logger.Println("Chart: %s", params.Chart)
	}

	a.Logger.Println("")
}

func (a *AppSvc) updateParametersIfNecessary(ctx context.Context, gitProvider gitproviders.GitProvider, params AddParams) (AddParams, error) {
	params.SourceType = wego.SourceTypeGit

	var appRepoUrl gitproviders.RepoURL

	switch {
	case params.Chart != "":
		params.SourceType = wego.SourceTypeHelm
		params.DeploymentType = string(wego.DeploymentTypeHelm)
		params.Path = params.Chart

		if params.Name == "" {
			if models.ApplicationNameTooLong(params.Chart) {
				return params, fmt.Errorf("chart name %q is too long to use as application name; please specify name with '--name'", params.Chart)
			}

			params.Name = params.Chart
		}

		if params.Url == "" {
			return params, fmt.Errorf("--url must be specified for helm repositories")
		}

		if params.AppConfigUrl == string(models.ConfigTypeUserRepo) {
			return params, errors.New("--app-config-url should be provided or set to NONE")
		}
	default:
		var err error

		appRepoUrl, err = gitproviders.NewRepoURL(params.Url)
		if err != nil {
			return params, fmt.Errorf("error normalizing url: %w", err)
		}

		params.Url = appRepoUrl.String()

		// resetting Dir param since Url has priority over it
		params.Dir = ""
	}

	// making sure the config url is in good format
	if models.IsExternalConfigUrl(params.AppConfigUrl) {
		configRepoUrl, err := gitproviders.NewRepoURL(params.AppConfigUrl)
		if err != nil {
			return params, fmt.Errorf("error normalizing url: %w", err)
		}

		params.AppConfigUrl = configRepoUrl.String()
	}

	if params.Name == "" {
		repoName := utils.UrlToRepoName(params.Url)
		if models.ApplicationNameTooLong(repoName) {
			return params, fmt.Errorf("url base name %q is too long to use as application name; please specify name with '--name'", repoName)
		}

		params.Name = models.GenerateResourceName(appRepoUrl)
	}

	if params.Path == "" {
		params.Path = DefaultPath
	}

	if params.DeploymentType == "" {
		params.DeploymentType = DefaultDeploymentType
	}

	if params.Branch == "" {
		params.Branch = DefaultBranch

		if params.SourceType == wego.SourceTypeGit {
			branch, err := gitProvider.GetDefaultBranch(ctx, appRepoUrl)
			if err != nil {
				return params, err
			} else {
				params.Branch = branch
			}
		}
	}

	if models.ApplicationNameTooLong(params.Name) {
		return params, fmt.Errorf("application name too long: %s; must be <= 63 characters", params.Name)
	}

	// Validate namespace argument for helm
	if params.HelmReleaseTargetNamespace != "" {
		if ok, _ := a.Kube.NamespacePresent(context.Background(), params.HelmReleaseTargetNamespace); !ok {
			return params, fmt.Errorf("Helm Release Target Namespace %s does not exist", params.HelmReleaseTargetNamespace)
		}

		if nserr := utils.ValidateNamespace(params.HelmReleaseTargetNamespace); nserr != nil {
			return params, nserr
		}
	}

	return params, nil
}

func (a *AppSvc) addAppWithNoConfigRepo(app models.Application, dryRun bool, clusterName string) error {
	manifests, err := a.Automation.GenerateManifests(context.Background(), app, clusterName)
	if err != nil {
		return fmt.Errorf("could not generate application GitOps Automation manifests: %w", err)
	}

	a.Logger.Actionf("Applying manifests to the cluster")

	return a.applyToCluster(app, dryRun, manifests)
}

func (a *AppSvc) addAppWithConfigInAppRepo(ctx context.Context, configGit git.Git, gitProvider gitproviders.GitProvider, info *AppResourceInfo, params AddParams, secretRef string, appHash string) error {
	// Returns the source, app spec and kustomization
	source, appGoat, appSpec, err := a.generateAppManifests(info, secretRef, appHash)
	if err != nil {
		return fmt.Errorf("could not generate application GitOps Automation manifests: %w", err)
	}

	// a local directory has not been passed, so we clone the repo passed in the --url
	if params.Dir == "" {
		a.Logger.Actionf("Cloning %s", app.Spec.URL)

		remover, _, err := a.cloneRepo(a.ConfigGit, app.Spec.URL, app.Spec.Branch, params.DryRun)
		if err != nil {
			return fmt.Errorf("failed to clone application repo: %w", err)
		}

		defer remover()
	}

	if !params.DryRun {
		if !params.AutoMerge {
			if err := a.createPullRequestToRepo(ctx, gitProvider, info, info.appRepoUrl, appHash, appSpec, source, appGoat); err != nil {
				return err
			}
		} else {
			a.Logger.Actionf("Writing manifests to disk")

			if err := a.writeYaml(app, manifests, params.MigrateToNewDirStructure); err != nil {
				return fmt.Errorf("failed writing app.yaml to disk: %w", err)
			}
		}
	}

	a.Logger.Actionf("Applying manifests to the cluster")

	if err := a.applyToCluster(app, params.DryRun, manifests); err != nil {
		return fmt.Errorf("could not apply manifests to the cluster: %w", err)
	}

	return a.commitAndPush(configGit, AddCommitMessage, params.DryRun, func(fname string) bool {
		return strings.Contains(fname, ".wego")
	})
}

func (a *AppSvc) addAppWithConfigInExternalRepo(ctx context.Context, configGit git.Git, gitProvider gitproviders.GitProvider, info *AppResourceInfo, params AddParams, appSecretRef string, appHash string) error {
	// Returns the source, app spec and kustomization
	appSource, appGoat, appSpec, err := a.generateAppManifests(info, appSecretRef, appHash)
	if err != nil {
		return fmt.Errorf("could not generate application GitOps Automation manifests: %w", err)
	}

	configUrl, err := app.GetConfigUrl()
	if err != nil {
		return err
	}

	configBranch, err := a.GitProvider.GetDefaultBranch(ctx, configUrl)
	if err != nil {
		return fmt.Errorf("could not determine default branch for config repository: %w", err)
	}

	remover, repoAbsPath, err := a.cloneRepo(a.ConfigGit, configUrl.String(), configBranch, params.DryRun)
	if err != nil {
		return fmt.Errorf("failed to clone configuration repo: %w", err)
	}

	defer remover()

	if !params.DryRun {
		if !params.AutoMerge {
			if err := a.createPullRequestToRepo(ctx, gitProvider, app, app.ConfigRepoUrl, manifests); err != nil {
				return err
			}
		} else {
			a.Logger.Actionf("Writing manifests to disk")

			if err := a.writeYaml(configGit, app, manifests, params.MigrateToNewDirStructure); err != nil {
				return fmt.Errorf("failed writing app.yaml to disk: %w", err)
			}
		}
	}

	a.Logger.Actionf("Applying manifests to the cluster")
	// if params.MigrateToNewDirStructure is defined we skip applying to the cluster
	if params.MigrateToNewDirStructure != nil {
		appKustomization := filepath.Join(git.WegoRoot, git.WegoAppDir, app.AppDeployName(), "kustomization.yaml")
		k, err := a.createOrUpdateKustomize(app, params, app.AppDeployName(), []string{filepath.Base(app.AppYamlPath()),
			filepath.Base(app.AppAutomationSourcePath(clusterName)), filepath.Base(app.AppAutomationDeployPath(clusterName))},
			filepath.Join(repoAbsPath, appKustomization))

		if err != nil {
			return fmt.Errorf("failed to create app kustomization: %w", err)
		}

		if err := configGit.Write(appKustomization, k); err != nil {
			return fmt.Errorf("failed writing app kustomization.yaml to disk: %w", err)
		}
		// TODO move to a deploy or apply command.
		userKustomization := filepath.Join(git.WegoRoot, git.WegoClusterDir, app.ClusterName, "user", "kustomization.yaml")

		uk, err := a.createOrUpdateKustomize(app, params, app.AppDeployName(), []string{"../../../apps/" + app.AppDeployName()},
			filepath.Join(repoAbsPath, userKustomization))
		if err != nil {
			return fmt.Errorf("failed to create app kustomization: %w", err)
		}
		//TODO - need to read the existing kustomization instead of writing it here
		if err := configGit.Write(userKustomization, uk); err != nil {
			return fmt.Errorf("failed writing cluster kustomization.yaml to disk: %w", err)
		}
	} else {
		if err := a.applyToCluster(app, params.DryRun, manifests); err != nil {
			return fmt.Errorf("could not apply manifests to the cluster: %w", err)
		}
	}

	return a.commitAndPush(configGit, AddCommitMessage, params.DryRun)
}

func (a *AppSvc) createOrUpdateKustomize(app models.Application, params AddParams, name string, resources []string, kustomizeFile string) ([]byte, error) {
	k := types.Kustomization{}

	contents, err := os.ReadFile(kustomizeFile)
	if err == nil {
		if err := yaml.Unmarshal(contents, &k); err != nil {
			return nil, fmt.Errorf("failed to read existing kustomize file %s %w", kustomizeFile, err)
		}
	} else {
		k.MetaData = &types.ObjectMeta{
			Name:      name,
			Namespace: app.Namespace,
		}
		k.APIVersion = types.KustomizationVersion
		k.Kind = types.KustomizationKind
	}

	k.Resources = append(k.Resources, resources...)

	kustomize, err := yaml.Marshal(&k)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal kustomize %v : %w", k, err)
	}

	return kustomize, nil
}

func (a *AppSvc) commitAndPush(client git.Git, commitMsg string, dryRun bool, filters ...func(string) bool) error {
	a.Logger.Actionf("Committing and pushing gitops updates for application")
	return CommitAndPush(client, commitMsg, dryRun, a.Logger, filters...)
}
func CommitAndPush(client git.Git, commitMsg string, dryRun bool, logger logger.Logger, filters ...func(string) bool) error {
	logger.Actionf("Committing and pushing gitops updates for application")

	if dryRun {
		return nil
	}

	_, err := client.Commit(git.Commit{
		Author:  git.Author{Name: "Weave Gitops", Email: "weave-gitops@weave.works"},
		Message: commitMsg,
	}, filters...)
	if err != nil && err != git.ErrNoStagedFiles {
		return fmt.Errorf("failed to update the repository: %w", err)
	}

	if err == nil {
		logger.Actionf("Pushing app changes to repository")

		if err = client.Push(context.Background()); err != nil {
			return fmt.Errorf("failed to push changes: %w", err)
		}
	} else {
		logger.Successf("App is up to date")
	}

	return nil
}

func (a *AppSvc) applyToCluster(app models.Application, dryRun bool, manifests []automation.AutomationManifest) error {
	if dryRun {
		for _, manifest := range manifests {
			fmt.Fprintf(a.Osys.Stdout(), "%s\n", manifest.Manifest)
		}

		return nil
	}

	for _, manifest := range manifests {
		if manifest.Path != "" {
			continue
		}

		if err := a.Kube.Apply(context.Background(), manifest.Manifest, app.Namespace); err != nil {
			return fmt.Errorf("could not apply manifest: %w", err)
		}
	}

	return nil
}

func (a *AppSvc) cloneRepo(client git.Git, url string, branch string, dryRun bool) (func(), string, error) {
	return CloneRepo(client, url, branch, dryRun)
}

// CloneRepo uses the git client to clone the reop from the URL and branch.  It clones into a temp
// directory and returns a function to use by the caller for cleanup.  The temp directory is
// also returned.
func CloneRepo(client git.Git, url string, branch string, dryRun bool) (func(), string, error) {
	if dryRun {
		return func() {}, "", nil
	}

	repoDir, err := ioutil.TempDir("", "user-repo-")
	if err != nil {
		return nil, "", fmt.Errorf("failed creating temp. directory to clone repo: %w", err)
	}

	_, err = client.Clone(context.Background(), repoDir, url, branch)
	if err != nil {
		return nil, "", fmt.Errorf("failed cloning user repo: %s: %w", url, err)
	}

	return func() {
		os.RemoveAll(repoDir)
	}, repoDir, nil
}

func (a *AppSvc) writeYaml(app models.Application, manifests []automation.AutomationManifest, overridePath func(string) string) error {
	if overridePath != nil {
		for _, manifest := range manifests {
			if manifest.Path == "" {
				continue
			}

			if err := a.ConfigGit.Write(overridePath(manifest.Path), manifest.Manifest); err != nil {
				return err
			}
		}
	}

	for _, manifest := range manifests {
		if manifest.Path == "" {
			continue
		}

		if err := a.ConfigGit.Write(defaultMigrateToNewDirStructure(manifest.Path), manifest.Manifest); err != nil {
			return err
		}
	}

	return nil
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
			ConfigURL:           params.AppConfigUrl,
			Branch:              params.Branch,
			URL:                 params.Url,
			Path:                params.Path,
			DeploymentType:      wego.DeploymentType(params.DeploymentType),
			SourceType:          wego.SourceType(params.SourceType),
			HelmTargetNamespace: params.HelmReleaseTargetNamespace,
		},
	}

	return app
}

func (a *AppSvc) createPullRequestToRepo(ctx context.Context, app models.Application, repoUrl gitproviders.RepoURL, manifests []automation.AutomationManifest) error {
	var files []gitprovider.CommitFile

	for _, manifest := range manifests {
		if manifest.Path == "" {
			continue
		}

		content := string(manifest.Manifest)
		files = append(files, gitprovider.CommitFile{Path: &manifest.Path, Content: &content})
	}

	defaultBranch, err := gitProvider.GetDefaultBranch(ctx, repoUrl)
	if err != nil {
		return err
	}

	prApp := gitproviders.PullRequestInfo{
		Title:         fmt.Sprintf("Gitops add %s", app.Name),
		Description:   fmt.Sprintf("Added yamls for %s", app.Name),
		CommitMessage: "Add App Manifests",
		TargetBranch:  defaultBranch,
		NewBranch:     app.GetAppHash(),
		Files:         files,
	}

	pr, err := a.GitProvider.CreatePullRequest(ctx, repoUrl, prApp)
	if err != nil {
		return fmt.Errorf("unable to create pull request: %w", err)
	}

	a.Logger.Println("Pull Request created: %s\n", pr.Get().WebURL)

	return nil
}

// NOTE: ready to save the targets automation in phase 2
// func (a *AppSvc) writeTargetGoats(basePath string, name string, manifests ...[]byte) error {
//  goatPath := filepath.Join(basePath, "targets", fmt.Sprintf("%s-gitops-runtime.yaml", name))

//  goat := bytes.Join(manifests, []byte(""))
//  return a.Git.Write(goatPath, goat)
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

func sanitizeWegoDirectory(manifest []byte) []byte {
	return bytes.ReplaceAll(manifest, []byte("path: ./wego"), []byte("path: .wego"))
}
