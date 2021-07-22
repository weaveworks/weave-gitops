package app

import "github.com/weaveworks/weave-gitops/pkg/flux"

type UnpauseParams struct {
	Name      string
	Namespace string
}

func (a *App) Unpause(params UnpauseParams) error {
	return a.pauseOrUnpause(flux.Resume, params.Name, params.Namespace)
}
