package app

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
)

type PauseParams struct {
	Name           string
	Namespace      string
	DeploymentType string
}

func (a *App) Pause(params PauseParams) error {
	switch params.DeploymentType {
	case string(DeployTypeKustomize):
		params.DeploymentType = "kustomization"
	case string(DeployTypeHelm):
		params.DeploymentType = "helmrelease"
	default:
		return fmt.Errorf("invalid deployment type: %v", params.DeploymentType)
	}

	_, err := fluxops.CallFlux("suspend", params.DeploymentType, params.Name, fmt.Sprintf("--namespace=%s", params.Namespace))
	if err != nil {
		return fmt.Errorf("unable to pause %s err: %s", params.Name, err)
	}
	a.logger.Printf("gitops automation paused for %s", params.Name)
	return nil
}
