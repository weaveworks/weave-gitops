package apputils

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

//counterfeiter:generate . ServerAppFactory
type ServerAppFactory interface {
	GetKubeService() (kube.Kube, error)
	GetAppService(ctx context.Context, params AppServiceParams) (app.AppService, error)
}

type factory struct {
	k           client.Client
	l           logger.Logger
	rest        *rest.Config
	clusterName string
}

func NewServerAppFactory(rest *rest.Config, l logger.Logger, clusterName string) (ServerAppFactory, error) {
	_, k, err := kube.NewKubeHTTPClientWithConfig(rest, clusterName)
	if err != nil {
		return nil, err
	}

	return factory{
		k:           k,
		l:           l,
		clusterName: clusterName,
		rest:        rest,
	}, nil
}

func (f factory) GetKubeService() (kube.Kube, error) {
	k, _, err := kube.NewKubeHTTPClientWithConfig(f.rest, f.clusterName)
	return k, err
}

func (f factory) GetAppService(ctx context.Context, params AppServiceParams) (app.AppService, error) {
	osysClient := osys.New()
	fluxClient := flux.New(osysClient, &runner.CLIRunner{})

	kube, _, err := kube.NewKubeHTTPClientWithConfig(f.rest, f.clusterName)
	if err != nil {
		return nil, err
	}

	appURL, err := gitproviders.NewRepoURL(params.URL)
	if err != nil {
		return nil, fmt.Errorf("error creating normalized url for app url: %w", err)
	}

	configURL := appURL
	if params.ConfigURL != "" {
		configURL, err = gitproviders.NewRepoURL(params.ConfigURL)
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

	// Note that we assume the same git provider here.
	// If someone has an app source in Github and a config repo in Gitlab, we will get auth errors.
	authSvc, err := auth.NewAuthService(fluxClient, f.k, provider, f.l)
	if err != nil {
		return nil, fmt.Errorf("error creating auth service: %w", err)
	}

	appGit, err := authSvc.CreateGitClient(ctx, appURL, f.clusterName, params.Namespace)
	if err != nil {
		return nil, fmt.Errorf("error creating git client for app repo: %w", err)
	}

	configGit, err := authSvc.CreateGitClient(ctx, configURL, f.clusterName, params.Namespace)
	if err != nil {
		return nil, fmt.Errorf("error creating git client for config repo: %w", err)
	}

	appSrv := app.New(ctx, f.l, appGit, configGit, provider, fluxClient, kube, osysClient,
		automation.NewAutomationService(provider, fluxClient, f.l))

	return appSrv, nil
}
