package services

import (
	"context"
	"fmt"

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
	GetAppService(ctx context.Context, kubeClient kube.Kube) (app.AppService, error)
	GetGitClients(ctx context.Context, kubeClient kube.Kube, gpClient gitproviders.Client, params GitConfigParams) (git.Git, gitproviders.GitProvider, error)
}

type GitConfigParams struct {
	URL              string
	ConfigRepo       string
	Namespace        string
	IsHelmRepository bool
	DryRun           bool
}

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
	fluxClient flux.Flux
	log        logger.Logger
}

func NewFactory(fluxClient flux.Flux, log logger.Logger) Factory {
	return &defaultFactory{
		fluxClient: fluxClient,
		log:        log,
	}
}

func (f *defaultFactory) GetAppService(ctx context.Context, kubeClient kube.Kube) (app.AppService, error) {
	return app.New(ctx, f.log, f.fluxClient, kubeClient, osys.New()), nil
}

func (f *defaultFactory) GetGitClients(ctx context.Context, kubeClient kube.Kube, gpClient gitproviders.Client, params GitConfigParams) (git.Git, gitproviders.GitProvider, error) {
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

	normalizedUrl, err := gitproviders.NewRepoURL(providerUrl, false)
	if err != nil {
		return nil, nil, fmt.Errorf("error normalizing url: %w", err)
	}

	authSvc, err := f.getAuthService(kubeClient, normalizedUrl, gpClient, params.DryRun)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting auth service: %w", err)
	}

	var appClient, configClient git.Git

	if !params.IsHelmRepository {
		// We need to do this even if we have an external config to set up the deploy key for the app repo
		appRepoClient, appRepoErr := authSvc.CreateGitClient(ctx, normalizedUrl, params.Namespace, params.DryRun)
		if appRepoErr != nil {
			return nil, nil, appRepoErr
		}

		appClient = appRepoClient
	}

	if isExternalConfig {
		normalizedConfigRepo, err := gitproviders.NewRepoURL(params.ConfigRepo, true)
		if err != nil {
			return nil, nil, fmt.Errorf("error normalizing url: %w", err)
		}

		configRepoClient, configRepoErr := authSvc.CreateGitClient(ctx, normalizedConfigRepo, params.Namespace, params.DryRun)
		if configRepoErr != nil {
			return nil, nil, configRepoErr
		}

		configClient = configRepoClient
	} else {
		configClient = appClient
	}

	return configClient, authSvc.GetGitProvider(), nil
}

func (f *defaultFactory) getAuthService(kubeClient kube.Kube, normalizedUrl gitproviders.RepoURL, gpClient gitproviders.Client, dryRun bool) (auth.AuthService, error) {
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

	return auth.NewAuthService(f.fluxClient, kubeClient.Raw(), gitProvider, f.log)
}
