package gitops

import (
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

type GitopsService interface {
	Install(params InstallParams) ([]byte, error)
	Uninstall(params UinstallParams) error
}

type Gitops struct {
	flux flux.Flux
	kube kube.Kube
}

func New(flux flux.Flux, kube kube.Kube) *Gitops {
	return &Gitops{
		flux: flux,
		kube: kube,
	}
}

// Make sure App implements all the required methods.
var _ GitopsService = &Gitops{}
