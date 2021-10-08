package app

import (
	"bytes"
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/utils"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/kustomize/api/types"
	"sigs.k8s.io/yaml"
)

type ConfigType string

type ConfigMode string

type ResourceKind string

type ResourceRef struct {
	kind           ResourceKind
	name           string
	repositoryPath string
}

type AppResourceInfo struct {
	wego.Application
	clusterName   string
	targetName    string
	appRepoUrl    gitproviders.RepoURL
	configRepoUrl gitproviders.RepoURL
}

const (
	ConfigTypeUserRepo ConfigType = ""
	ConfigTypeNone     ConfigType = "NONE"

	ConfigModeClusterOnly  ConfigMode = "clusterOnly"
	ConfigModeUserRepo     ConfigMode = "userRepo"
	ConfigModeExternalRepo ConfigMode = "externalRepo"

	ResourceKindApplication    ResourceKind = "Application"
	ResourceKindSecret         ResourceKind = "Secret"
	ResourceKindGitRepository  ResourceKind = "GitRepository"
	ResourceKindHelmRepository ResourceKind = "HelmRepository"
	ResourceKindKustomization  ResourceKind = "Kustomization"
	ResourceKindHelmRelease    ResourceKind = "HelmRelease"

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

type externalRepoManifests struct {
	source []byte
	target []byte
	appDir []byte
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

	info, err := getAppResourceInfo(makeWegoApplication(params), clusterName)
	if err != nil {
		return err
	}

	appHash := info.getAppHash()

	apps, err := a.Kube.GetApplications(ctx, params.Namespace)
	if err != nil {
		return err
	}

	for _, app := range apps {
		info, err := getAppResourceInfo(app, clusterName)
		if err != nil {
			return err
		}

		if appHash == info.getAppHash() {
			return fmt.Errorf("unable to create resource, resource already exists in cluster")
		}
	}

	secretRef := ""

	if params.SourceType != wego.SourceTypeHelm {
		visibility, visibilityErr := gitProvider.GetRepoVisibility(ctx, info.appRepoUrl)
		if visibilityErr != nil {
			return visibilityErr
		}

		if *visibility != gitprovider.RepositoryVisibilityPublic {
			secretRef = info.repoSecretName(info.appRepoUrl.String()).String()
		}
	}

	switch strings.ToUpper(info.Spec.ConfigURL) {
	case string(ConfigTypeNone):
		return a.addAppWithNoConfigRepo(info, params.DryRun, secretRef, appHash)
	case string(ConfigTypeUserRepo):
		return a.addAppWithConfigInAppRepo(ctx, configGit, gitProvider, info, params, secretRef, appHash)
	default:
		return a.addAppWithConfigInExternalRepo(ctx, configGit, gitProvider, info, params, secretRef, appHash)
	}
}

func (a *App) printAddSummary(params AddParams) {
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

const maxKubernetesResourceNameLength = 63

func IsExternalConfigUrl(url string) bool {
	return strings.ToUpper(url) != string(ConfigTypeNone) &&
		strings.ToUpper(url) != string(ConfigTypeUserRepo)
}

func (a *App) updateParametersIfNecessary(ctx context.Context, gitProvider gitproviders.GitProvider, params AddParams) (AddParams, error) {
	params.SourceType = wego.SourceTypeGit

	var appRepoUrl gitproviders.RepoURL

	switch {
	case params.Chart != "":
		params.SourceType = wego.SourceTypeHelm
		params.DeploymentType = string(wego.DeploymentTypeHelm)
		params.Path = params.Chart

		if params.Name == "" {
			if nameTooLong(params.Chart) {
				return params, fmt.Errorf("chart name %q is too long to use as application name; please specify name with '--name'", params.Chart)
			}

			params.Name = params.Chart
		}

		if params.Url == "" {
			return params, fmt.Errorf("--url must be specified for helm repositories")
		}

		if params.AppConfigUrl == string(ConfigTypeUserRepo) {
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
	if IsExternalConfigUrl(params.AppConfigUrl) {
		configRepoUrl, err := gitproviders.NewRepoURL(params.AppConfigUrl)
		if err != nil {
			return params, fmt.Errorf("error normalizing url: %w", err)
		}

		params.AppConfigUrl = configRepoUrl.String()
	}

	if params.Name == "" {
		repoName := utils.UrlToRepoName(params.Url)
		if nameTooLong(repoName) {
			return params, fmt.Errorf("url base name %q is too long to use as application name; please specify name with '--name'", repoName)
		}

		params.Name = generateResourceName(params.Url)
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

	if nameTooLong(params.Name) {
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

func (a *App) addAppWithNoConfigRepo(info *AppResourceInfo, dryRun bool, secretRef string, appHash string) error {
	// Returns the source, app spec and kustomization
	source, appGoat, appSpec, err := a.generateAppManifests(info, secretRef, appHash)
	if err != nil {
		return fmt.Errorf("could not generate application GitOps Automation manifests: %w", err)
	}

	a.Logger.Actionf("Applying manifests to the cluster")

	return a.applyToCluster(info, dryRun, source.ToSourceYAML(), appGoat.ToAutomationYAML(), appSpec.ToAppYAML())
}

func (a *App) addAppWithConfigInAppRepo(ctx context.Context, configGit git.Git, gitProvider gitproviders.GitProvider, info *AppResourceInfo, params AddParams, secretRef string, appHash string) error {
	// Returns the source, app spec and kustomization
	source, appGoat, appSpec, err := a.generateAppManifests(info, secretRef, appHash)
	if err != nil {
		return fmt.Errorf("could not generate application GitOps Automation manifests: %w", err)
	}

	appDirGoat, targetDirGoat, err := a.generateAppWegoManifests(info)
	if err != nil {
		return fmt.Errorf("could not create GitOps automation for .wego directory: %w", err)
	}

	// a local directory has not been passed, so we clone the repo passed in the --url
	if params.Dir == "" {
		a.Logger.Actionf("Cloning %s", info.Spec.URL)

		remover, _, err := a.cloneRepo(configGit, info.Spec.URL, info.Spec.Branch, params.DryRun)
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

			if err := a.writeAppYaml(configGit, info, appSpec, params.MigrateToNewDirStructure); err != nil {
				return fmt.Errorf("failed writing app.yaml to disk: %w", err)
			}

			if err := a.writeAppGoats(configGit, info, source, appGoat, params.MigrateToNewDirStructure); err != nil {
				return fmt.Errorf("failed writing app.yaml to disk: %w", err)
			}
		}
	}

	a.Logger.Actionf("Applying manifests to the cluster")

	if err := a.applyToCluster(info, params.DryRun, source.ToSourceYAML(), appDirGoat, targetDirGoat); err != nil {
		return fmt.Errorf("could not apply manifests to the cluster: %w", err)
	}

	return a.commitAndPush(configGit, AddCommitMessage, params.DryRun, func(fname string) bool {
		return strings.Contains(fname, ".wego")
	})
}

func (a *App) addAppWithConfigInExternalRepo(ctx context.Context, configGit git.Git, gitProvider gitproviders.GitProvider, info *AppResourceInfo, params AddParams, appSecretRef string, appHash string) error {
	// Returns the source, app spec and kustomization
	appSource, appGoat, appSpec, err := a.generateAppManifests(info, appSecretRef, appHash)
	if err != nil {
		return fmt.Errorf("could not generate application GitOps Automation manifests: %w", err)
	}

	configBranch, err := gitProvider.GetDefaultBranch(context.Background(), info.configRepoUrl)
	if err != nil {
		return fmt.Errorf("could not determine default branch for config repository: %w", err)
	}

	extRepoMan, err := a.generateExternalRepoManifests(gitProvider, info, configBranch)
	if err != nil {
		return fmt.Errorf("could not generate target GitOps Automation manifests: %w", err)
	}

	remover, repoAbsPath, err := a.cloneRepo(configGit, info.Spec.ConfigURL, configBranch, params.DryRun)
	if err != nil {
		return fmt.Errorf("failed to clone configuration repo: %w", err)
	}

	defer remover()

	if !params.DryRun {
		if !params.AutoMerge {
			if err := a.createPullRequestToRepo(ctx, gitProvider, info, info.configRepoUrl, appHash, appSpec, appSource, appGoat); err != nil {
				return err
			}
		} else {
			a.Logger.Actionf("Writing manifests to disk")

			if err := a.writeAppYaml(configGit, info, appSpec, params.MigrateToNewDirStructure); err != nil {
				return fmt.Errorf("failed writing app.yaml to disk: %w", err)
			}
			if err := a.writeAppGoats(configGit, info, appSource, appGoat, params.MigrateToNewDirStructure); err != nil {
				return fmt.Errorf("failed writing application gitops manifests to disk: %w", err)
			}
		}
	}

	a.Logger.Actionf("Applying manifests to the cluster")
	// if params.MigrateToNewDirStructure is defined we skip applying to the cluster
	if params.MigrateToNewDirStructure != nil {
		appKustomization := filepath.Join(git.WegoRoot, git.WegoAppDir, info.appDeployName(), "kustomization.yaml")
		k, err := a.createOrUpdateKustomize(info, params, info.appDeployName(), []string{filepath.Base(info.appYamlPath()),
			filepath.Base(info.appAutomationSourcePath()), filepath.Base(info.appAutomationDeployPath())},
			filepath.Join(repoAbsPath, appKustomization))

		if err != nil {
			return fmt.Errorf("failed to create app kustomization: %w", err)
		}

		if err := configGit.Write(appKustomization, k); err != nil {
			return fmt.Errorf("failed writing app kustomization.yaml to disk: %w", err)
		}
		// TODO move to a deploy or apply command.
		userKustomization := filepath.Join(git.WegoRoot, git.WegoClusterDir, info.clusterName, "user", "kustomization.yaml")

		uk, err := a.createOrUpdateKustomize(info, params, info.appDeployName(), []string{"../../../apps/" + info.appDeployName()},
			filepath.Join(repoAbsPath, userKustomization))
		if err != nil {
			return fmt.Errorf("failed to create app kustomization: %w", err)
		}
		//TODO - need to read the existing kustomization instead of writing it here
		if err := configGit.Write(userKustomization, uk); err != nil {
			return fmt.Errorf("failed writing cluster kustomization.yaml to disk: %w", err)
		}
	} else {
		if err := a.applyToCluster(info, params.DryRun, extRepoMan.source, extRepoMan.target, extRepoMan.appDir); err != nil {
			return fmt.Errorf("could not apply manifests to the cluster: %w", err)
		}
	}

	return a.commitAndPush(configGit, AddCommitMessage, params.DryRun)
}

func (a *App) createOrUpdateKustomize(info *AppResourceInfo, params AddParams, name string, resources []string, kustomizeFile string) ([]byte, error) {
	k := types.Kustomization{}

	contents, err := os.ReadFile(kustomizeFile)
	if err == nil {
		if err := yaml.Unmarshal(contents, &k); err != nil {
			return nil, fmt.Errorf("failed to read existing kustomize file %s %w", kustomizeFile, err)
		}
	} else {
		k.MetaData = &types.ObjectMeta{
			Name:      name,
			Namespace: info.Namespace,
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

func (a *App) generateAppManifests(info *AppResourceInfo, secretRef string, appHash string) (Source, Automation, AppManifest, error) {
	var err error

	a.Logger.Generatef("Generating Source manifest")

	sourceManifest, err := a.generateSource(info, secretRef)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not set up GitOps for user repository: %w", err)
	}

	a.Logger.Generatef("Generating GitOps automation manifests")
	appGoatManifest, err := a.generateApplicationGoat(info)

	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not create GitOps automation for '%s': %w", info.Name, err)
	}

	a.Logger.Generatef("Generating Application spec manifest")

	appManifest, err := generateAppYaml(info, appHash)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not create app.yaml for '%s': %w", info.Name, err)
	}

	return sourceManifest, appGoatManifest, appManifest, nil
}

func (a *App) generateAppWegoManifests(info *AppResourceInfo) ([]byte, []byte, error) {
	appsDirManifest, err := a.Flux.CreateKustomization(
		info.automationAppsDirKustomizationName(),
		info.Name,
		info.appYamlDir(),
		info.Namespace)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create app dir kustomization for '%s': %w", info.Name, err)
	}

	targetDirManifest, err := a.Flux.CreateKustomization(
		info.automationTargetDirKustomizationName(),
		info.Name,
		info.appAutomationDir(),
		info.Namespace)
	if err != nil {
		return nil, nil, fmt.Errorf("could not create target dir kustomization for '%s': %w", info.Name, err)
	}

	return sanitizeWegoDirectory(appsDirManifest), sanitizeWegoDirectory(targetDirManifest), nil
}

func (a *App) generateExternalRepoManifests(gitProvider gitproviders.GitProvider, info *AppResourceInfo, branch string) (*externalRepoManifests, error) {
	repoName := generateResourceName(info.Spec.ConfigURL)

	secretRef := ""

	visibility, visibilityErr := gitProvider.GetRepoVisibility(context.Background(), info.configRepoUrl)
	if visibilityErr != nil {
		return nil, visibilityErr
	}

	if *visibility != gitprovider.RepositoryVisibilityPublic {
		secretRef = info.repoSecretName(info.Spec.ConfigURL).String()
	}

	targetSource, err := a.Flux.CreateSourceGit(repoName, info.Spec.ConfigURL, branch, secretRef, info.Namespace)
	if err != nil {
		return nil, fmt.Errorf("could not generate target source manifests: %w", err)
	}

	appDirGoat, err := a.Flux.CreateKustomization(
		info.automationAppsDirKustomizationName(),
		repoName,
		info.appYamlDir(),
		info.Namespace)
	if err != nil {
		return nil, fmt.Errorf("could not generate app dir kustomization for '%s': %w", info.Name, err)
	}

	targetGoat, err := a.Flux.CreateKustomization(
		info.automationTargetDirKustomizationName(),
		repoName,
		info.appAutomationDir(),
		info.Namespace)
	if err != nil {
		return nil, fmt.Errorf("could not generate target dir kustomization for '%s': %w", info.Name, err)
	}

	return &externalRepoManifests{source: targetSource, target: targetGoat, appDir: appDirGoat}, nil
}

func (a *App) commitAndPush(client git.Git, commitMsg string, dryRun bool, filters ...func(string) bool) error {
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

func (a *App) generateSource(info *AppResourceInfo, secretRef string) (Source, error) {
	var (
		b   []byte
		err error
	)

	switch info.Spec.SourceType {
	case wego.SourceTypeGit:
		b, err = a.Flux.CreateSourceGit(info.Name, info.Spec.URL, info.Spec.Branch, secretRef, info.Namespace)
	case wego.SourceTypeHelm:
		b, err = a.Flux.CreateSourceHelm(info.Name, info.Spec.URL, info.Namespace)

	default:
		return nil, fmt.Errorf("unknown source type: %v", info.Spec.SourceType)
	}

	return source{yaml: b}, err
}

func (a *App) generateApplicationGoat(info *AppResourceInfo) (Automation, error) {
	var (
		b   []byte
		err error
	)

	switch info.Spec.DeploymentType {
	case wego.DeploymentTypeKustomize:
		b, err = a.Flux.CreateKustomization(info.Name, info.Name, info.Spec.Path, info.Namespace)
	case wego.DeploymentTypeHelm:
		switch info.Spec.SourceType {
		case wego.SourceTypeHelm:
			b, err = a.Flux.CreateHelmReleaseHelmRepository(info.Name, info.Spec.Path, info.Namespace, info.Spec.HelmTargetNamespace)
		case wego.SourceTypeGit:
			b, err = a.Flux.CreateHelmReleaseGitRepository(info.Name, info.Name, info.Spec.Path, info.Namespace, info.Spec.HelmTargetNamespace)
		default:
			return nil, fmt.Errorf("invalid source type: %v", info.Spec.SourceType)
		}
	default:
		return nil, fmt.Errorf("invalid deployment type: %v", info.Spec.DeploymentType)
	}

	return automation{yaml: b}, err
}

func (a *App) applyToCluster(info *AppResourceInfo, dryRun bool, manifests ...[]byte) error {
	if dryRun {
		for _, manifest := range manifests {
			a.Logger.Printf("%s\n", manifest)
		}

		return nil
	}

	for _, manifest := range manifests {
		if err := a.Kube.Apply(context.Background(), manifest, info.Namespace); err != nil {
			return fmt.Errorf("could not apply manifest: %w", err)
		}
	}

	return nil
}

func (a *App) cloneRepo(client git.Git, url string, branch string, dryRun bool) (func(), string, error) {
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

func (a *App) writeAppYaml(configGit git.Git, info *AppResourceInfo, manifest AppManifest, overridePath func(string) string) error {
	if overridePath != nil {
		return configGit.Write(overridePath(info.appYamlPath()), manifest.ToAppYAML())
	}

	return configGit.Write(defaultMigrateToNewDirStructure(info.appYamlPath()), manifest.ToAppYAML())
}

func (a *App) writeAppGoats(configGit git.Git, info *AppResourceInfo, sourceManifest Source, deployManifest Automation, overridePath func(string) string) error {
	op := defaultMigrateToNewDirStructure
	if overridePath != nil {
		op = overridePath
	}

	if err := configGit.Write(op(info.appAutomationSourcePath()), sourceManifest.ToSourceYAML()); err != nil {
		return err
	}

	return configGit.Write(op(info.appAutomationDeployPath()), deployManifest.ToAutomationYAML())
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

func generateAppYaml(info *AppResourceInfo, appHash string) (AppManifest, error) {
	app := info.Application

	app.ObjectMeta.Labels = map[string]string{
		WeGOAppIdentifierLabelKey: appHash,
	}

	b, err := yaml.Marshal(&app)
	if err != nil {
		return nil, fmt.Errorf("could not marshal yaml: %w", err)
	}

	return appManifest{yaml: sanitizeK8sYaml(b)}, nil
}

func generateResourceName(url string) string {
	return hashNameIfTooLong(strings.ReplaceAll(utils.UrlToRepoName(url), "_", "-"))
}

func (a *App) createPullRequestToRepo(ctx context.Context, gitProvider gitproviders.GitProvider, info *AppResourceInfo, repoUrl gitproviders.RepoURL, newBranch string, appYaml AppManifest, goatSource Source, goatDeploy Automation) error {
	appPath := info.appYamlPath()
	goatSourcePath := info.appAutomationSourcePath()
	goatDeployPath := info.appAutomationDeployPath()

	appContent := string(appYaml.ToAppYAML())
	goatSourceContent := string(goatSource.ToSourceYAML())
	goatDeployContent := string(goatDeploy.ToAutomationYAML())

	files := []gitprovider.CommitFile{
		{
			Path:    &appPath,
			Content: &appContent,
		},
		{
			Path:    &goatSourcePath,
			Content: &goatSourceContent,
		},
		{
			Path:    &goatDeployPath,
			Content: &goatDeployContent,
		},
	}

	defaultBranch, err := gitProvider.GetDefaultBranch(ctx, repoUrl)
	if err != nil {
		return err
	}

	prInfo := gitproviders.PullRequestInfo{
		Title:         fmt.Sprintf("Gitops add %s", info.Name),
		Description:   fmt.Sprintf("Added yamls for %s", info.Name),
		CommitMessage: "Add App Manifests",
		TargetBranch:  defaultBranch,
		NewBranch:     newBranch,
		Files:         files,
	}

	pr, err := gitProvider.CreatePullRequest(ctx, repoUrl, prInfo)
	if err != nil {
		return fmt.Errorf("unable to create pull request: %w", err)
	}

	a.Logger.Println("Pull Request created: %s\n", pr.Get().WebURL)

	return nil
}

func getAppResourceInfo(app wego.Application, clusterName string) (*AppResourceInfo, error) {
	var (
		appRepoUrl    gitproviders.RepoURL
		configRepoUrl gitproviders.RepoURL
		err           error
	)

	if wego.DeploymentType(app.Spec.SourceType) == wego.DeploymentType(wego.SourceTypeGit) {
		appRepoUrl, err = gitproviders.NewRepoURL(app.Spec.URL)
		if err != nil {
			return nil, err
		}
	}

	if IsExternalConfigUrl(app.Spec.ConfigURL) {
		configRepoUrl, err = gitproviders.NewRepoURL(app.Spec.ConfigURL)
		if err != nil {
			return nil, err
		}
	}

	return &AppResourceInfo{
		Application:   app,
		clusterName:   clusterName,
		targetName:    clusterName,
		appRepoUrl:    appRepoUrl,
		configRepoUrl: configRepoUrl,
	}, nil
}

func (a *AppResourceInfo) configMode() ConfigMode {
	if strings.ToUpper(a.Spec.ConfigURL) == string(ConfigTypeNone) {
		return ConfigModeClusterOnly
	}

	if a.Spec.ConfigURL == string(ConfigTypeUserRepo) || a.Spec.ConfigURL == a.Spec.URL {
		return ConfigModeUserRepo
	}

	return ConfigModeExternalRepo
}

func (a *AppResourceInfo) automationRoot() string {
	root := "."

	if a.configMode() == ConfigModeUserRepo {
		root = ".wego"
	}

	return root
}

func (a *AppResourceInfo) appYamlPath() string {
	return filepath.Join(a.appYamlDir(), "app.yaml")
}

func (a *AppResourceInfo) appYamlDir() string {
	return filepath.Join(a.automationRoot(), "apps", a.Name)
}

func (a *AppResourceInfo) appAutomationSourcePath() string {
	return filepath.Join(a.appAutomationDir(), fmt.Sprintf("%s-gitops-source.yaml", a.Name))
}

func (a *AppResourceInfo) appAutomationDeployPath() string {
	return filepath.Join(a.appAutomationDir(), fmt.Sprintf("%s-gitops-deploy.yaml", a.Name))
}

func (a *AppResourceInfo) appAutomationDir() string {
	return filepath.Join(a.automationRoot(), "targets", a.clusterName, a.Name)
}

func (a *AppResourceInfo) appSourceName() string {
	return a.Name
}

func (a *AppResourceInfo) appDeployName() string {
	return a.Name
}

func (a *AppResourceInfo) appResourceName() string {
	return a.Name
}

type GeneratedSecretName string

func (s GeneratedSecretName) String() string {
	return string(s)
}

func (a *AppResourceInfo) repoSecretName(repoURL string) GeneratedSecretName {
	return CreateRepoSecretName(a.clusterName, repoURL)
}

func nameTooLong(name string) bool {
	return len(name) > maxKubernetesResourceNameLength
}

func hashNameIfTooLong(name string) string {
	if !nameTooLong(name) {
		return name
	}

	return fmt.Sprintf("wego-%x", md5.Sum([]byte(name)))
}

func CreateRepoSecretName(targetName string, repoURL string) GeneratedSecretName {
	return GeneratedSecretName(hashNameIfTooLong(fmt.Sprintf("wego-%s-%s", targetName, generateResourceName(repoURL))))
}

func (a *AppResourceInfo) automationAppsDirKustomizationName() string {
	return hashNameIfTooLong(fmt.Sprintf("%s-apps-dir", a.Name))
}

func (a *AppResourceInfo) automationTargetDirKustomizationName() string {
	return hashNameIfTooLong(fmt.Sprintf("%s-%s", a.targetName, a.Name))
}

func (a *AppResourceInfo) sourceKind() ResourceKind {
	result := ResourceKindGitRepository

	if a.Spec.SourceType == "helm" {
		result = ResourceKindHelmRepository
	}

	return result
}

func (a *AppResourceInfo) deployKind() ResourceKind {
	result := ResourceKindKustomization

	if a.Spec.DeploymentType == "helm" {
		result = ResourceKindHelmRelease
	}

	return result
}

func (a *AppResourceInfo) clusterResources() []ResourceRef {
	resources := []ResourceRef{}

	// Application GOAT, common to all three modes
	appPath := a.appYamlPath()
	automationSourcePath := a.appAutomationSourcePath()
	automationDeployPath := a.appAutomationDeployPath()

	if a.configMode() == ConfigModeClusterOnly {
		appPath = ""
		automationSourcePath = ""
		automationDeployPath = ""
	}

	resources = append(
		resources,
		ResourceRef{
			kind:           a.sourceKind(),
			name:           a.appSourceName(),
			repositoryPath: automationSourcePath},
		ResourceRef{
			kind:           a.deployKind(),
			name:           a.appDeployName(),
			repositoryPath: automationDeployPath},
		ResourceRef{
			kind:           ResourceKindApplication,
			name:           a.appResourceName(),
			repositoryPath: appPath})

	// Secret for deploy key associated with app repository;
	// common to all three modes when not using upstream Helm repository
	if a.sourceKind() == ResourceKindGitRepository {
		resources = append(
			resources,
			ResourceRef{
				kind: ResourceKindSecret,
				name: a.repoSecretName(a.Spec.URL).String()})
	}

	if strings.ToUpper(a.Spec.ConfigURL) == string(ConfigTypeNone) {
		// Only app resources present in cluster; no resources to manage config
		return resources
	}

	// App dir and target dir resources are common to app and external repo modes
	resources = append(
		resources,
		// Kustomization for .wego/apps/<app-name> directory
		ResourceRef{
			kind: ResourceKindKustomization,
			name: a.automationAppsDirKustomizationName()},
		// Kustomization for .wego/targets/<cluster-name>/<app-name> directory
		ResourceRef{
			kind: ResourceKindKustomization,
			name: a.automationTargetDirKustomizationName()})

	// External repo adds a secret and source for the external repo
	if a.Spec.ConfigURL != string(ConfigTypeUserRepo) && a.Spec.ConfigURL != a.Spec.URL {
		// Config stored in external repo
		resources = append(
			resources,
			// Secret for deploy key associated with config repository
			ResourceRef{
				kind: ResourceKindSecret,
				name: a.repoSecretName(a.Spec.ConfigURL).String()},
			// Source for config repository
			ResourceRef{
				kind: ResourceKindGitRepository,
				name: generateResourceName(a.Spec.ConfigURL)})
	}

	return resources
}

func (a *AppResourceInfo) clusterResourcePaths() []string {
	if a.configMode() == ConfigModeClusterOnly {
		return []string{}
	}

	return []string{a.appYamlPath(), a.appAutomationSourcePath(), a.appAutomationDeployPath()}
}

func (info *AppResourceInfo) getAppHash() string {
	var getHash = func(inputs ...string) string {
		final := []byte(strings.Join(inputs, ""))
		return fmt.Sprintf("%x", md5.Sum(final))
	}

	if info.Spec.DeploymentType == wego.DeploymentTypeHelm {
		return "wego-" + getHash(info.Spec.URL, info.Name, info.Spec.Branch)
	} else {
		return "wego-" + getHash(info.Spec.URL, info.Spec.Path, info.Spec.Branch)
	}
}

// NOTE: ready to save the targets automation in phase 2
// func (a *App) writeTargetGoats(basePath string, name string, manifests ...[]byte) error {
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

func (rk ResourceKind) ToGVR() (schema.GroupVersionResource, error) {
	switch rk {
	case ResourceKindApplication:
		return kube.GVRApp, nil
	case ResourceKindSecret:
		return kube.GVRSecret, nil
	case ResourceKindGitRepository:
		return kube.GVRGitRepository, nil
	case ResourceKindHelmRepository:
		return kube.GVRHelmRepository, nil
	case ResourceKindHelmRelease:
		return kube.GVRHelmRelease, nil
	case ResourceKindKustomization:
		return kube.GVRKustomization, nil
	default:
		return schema.GroupVersionResource{}, fmt.Errorf("no matching schema.GroupVersionResource to the ResourceKind: %s", string(rk))
	}
}
