package app

import (
	"context"
	"fmt"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	"github.com/fluxcd/pkg/apis/meta"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

type StatusParams struct {
	Namespace string
	Name      string
}

func (a *App) Status(params StatusParams) (string, string, error) {
	fluxOutput, err := a.Flux.GetAllResourcesStatus(params.Name, params.Namespace)
	if err != nil {
		return "", "", fmt.Errorf("failed getting app status: %w", err)
	}

	ctx := context.Background()

	deploymentType, err := a.getDeploymentType(ctx, params.Name, params.Namespace)
	if err != nil {
		return "", "", fmt.Errorf("failed getting app deployment type: %w", err)
	}

	lastRecon, err := a.getLastSuccessfulReconciliation(ctx, deploymentType, params)
	if err != nil {
		return "", "", fmt.Errorf("failed getting last successful reconciliation: %w", err)
	}

	return string(fluxOutput), lastRecon, nil
}

func (a *App) getLastSuccessfulReconciliation(ctx context.Context, deploymentType wego.DeploymentType, params StatusParams) (string, error) {
	conditions := []metav1.Condition{}

	switch deploymentType {
	case wego.DeploymentTypeKustomize:
		kust := &kustomizev1.Kustomization{}
		if err := a.Kube.GetResource(ctx, types.NamespacedName{Name: params.Name, Namespace: params.Namespace}, kust); err != nil {
			return "", fmt.Errorf("failed getting resource: %w", err)
		}

		conditions = kust.Status.Conditions
	case wego.DeploymentTypeHelm:
		helm := &helmv2.HelmRelease{}
		if err := a.Kube.GetResource(ctx, types.NamespacedName{Name: params.Name, Namespace: params.Namespace}, helm); err != nil {
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
