package app

import (
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
)

func (a *App) Get(name string) (*wego.Application, error) {
	return a.kube.GetApplication(name)
}
