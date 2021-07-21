package app

import (
	"context"
	"fmt"
)

type PauseParams struct {
	Name      string
	Namespace string
}

func (a *App) Pause(params PauseParams) error {
	ctx := context.Background()
	deploymentType, err := a.getDeploymentType(ctx, params.Name, params.Namespace)
	if err != nil {
		return fmt.Errorf("unable to determine deployment type for %s: %s", params.Name, err)
	}

	suspendStatus, err := a.getSuspendedStatus(ctx, params.Name, params.Namespace, deploymentType)
	if err != nil {
		return fmt.Errorf("failed to get suspended status: %s", err)
	}

	if suspendStatus {
		a.logger.Printf("app %s is already paused\n", params.Name)
		return nil
	}

	switch deploymentType {
	case DeployTypeKustomize:
		deploymentType = "kustomization"
	case DeployTypeHelm:
		deploymentType = "helmrelease"
	default:
		return fmt.Errorf("invalid deployment type: %v", deploymentType)
	}

	out, err := a.flux.SuspendApp(params.Name, params.Namespace, string(deploymentType))
	if err != nil {
		return fmt.Errorf("unable to pause %s err: %s", params.Name, err)
	}
	a.logger.Printf("%s\n gitops automation paused for %s\n", string(out), params.Name)
	return nil
}
