package services

import (
	"context"
	"fmt"

	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Factory

// Factory provides helpers for generating various WeGO service objects at runtime.
type Factory interface {
	GetKubeService() (kube.Kube, error)
	GetAppService(ctx context.Context) (app.AppService, error)
	GetGitClients(ctx context.Context, gpClient gitproviders.Client, params GitConfigParams) (git.Git, gitproviders.GitProvider, error)
}

// GitConfigParams represents the client configuration for accessing a Git
// repository.
type GitConfigParams struct {
	URL               string
	ConfigRepo        string
	Namespace         string
	IsHelmRepository  bool
	DryRun            bool
	FluxHTTPSUsername string
	FluxHTTPSPassword string
}

// NewGitConfigParamsFromApp allocates and returns a set of parameters using
// fields from the Application.
func NewGitConfigParamsFromApp(app *wego.Application, dryRun bool) GitConfigParams {
	isHelmRepository := app.Spec.SourceType == wego.SourceTypeHelm
	return GitConfigParams{
		URL:              app.Spec.URL,
		ConfigRepo:       app.Spec.ConfigRepo,
		Namespace:        app.Namespace,
		IsHelmRepository: isHelmRepository,
		DryRun:           dryRun,
	}
}

type defaultFactory struct {
	fluxClient  flux.Flux
	log         logger.Logger
	rest        *rest.Config
	clusterName string
}

func NewFactory(fluxClient flux.Flux, log logger.Logger) Factory {
	return &defaultFactory{
		fluxClient: fluxClient,
		log:        log,
	}
}

func NewServerFactory(fluxClient flux.Flux, log logger.Logger, rest *rest.Config, clusterName string) Factory {
	return &defaultFactory{
		fluxClient:  fluxClient,
		log:         log,
		rest:        rest,
		clusterName: clusterName,
	}
}

func (f *defaultFactory) GetAppService(ctx context.Context) (app.AppService, error) {
	kubeClient, err := f.GetKubeService()
	if err != nil {
		return nil, fmt.Errorf("error initializing clients: %w", err)
	}

	return app.New(ctx, f.log, f.fluxClient, kubeClient, osys.New()), nil
}

func (f *defaultFactory) GetKubeService() (kube.Kube, error) {
	var mainKubeClient kube.Kube

	if f.rest == nil {
		kubeClient, _, err := kube.NewKubeHTTPClient()
		if err != nil {
			return nil, fmt.Errorf("error creating k8s http client: %w", err)
		}

		mainKubeClient = kubeClient
	} else {
		kubeClient, _, err := kube.NewKubeHTTPClientWithConfig(f.rest, f.clusterName)
		if err != nil {
			return nil, err
		}

		mainKubeClient = kubeClient
	}

	return mainKubeClient, nil
}

func (f *defaultFactory) GetGitClients(ctx context.Context, gpClient gitproviders.Client, params GitConfigParams) (git.Git, gitproviders.GitProvider, error) {
	isExternalConfig := models.IsExternalConfigRepo(params.ConfigRepo)
	var providerUrl string

	switch {
	case !params.IsHelmRepository:
		providerUrl = params.URL
	case isExternalConfig:
		providerUrl = params.ConfigRepo
	default:
		return nil, nil, nil
	}

	normalizedUrl, err := gitproviders.NewRepoURL(providerUrl)
	if err != nil {
		return nil, nil, fmt.Errorf("error normalizing url: %w", err)
	}

	kube, err := f.GetKubeService()
	if err != nil {
		return nil, nil, fmt.Errorf("error creating k8s http client: %w", err)
	}

	targetName, err := kube.GetClusterName(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting target name: %w", err)
	}

	authSvc, err := f.getAuthService(normalizedUrl, gpClient, params.DryRun)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating auth service: %w", err)
	}

	var httpsCreds *flux.HTTPSCreds
	if params.FluxHTTPSUsername != "" && params.FluxHTTPSPassword != "" {
		httpsCreds = &flux.HTTPSCreds{Username: params.FluxHTTPSUsername, Password: params.FluxHTTPSPassword}
	}

	var appClient, configClient git.Git
	// TODO: KEVIN!!!! Do we need to make these flux.HTTPSCreds?
	if !params.IsHelmRepository {
		// We need to do this even if we have an external config to set up the deploy key for the app repo
		appRepoClient, appRepoErr := authSvc.CreateGitClient(ctx, normalizedUrl, targetName, params.Namespace, params.DryRun, httpsCreds)
		if appRepoErr != nil {
			return nil, nil, appRepoErr
		}

		appClient = appRepoClient
	}

	if isExternalConfig {
		normalizedConfigRepo, err := gitproviders.NewRepoURL(params.ConfigRepo)
		if err != nil {
			return nil, nil, fmt.Errorf("error normalizing url: %w", err)
		}

		configRepoClient, configRepoErr := authSvc.CreateGitClient(ctx, normalizedConfigRepo, targetName, params.Namespace, params.DryRun, httpsCreds)
		if configRepoErr != nil {
			return nil, nil, configRepoErr
		}

		configClient = configRepoClient
	} else {
		configClient = appClient
	}

	return configClient, authSvc.GetGitProvider(), nil
}

func (f *defaultFactory) getAuthService(normalizedUrl gitproviders.RepoURL, gpClient gitproviders.Client, dryRun bool) (auth.AuthService, error) {
	var (
		gitProvider gitproviders.GitProvider
		err         error
	)

	if dryRun {
		if gitProvider, err = gitproviders.NewDryRun(); err != nil {
			return nil, fmt.Errorf("error creating git provider client: %w", err)
		}
	} else {
		if gitProvider, err = gpClient.GetProvider(normalizedUrl, gitproviders.GetAccountType); err != nil {
			return nil, fmt.Errorf("error obtaining git provider token: %w", err)
		}
	}

	var rawClient client.Client
	if f.rest == nil {
		_, rawClient, err = kube.NewKubeHTTPClient()
		if err != nil {
			return nil, fmt.Errorf("error creating k8s http client: %w", err)
		}
	} else {
		_, rawClient, err = kube.NewKubeHTTPClientWithConfig(f.rest, f.clusterName)
		if err != nil {
			return nil, fmt.Errorf("error creating k8s http client: %w", err)
		}
	}

	return auth.NewAuthService(f.fluxClient, rawClient, gitProvider, f.log)
}
