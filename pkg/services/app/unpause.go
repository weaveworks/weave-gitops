package app

import (
	"context"
	"fmt"
)

type UnpauseParams struct {
	Name           string
	Namespace      string
	DeploymentType string
}

func (a *App) Unpause(params UnpauseParams) error {
	ctx := context.Background()
	deploymentType, err := a.getDeploymentType(ctx, params.Name, params.Namespace)
	if err != nil {
		return fmt.Errorf("unable to determine deployment type: %s", err)
	}

	switch deploymentType {
	case DeployTypeKustomize:
		deploymentType = "kustomization"
	case DeployTypeHelm:
		deploymentType = "helmrelease"
	default:
		return fmt.Errorf("invalid deployment type: %v", deploymentType)
	}

	_, err = a.flux.ResumeApp(params.Name, params.Namespace, string(deploymentType))
	if err != nil {
		return fmt.Errorf("unable to resume %s err: %s", params.Name, err)
	}
	a.logger.Printf("gitops automation resumed for %s\n", params.Name)
	return nil
}
