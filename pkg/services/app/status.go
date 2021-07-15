package app

import (
	"context"
	"fmt"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	"github.com/fluxcd/pkg/apis/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type StatusParams struct {
	Namespace string
	Name      string
}

func (a *App) Status(params StatusParams) (string, string, error) {
	fluxOutput, err := a.flux.GetAllResourcesStatus(params.Name, params.Namespace)
	if err != nil {
		return "", "", fmt.Errorf("failed getting app status: %w", err)
	}

	ctx := context.Background()
	deploymentType, err := a.getDeploymentType(ctx, params)
	if err != nil {
		return "", "", fmt.Errorf("failed getting app deployment type: %w", err)
	}

	lastRecon, err := a.getLastSuccessfulReconciliation(ctx, deploymentType, params)
	if err != nil {
		return "", "", fmt.Errorf("failed getting last successful reconciliation: %w", err)
	}

	return string(fluxOutput), lastRecon, nil
}

func (a *App) getDeploymentType(ctx context.Context, params StatusParams) (DeploymentType, error) {
	app, err := a.kube.GetApplication(ctx, types.NamespacedName{Name: params.Name, Namespace: params.Namespace})
	if err != nil {
		return DeployTypeKustomize, err
	}

	return DeploymentType(app.Spec.DeploymentType), nil
}

func (a *App) getLastSuccessfulReconciliation(ctx context.Context, deploymentType DeploymentType, params StatusParams) (string, error) {
	conditions := []metav1.Condition{}
	switch deploymentType {
	case DeployTypeKustomize:
		kust := &kustomizev1.Kustomization{}
		if err := a.kube.GetResource(ctx, types.NamespacedName{Name: params.Name, Namespace: params.Namespace}, kust); err != nil {
			return "", fmt.Errorf("failed getting resource: %w", err)
		}
		conditions = kust.Status.Conditions
	case DeployTypeHelm:
		helm := &helmv2.HelmRelease{}
		if err := a.kube.GetResource(ctx, types.NamespacedName{Name: params.Name, Namespace: params.Namespace}, helm); err != nil {
			return "", fmt.Errorf("failed getting resource: %w", err)
		}
		conditions = helm.Status.Conditions
	}

	for _, c := range conditions {
		if c.Type == meta.ReadyCondition && c.Status == metav1.ConditionTrue {
			return c.LastTransitionTime.String(), nil
		}
	}

	return "No succesfull reconciliation", nil
}
