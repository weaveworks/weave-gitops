package app

import wego "github.com/weaveworks/weave-gitops/api/v1alpha1"

type PauseParams struct {
	Name      string
	Namespace string
}

func (a *AppSvc) Pause(params PauseParams) error {
	return a.pauseOrUnpause(wego.SuspendAction, params.Name, params.Namespace)
}
