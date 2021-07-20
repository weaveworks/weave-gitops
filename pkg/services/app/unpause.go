package app

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
)

type UnpauseParams struct {
	Name           string
	Namespace      string
	DeploymentType string
}

func (a *App) Unpause(params UnpauseParams) error {
	switch params.DeploymentType {
	case string(DeployTypeKustomize):
		params.DeploymentType = "kustomization"
	case string(DeployTypeHelm):
		params.DeploymentType = "helmrelease"
	default:
		return fmt.Errorf("invalid deployment type: %v", params.DeploymentType)
	}

	_, err := fluxops.CallFlux("resume", params.DeploymentType, params.Name, fmt.Sprintf("--namespace=%s", params.Namespace))
	if err != nil {
		return fmt.Errorf("unable to unpause %s err: %s", params.Name, err)
	}
	a.logger.Printf("gitops automation unpaused for %s", params.Name)
	return nil
}
