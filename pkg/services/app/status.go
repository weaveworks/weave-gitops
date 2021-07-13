package app

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
)

type StatusParams struct {
	AppName   string
	Namespace string
}

func (a *App) Status(params StatusParams) error {

	_, err := a.Get(types.NamespacedName{Name: params.AppName, Namespace: params.Namespace})
	if err != nil {
		return err
	}

	deploymentType, err := a.flux.GetDeploymentType(params.Namespace, params.AppName)
	if err != nil {
		return err
	}

	ctx := context.Background()
	latestDeploymentTime, err := a.kube.LatestSuccessfulDeploymentTime(ctx, types.NamespacedName{Namespace: params.Namespace, Name: params.AppName}, string(deploymentType))
	if err != nil {
		return err
	}
	fmt.Printf("Latest successful deployment time: %s\n", latestDeploymentTime)

	output, err := a.flux.GetAllResourcesStatus(params.AppName)
	if err != nil {
		return fmt.Errorf("error getting flux app resources status %s", err)
	}

	fmt.Println(string(output))

	return nil

}
