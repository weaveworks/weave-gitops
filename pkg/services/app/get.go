package app

import (
	"context"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
)

func (a *App) Get(name, namespace string) (*wego.Application, error) {
	return a.kube.GetApplication(context.Background(), name, namespace)
}
