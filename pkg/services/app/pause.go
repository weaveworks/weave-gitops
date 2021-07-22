package app

import "github.com/weaveworks/weave-gitops/pkg/flux"

type PauseParams struct {
	Name      string
	Namespace string
}

func (a *App) Pause(params PauseParams) error {
	return a.pauseOrUnpause(flux.Suspend, params.Name, params.Namespace)
}
