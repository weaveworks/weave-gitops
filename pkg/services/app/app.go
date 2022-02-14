package app

import (
	"context"

	"github.com/benbjohnson/clock"
	"github.com/fluxcd/go-git-providers/gitprovider"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"k8s.io/apimachinery/pkg/types"
)

// AppService entity that manages applications
type AppService interface {
	// Add adds a new application to the cluster
	Add(configGit git.Git, gitProvider gitproviders.GitProvider, params AddParams) error
	// Get returns a given applicaiton
	Get(name types.NamespacedName) (*wego.Application, error)
	// GetCommits returns a list of commits for an application
	GetCommits(gitProvider gitproviders.GitProvider, params CommitParams, application *wego.Application) ([]gitprovider.Commit, error)
	// Remove removes an application from the cluster
	Remove(configGit git.Git, gitProvider gitproviders.GitProvider, params RemoveParams) error
	// Status returns flux resources status and the last successful reconciliation time
	Status(params StatusParams) (string, string, error)
	// Sync trigger reconciliation loop for an application
	Sync(params SyncParams) error
}

type AppSvc struct {
	Context context.Context
	Osys    osys.Osys
	Flux    flux.Flux
	Kube    kube.Kube
	Logger  logger.Logger
	Clock   clock.Clock
}

func New(ctx context.Context, logger logger.Logger, flux flux.Flux, kube kube.Kube, osys osys.Osys) AppService {
	return &AppSvc{
		Context: ctx,
		Flux:    flux,
		Kube:    kube,
		Logger:  logger,
		Osys:    osys,
		Clock:   clock.New(),
	}
}

// Make sure App implements all the required methods.
var _ AppService = &AppSvc{}

func (a *AppSvc) getDeploymentType(ctx context.Context, name string, namespace string) (wego.DeploymentType, error) {
	app, err := a.Kube.GetApplication(ctx, types.NamespacedName{Name: name, Namespace: namespace})
	if err != nil {
		return wego.DeploymentTypeKustomize, err
	}

	return app.Spec.DeploymentType, nil
}
