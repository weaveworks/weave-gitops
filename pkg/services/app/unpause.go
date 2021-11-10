package app

import wego "github.com/weaveworks/weave-gitops/api/v1alpha1"

type UnpauseParams struct {
	Name      string
	Namespace string
}

func (a *AppSvc) Unpause(params UnpauseParams) error {
	return a.pauseOrUnpause(wego.ResumeAction, params.Name, params.Namespace)
}
