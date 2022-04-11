package services

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . Factory

// Factory provides helpers for generating various WeGO service objects at runtime.
type Factory interface {
	GetGitClients(ctx context.Context, kubeClient kube.Kube, gpClient gitproviders.Client, params GitConfigParams) (git.Git, gitproviders.GitProvider, error)
}

type GitConfigParams struct {
	URL              string
	ConfigRepo       string
	Namespace        string
	IsHelmRepository bool
	DryRun           bool
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

func (f *defaultFactory) GetGitClients(ctx context.Context, kubeClient kube.Kube, gpClient gitproviders.Client, params GitConfigParams) (git.Git, gitproviders.GitProvider, error) {
	configNormalizedUrl, err := gitproviders.NewRepoURL(params.ConfigRepo)
	if err != nil {
		return nil, nil, fmt.Errorf("error normalizing config url: %w", err)
	}

	authSvc, err := f.getAuthService(kubeClient, configNormalizedUrl, gpClient, params.DryRun)
	if err != nil {
		return nil, nil, fmt.Errorf("error getting auth service: %w", err)
	}

	client, err := authSvc.CreateGitClient(ctx, configNormalizedUrl, params.Namespace, params.DryRun)
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
