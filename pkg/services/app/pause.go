package app

import wego "github.com/weaveworks/weave-gitops/api/v1alpha1"

type PauseParams struct {
	Name      string
	Namespace string
}

func (a *App) Pause(params PauseParams) error {
	return a.pauseOrUnpause(wego.Suspend, params.Name, params.Namespace)
}
