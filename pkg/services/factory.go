package services

import (
	"context"
	"fmt"

	"github.com/fluxcd/go-git-providers/gitprovider"
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
	if params.DryRun {
		return nil, nil, nil
	}

	configNormalizedUrl, err := gitproviders.NewRepoURL(params.ConfigRepo)
	if err != nil {
		return nil, nil, fmt.Errorf("error normalizing config url: %w", err)
	}

	targetName, err := kubeClient.GetClusterName(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting target name: %w", err)
	}

	authSvc, err := f.getAuthService(kubeClient, configNormalizedUrl, gpClient, params.DryRun)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating auth service: %w", err)
	}

	// Do not add deploy key for helm repo, empty url or if its gonna be added below
	if !params.IsHelmRepository && params.URL != "" && params.URL != params.ConfigRepo {
		normalizedUrl, err := gitproviders.NewRepoURL(params.URL)
		if err != nil {
			return nil, nil, fmt.Errorf("error normalizing url: %w", err)
		}

		provider := authSvc.GetGitProvider()

		repoVisibility, err := provider.GetRepoVisibility(ctx, normalizedUrl)
		if err != nil {
			return nil, nil, fmt.Errorf("error getting repo visibility: %w", err)
		}

		// Do not add deploy key for public repo. Issue https://github.com/weaveworks/weave-gitops/issues/1111
		if *repoVisibility == gitprovider.RepositoryVisibilityPrivate {
			secretName := auth.SecretName{
				Name:      models.CreateRepoSecretName(normalizedUrl),
				Namespace: params.Namespace,
			}

			_, err = authSvc.SetupDeployKey(ctx, secretName, targetName, normalizedUrl)
			if err != nil {
				return nil, nil, fmt.Errorf("error setting up deploy key: %w", err)
			}
		}
	}

	client, err := authSvc.CreateGitClient(ctx, configNormalizedUrl, targetName, params.Namespace, params.DryRun)
	if err != nil {
		return nil, nil, err
	}

	return client, authSvc.GetGitProvider(), nil
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
