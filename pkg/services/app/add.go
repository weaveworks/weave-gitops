package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"github.com/weaveworks/weave-gitops/pkg/services/gitopswriter"
	"github.com/weaveworks/weave-gitops/pkg/services/gitrepo"
	"github.com/weaveworks/weave-gitops/pkg/utils"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
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
	ConfigRepo                 string
	Namespace                  string
	DryRun                     bool
	AutoMerge                  bool
	GitProviderToken           string
	HelmReleaseTargetNamespace string
	MigrateToNewDirStructure   func(string) string
}

const (
	DefaultPath           = "./"
	DefaultBranch         = "main"
	DefaultDeploymentType = "kustomize"
)

func (a AddParams) IsHelmRepository() bool {
	return a.Chart != ""
}

func (a *AppSvc) Add(configGit git.Git, gitProvider gitproviders.GitProvider, params AddParams) error {
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

	app, err := makeApplication(params)
	if err != nil {
		return err
	}

	if strings.HasPrefix(params.Name, "wego") {
		return fmt.Errorf("the prefix 'wego' is used by weave gitops and is not allowed for an app name")
	}

	appHash := automation.GetAppHash(app)

	wegoapps, err := a.Kube.GetApplications(ctx, params.Namespace)
	if err != nil {
		return err
	}

	for _, wegoapp := range wegoapps {
		clusterApp, err := automation.WegoAppToApp(wegoapp)
		if err != nil {
			return err
		}

		if appHash == automation.GetAppHash(clusterApp) {
			return fmt.Errorf("unable to create resource, resource already exists in cluster")
		}
	}

	if params.DryRun {
		return nil
	}

	return a.addApp(ctx, configGit, gitProvider, app, clusterName, params.AutoMerge)
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
			if err := models.ValidateApplicationName(params.Chart); err != nil {
				return params, fmt.Errorf("unable to use chart name %q as the application name; please specify name with '--name' :%s",
					params.Chart, err)
			}

			params.Name = params.Chart
		}

		if params.Url == "" {
			return params, fmt.Errorf("--url must be specified for helm repositories")
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
	if models.IsExternalConfigRepo(params.ConfigRepo) {
		configRepoUrl, err := gitproviders.NewRepoURL(params.ConfigRepo)
		if err != nil {
			return params, fmt.Errorf("error normalizing url: %w", err)
		}

		params.ConfigRepo = configRepoUrl.String()
	}

	if params.Name == "" {
		repoName := utils.UrlToRepoName(params.Url)
		if err := models.ValidateApplicationName(repoName); err != nil {
			return params, err
		}

		params.Name = automation.GenerateResourceName(appRepoUrl)
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

	if err := models.ValidateApplicationName(params.Name); err != nil {
		return params, err
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

func (a *AppSvc) addApp(ctx context.Context, configGit git.Git, gitProvider gitproviders.GitProvider, app models.Application, clusterName string, autoMerge bool) error {
	repoWriter := gitrepo.NewRepoWriter(app.ConfigRepo, gitProvider, configGit, a.Logger)
	automationGen := automation.NewAutomationGenerator(gitProvider, a.Flux, a.Logger)
	gitOpsDirWriter := gitopswriter.NewGitOpsDirectoryWriter(automationGen, repoWriter, a.Osys, a.Logger)

	return gitOpsDirWriter.AddApplication(ctx, app, clusterName, autoMerge)
}

func makeApplication(params AddParams) (models.Application, error) {
	var (
		gitSourceURL  gitproviders.RepoURL
		helmSourceURL string
		err           error
	)

	if models.SourceType(params.SourceType) == models.SourceTypeHelm {
		helmSourceURL = params.Url
	} else {
		gitSourceURL, err = gitproviders.NewRepoURL(params.Url)
		if err != nil {
			return models.Application{}, err
		}
	}

	configRepo := gitSourceURL

	if params.ConfigRepo != "" {
		curl, err := gitproviders.NewRepoURL(params.ConfigRepo)
		if err != nil {
			return models.Application{}, err
		}

		configRepo = curl
	}

	app := models.Application{
		Name:                params.Name,
		Namespace:           params.Namespace,
		GitSourceURL:        gitSourceURL,
		HelmSourceURL:       helmSourceURL,
		ConfigRepo:          configRepo,
		Branch:              params.Branch,
		Path:                params.Path,
		SourceType:          models.SourceType(params.SourceType),
		AutomationType:      models.AutomationType(params.DeploymentType),
		HelmTargetNamespace: params.HelmReleaseTargetNamespace,
	}

	return app, nil
}
