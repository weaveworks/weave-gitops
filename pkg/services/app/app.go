package app

import (
	wego "github.com/weaveworks/weave-gitops/api/v1alpha"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

type AppService interface {
	Add(params AddParams) error
	Get(name string) (*wego.Application, error)
}

type App struct {
	git          git.Git
	flux         flux.Flux
	kube         kube.Kube
	gitProviders gitproviders.GitProviderHandler
}

func New(git git.Git, flux flux.Flux, kube kube.Kube, gitProviders gitproviders.GitProviderHandler) *App {
	return &App{
		git:          git,
		flux:         flux,
		kube:         kube,
		gitProviders: gitProviders,
	}
}

// Make sure App implements all the required methods.
var _ AppService = &App{}
