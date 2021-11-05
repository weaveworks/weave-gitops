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

type GitConfigParams struct {
	URL              string
	ConfigURL        string
	Namespace        string
	IsHelmRepository bool
	DryRun           bool
}

func NewGitConfigParamsFromApp(app *wego.Application, dryRun bool) GitConfigParams {
	isHelmRepository := app.Spec.SourceType == wego.SourceTypeHelm

	return GitConfigParams{
		URL:              app.Spec.URL,
		ConfigURL:        app.Spec.ConfigURL,
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

func NewFactory(fluxClient flux.Flux, log logger.Logger, rest *rest.Config, clusterName string) Factory {
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

	return app.New(ctx, f.log, f.fluxClient, kubeClient), nil
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
	isExternalConfig := app.IsExternalConfigUrl(params.ConfigURL)

	var providerUrl string

	switch {
	case !params.IsHelmRepository:
		providerUrl = params.URL
	case isExternalConfig:
		providerUrl = params.ConfigURL
	default:
		return nil, nil, nil
	}

	normalizedUrl, err := gitproviders.NewRepoURL(providerUrl)
	if err != nil {
		return nil, nil, fmt.Errorf("error normalizing url: %w", err)
	}

	kube, err := f.GetKubeService()
	if err != nil {
		return nil, nil, fmt.Errorf("some descriptive error")
	}

	targetName, err := kube.GetClusterName(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting target name: %w", err)
	}

	authSvc, err := f.getAuthService(normalizedUrl, gpClient, params.DryRun)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating auth service: %w", err)
	}

	var appClient, configClient git.Git

	if !params.IsHelmRepository {
		// We need to do this even if we have an external config to set up the deploy key for the app repo
		appRepoClient, appRepoErr := authSvc.CreateGitClient(ctx, normalizedUrl, targetName, params.Namespace)
		if appRepoErr != nil {
			return nil, nil, appRepoErr
		}

		appClient = appRepoClient
	}

	if isExternalConfig {
		normalizedConfigUrl, err := gitproviders.NewRepoURL(params.ConfigURL)
		if err != nil {
			return nil, nil, fmt.Errorf("error normalizing url: %w", err)
		}

		configRepoClient, configRepoErr := authSvc.CreateGitClient(ctx, normalizedConfigUrl, targetName, params.Namespace)
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
