package apputils

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . AppFactory
type AppFactory interface {
	GetKubeService() (kube.Kube, error)
	GetAppService(ctx context.Context, url, configUrl, namespace string, isHelmRepository bool) (app.AppService, error)
}

type DefaultAppFactory struct {
}

func (f *DefaultAppFactory) GetAppService(ctx context.Context, url, configUrl, namespace string, isHelmRepository bool) (app.AppService, error) {
	return GetAppService(ctx, url, configUrl, namespace, isHelmRepository)
}

func (f *DefaultAppFactory) GetKubeService() (kube.Kube, error) {
	kubeClient, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("error creating k8s http client: %w", err)
	}

	return kubeClient, nil
}

func GetLogger() logger.Logger {
	osysClient := osys.New()
	return logger.NewCLILogger(osysClient.Stdout())
}

func GetBaseClients() (osys.Osys, flux.Flux, kube.Kube, logger.Logger, error) {
	osysClient := osys.New()
	cliRunner := &runner.CLIRunner{}
	fluxClient := flux.New(osysClient, cliRunner)

	kubeClient, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("error creating k8s http client: %w", err)
	}

	logger := logger.NewCLILogger(osysClient.Stdout())

	return osysClient, fluxClient, kubeClient, logger, nil
}

func IsClusterReady() error {
	logger := GetLogger()

	kube, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("error creating k8s http client: %w", err)
	}

	return app.IsClusterReady(logger, kube)
}

func GetAppService(ctx context.Context, url, configUrl, namespace string, isHelmRepository bool, dryRun bool) (app.AppService, error) {
	osysClient, fluxClient, kubeClient, logger, baseClientErr := GetBaseClients()
	if baseClientErr != nil {
		return nil, fmt.Errorf("error initializing clients: %w", baseClientErr)
	}

	fluxClient.SetupBin()

	appClient, configClient, gitProvider, err := getGitClients(ctx, url, configUrl, namespace, isHelmRepository, dryRun)
	if err != nil {
		return nil, fmt.Errorf("error getting git clients: %w", err)
	}

	return app.New(ctx, logger, appClient, configClient, gitProvider, fluxClient, kubeClient, osysClient), nil
}

func getGitClients(ctx context.Context, url, configUrl, namespace string, isHelmRepository bool, dryRun bool) (git.Git, git.Git, gitproviders.GitProvider, error) {
	isExternalConfig := app.IsExternalConfigUrl(configUrl)

	var providerUrl string

	switch {
	case !isHelmRepository:
		providerUrl = url
	case isExternalConfig:
		providerUrl = configUrl
	default:
		return nil, nil, nil, nil
	}

	kube, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error creating k8s http client: %w", err)
	}

	targetName, err := kube.GetClusterName(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error getting target name: %w", err)
	}

	authsvc, err := getAuthService(ctx, providerUrl, dryRun)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("error creating auth service: %w", err)
	}

	var appClient, configClient git.Git

	if !isHelmRepository {
		// We need to do this even if we have an external config to set up the deploy key for the app repo
		appRepoClient, err := authsvc.CreateGitClient(ctx, targetName, namespace, utils.SanitizeRepoUrl(url))
		if err != nil {
			return nil, nil, nil, err
		}

		appClient = appRepoClient
	}

	if isExternalConfig {
		configRepoClient, err := authsvc.CreateGitClient(ctx, targetName, namespace, utils.SanitizeRepoUrl(configUrl))
		if err != nil {
			return nil, nil, nil, err
		}

		configClient = configRepoClient
	} else {
		configClient = appClient
	}

	return appClient, configClient, authsvc.GetGitProvider(), nil
}

func getAuthService(ctx context.Context, providerUrl string, dryRun bool) (auth.AuthService, error) {
	var (
		gitProvider gitproviders.GitProvider
		err         error
	)

	if dryRun {
		if gitProvider, err = gitproviders.NewDryRun(); err != nil {
			return nil, fmt.Errorf("error creating git provider client: %w", err)
		}
	} else {
		if gitProvider, err = auth.GetGitProvider(ctx, providerUrl); err != nil {
			return nil, fmt.Errorf("error obtaining git provider token: %w", err)
		}
	}

	osysClient := osys.New()
	cliRunner := &runner.CLIRunner{}
	fluxClient := flux.New(osysClient, cliRunner)
	logger := logger.NewCLILogger(osysClient.Stdout())

	_, rawClient, err := kube.NewKubeHTTPClient()
	if err != nil {
		return nil, fmt.Errorf("error creating k8s http client: %w", err)
	}

	return auth.NewAuthService(fluxClient, rawClient, gitProvider, logger)
}
