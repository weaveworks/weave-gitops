package apputils

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//counterfeiter:generate . ServerAppFactory
type ServerAppFactory interface {
	GetKubeService() (kube.Kube, error)
	GetAppService(ctx context.Context, params AppServiceParams) (app.AppService, error)
}

type factory struct {
	k client.Client
	l logger.Logger
}

func NewServerAppFactory(k client.Client, l logger.Logger) ServerAppFactory {
	return factory{k: k, l: l}
}

func (f factory) GetKubeService() (kube.Kube, error) {
	k, _, err := kube.NewKubeHTTPClient()
	return k, err
}

func (f factory) GetAppService(ctx context.Context, params AppServiceParams) (app.AppService, error) {
	appURL, err := gitproviders.NewNormalizedRepoURL(params.URL)
	if err != nil {
		return nil, fmt.Errorf("error creating normalized url for app url: %w", err)
	}

	configURL := appURL
	if params.ConfigURL != "" {
		configURL, err = gitproviders.NewNormalizedRepoURL(params.ConfigURL)
		if err != nil {
			return nil, fmt.Errorf("error creating normalized url for config url: %w", err)
		}
	}

	cfg := gitproviders.Config{
		Provider: configURL.Provider(),
		Token:    params.Token,
	}

	provider, err := gitproviders.New(cfg, configURL.Owner(), gitproviders.GetAccountType)
	if err != nil {
		return nil, fmt.Errorf("error creating git provider: %w", err)
	}

	clients, err := GetBaseClients()
	if err != nil {
		return nil, fmt.Errorf("error initializing clients: %w", err)
	}

	// Note that we assume the same git provider here.
	// If someone has an app source in Github and a config repo in Gitlab, we will get auth errors.
	authSvc, err := auth.NewAuthService(clients.Flux, f.k, provider, f.l)
	if err != nil {
		return nil, fmt.Errorf("error creating auth service: %w", err)
	}

	clusterName, err := clients.Kube.GetClusterName(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not get cluster name: %w", err)
	}

	appGit, err := authSvc.CreateGitClient(ctx, appURL, clusterName, params.Namespace)
	if err != nil {
		return nil, fmt.Errorf("error creating git client for app repo: %w", err)
	}

	configGit, err := authSvc.CreateGitClient(ctx, configURL, clusterName, params.Namespace)
	if err != nil {
		return nil, fmt.Errorf("error creating git client for config repo: %w", err)
	}

	appSrv := app.New(ctx, logger.NewApiLogger(), appGit, configGit, provider, clients.Flux, clients.Kube, clients.Osys)

	return appSrv, nil
}
