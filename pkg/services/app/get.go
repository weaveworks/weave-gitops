package app

import (
	"context"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

func (a *App) Get(name types.NamespacedName) (*wego.Application, error) {
	return a.Kube.GetApplication(context.Background(), name)
}
