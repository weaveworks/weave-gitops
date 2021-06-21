package app

import (
	k8sApps "github.com/weaveworks/weave-gitops/api/v1alpha"
)

func (a *App) Get(name string) (*k8sApps.Application, error) {
	return a.kube.GetApplication(name)
}
