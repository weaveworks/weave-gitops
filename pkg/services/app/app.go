package app

import (
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

type Dependencies struct {
	Git  git.Git
	Flux flux.Flux
	Kube kube.Kube
}

type AppService interface {
	Add(params AddParams) error
	Status(params StatusParams) error
	Install(params InstallParams) error
	Uninstall(params UninstallParams) error
}

type App struct {
	git  git.Git
	flux flux.Flux
	kube kube.Kube
}

func New(deps *Dependencies) *App {
	return &App{
		git:  deps.Git,
		flux: deps.Flux,
		kube: deps.Kube,
	}
}

// Make sure App implements all the required methods.
var _ AppService = &App{}
