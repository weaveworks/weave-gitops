package app

import (
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

func (a *AppSvc) Get(name types.NamespacedName) (*wego.Application, error) {
	return a.Kube.GetApplication(a.Context, name)
}
