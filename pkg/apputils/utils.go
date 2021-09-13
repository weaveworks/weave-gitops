package apputils

import (
	"context"
	"fmt"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
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
	"k8s.io/apimachinery/pkg/types"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate
//counterfeiter:generate . AppFactory
type AppFactory interface {
	GetKubeService() (kube.Kube, error)
	GetAppService(ctx context.Context, name, namespace string) (app.AppService, error)
}

type DefaultAppFactory struct {
}

func (f *DefaultAppFactory) GetAppService(ctx context.Context, name, namespace string) (app.AppService, error) {
	return GetAppService(ctx, name, namespace)
}

func (f *DefaultAppFactory) GetKubeService() (kube.Kube, error) {
	kubeClient, _, kubeErr := kube.NewKubeHTTPClient()
	if kubeErr != nil {
		return nil, fmt.Errorf("error creating k8s http client: %w", kubeErr)
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

	kubeClient, _, kubeErr := kube.NewKubeHTTPClient()
	if kubeErr != nil {
		return nil, nil, nil, nil, fmt.Errorf("error creating k8s http client: %w", kubeErr)
	}

	logger := logger.NewCLILogger(osysClient.Stdout())

	return osysClient, fluxClient, kubeClient, logger, nil
}

func IsClusterReady() error {
	logger := GetLogger()

	kube, _, kubeErr := kube.NewKubeHTTPClient()
	if kubeErr != nil {
		return fmt.Errorf("error creating k8s http client: %w", kubeErr)
	}

	if readyErr := app.IsClusterReady(logger, kube); readyErr != nil {
		return readyErr
	}

	return nil
}

func GetAppService(ctx context.Context, appName, namespace string) (app.AppService, error) {
	osysClient, fluxClient, kubeClient, logger, baseClientErr := GetBaseClients()
	if baseClientErr != nil {
		return nil, fmt.Errorf("error initializing clients: %w", baseClientErr)
	}

	appClient, configClient, gitProvider, clientErr := getGitClientsForApp(ctx, appName, namespace)
	if clientErr != nil {
		return nil, fmt.Errorf("error getting git clients: %w", clientErr)
	}

	return app.New(ctx, logger, appClient, configClient, gitProvider, fluxClient, kubeClient, osysClient), nil
}

func GetAppServiceForAdd(ctx context.Context, url, configUrl, namespace string, isHelmRepository bool) (app.AppService, error) {
	osysClient, fluxClient, kubeClient, logger, baseClientErr := GetBaseClients()
	if baseClientErr != nil {
		return nil, fmt.Errorf("error initializing clients: %w", baseClientErr)
	}

	appClient, configClient, gitProvider, clientErr := getGitClients(ctx, url, configUrl, namespace, isHelmRepository)
	if clientErr != nil {
		return nil, fmt.Errorf("error getting git clients: %w", clientErr)
	}

	return app.New(ctx, logger, appClient, configClient, gitProvider, fluxClient, kubeClient, osysClient), nil
}

func getGitClientsForApp(ctx context.Context, appName, namespace string) (git.Git, git.Git, gitproviders.GitProvider, error) {
	kube, _, kubeErr := kube.NewKubeHTTPClient()
	if kubeErr != nil {
		return nil, nil, nil, fmt.Errorf("error creating k8s http client: %w", kubeErr)
	}

	app, appErr := kube.GetApplication(ctx, types.NamespacedName{Namespace: namespace, Name: appName})
	if appErr != nil {
		return nil, nil, nil, fmt.Errorf("could not retrieve application %q: %w", appName, appErr)
	}

	return getGitClients(ctx, app.Spec.URL, app.Spec.ConfigURL, namespace, app.Spec.SourceType == wego.SourceTypeHelm)
}

func getGitClients(ctx context.Context, url, configUrl, namespace string, isHelmRepository bool) (git.Git, git.Git, gitproviders.GitProvider, error) {
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

	kube, _, kubeErr := kube.NewKubeHTTPClient()
	if kubeErr != nil {
		return nil, nil, nil, fmt.Errorf("error creating k8s http client: %w", kubeErr)
	}

	targetName, targetErr := kube.GetClusterName(ctx)
	if targetErr != nil {
		return nil, nil, nil, fmt.Errorf("error getting target name: %w", targetErr)
	}

	authsvc, authsvcErr := getAuthService(ctx, providerUrl)
	if authsvcErr != nil {
		return nil, nil, nil, fmt.Errorf("error creating auth service: %w", authsvcErr)
	}

	var appClient, configClient git.Git

	if !isHelmRepository {
		// We need to do this even if we have an external config to set up the deploy key for the app repo
		appRepoClient, appRepoErr := authsvc.CreateGitClient(ctx, targetName, namespace, url)
		if appRepoErr != nil {
			return nil, nil, nil, appRepoErr
		}

		appClient = appRepoClient
	}

	if isExternalConfig {
		configRepoClient, configRepoErr := authsvc.CreateGitClient(ctx, targetName, namespace, utils.SanitizeRepoUrl(configUrl))
		if configRepoErr != nil {
			return nil, nil, nil, configRepoErr
		}

		configClient = configRepoClient
	} else {
		configClient = appClient
	}

	return appClient, configClient, authsvc.GetGitProvider(), nil
}

func getAuthService(ctx context.Context, providerUrl string) (auth.AuthService, error) {
	gitProvider, providerErr := auth.GetGitProvider(ctx, providerUrl)
	if providerErr != nil {
		return nil, fmt.Errorf("error obtaining git provider token: %w", providerErr)
	}

	osysClient := osys.New()
	cliRunner := &runner.CLIRunner{}
	fluxClient := flux.New(osysClient, cliRunner)
	logger := logger.NewCLILogger(osysClient.Stdout())

	_, rawClient, kubeErr := kube.NewKubeHTTPClient()
	if kubeErr != nil {
		return nil, fmt.Errorf("error creating k8s http client: %w", kubeErr)
	}

	return auth.NewAuthService(fluxClient, rawClient, gitProvider, logger)
}
