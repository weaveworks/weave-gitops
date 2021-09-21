package gitops

import (
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

type GitopsService interface {
	Install(params InstallParams) ([]byte, error)
	Uninstall(params UninstallParams) error
}

type Gitops struct {
	flux   flux.Flux
	kube   kube.Kube
	logger logger.Logger
}

func New(logger logger.Logger, flux flux.Flux, kube kube.Kube) *Gitops {
	return &Gitops{
		flux:   flux,
		kube:   kube,
		logger: logger,
	}
}

// Make sure App implements all the required methods.
var _ GitopsService = &Gitops{}
