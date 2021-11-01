package app

import (
	"context"
	"errors"
	"fmt"
	"time"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/fluxcd/pkg/apis/meta"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
)

const (
	k8sPollInterval = 2 * time.Second
	k8sTimeout      = 1 * time.Minute
)

type SyncParams struct {
	Name      string
	Namespace string
}

// Sync triggers the reconcile loop for an application
func (a *AppSvc) Sync(params SyncParams) error {
	ctx := context.Background()

	app, err := a.Kube.GetApplication(ctx, types.NamespacedName{Namespace: params.Namespace, Name: params.Name})
	if err != nil {
		return fmt.Errorf("failed getting application: %w", err)
	}

	if err := a.syncSource(ctx, app); err != nil {
		return fmt.Errorf("failed sync source: %w", err)
	}

	if err := a.syncDeployment(ctx, app); err != nil {
		return fmt.Errorf("failed sync deployment: %w", err)
	}

	return nil
}

func (a *AppSvc) syncSource(ctx context.Context, app *wego.Application) error {
	var source kube.Resource

	switch app.Spec.SourceType {
	case wego.SourceTypeGit:
		source = &sourcev1.GitRepository{}
	case wego.SourceTypeHelm:
		source = &sourcev1.HelmRepository{}
	}

	return a.syncResource(ctx, app, source)
}

func (a *AppSvc) syncDeployment(ctx context.Context, app *wego.Application) error {
	var deploy kube.Resource

	switch app.Spec.DeploymentType {
	case wego.DeploymentTypeKustomize:
		deploy = &kustomizev1.Kustomization{}
	case wego.DeploymentTypeHelm:
		deploy = &helmv2.HelmRelease{}
	}

	return a.syncResource(ctx, app, deploy)
}

func (a *AppSvc) syncResource(ctx context.Context, app *wego.Application, resource kube.Resource) error {
	name := types.NamespacedName{
		Name:      app.Name,
		Namespace: app.Namespace,
	}

	if err := a.Kube.GetResource(ctx, name, resource); err != nil {
		return err
	}

	a.setReconcileAnnotations(resource)

	if err := a.Kube.SetResource(ctx, resource); err != nil {
		return err
	}

	if err := utils.Poll(
		a.Clock,
		k8sPollInterval,
		k8sTimeout,
		a.checkResourceSync(ctx, name, resource),
	); err != nil {
		return err
	}

	return nil
}

func (a *AppSvc) setReconcileAnnotations(resource kube.Resource) {
	annotations := resource.GetAnnotations()

	if annotations == nil {
		annotations = map[string]string{
			meta.ReconcileRequestAnnotation: a.Clock.Now().Format(time.RFC3339Nano),
		}
	} else {
		annotations[meta.ReconcileRequestAnnotation] = a.Clock.Now().Format(time.RFC3339Nano)
	}

	resource.SetAnnotations(annotations)
}

func (a *AppSvc) checkResourceSync(ctx context.Context, name types.NamespacedName, resource kube.Resource) func() (bool, error) {
	reconcileAtBeforeUpdate, err := getLastHandledReconcileRequest(resource)
	if err != nil {
		return func() (bool, error) { return false, fmt.Errorf("error getting reconcile at before update: %w", err) }
	}

	return func() (bool, error) {
		updatedResource, err := initResourceType(resource)
		if err != nil {
			return false, err
		}

		err = a.Kube.GetResource(ctx, name, updatedResource)
		if err != nil {
			return false, err
		}

		lastReconcile, err := getLastHandledReconcileRequest(updatedResource)
		if err != nil {
			return false, fmt.Errorf("error getting reconcile at after update: %w", err)
		}

		return lastReconcile != reconcileAtBeforeUpdate, nil
	}
}

func initResourceType(resource kube.Resource) (kube.Resource, error) {
	switch resource.(type) {
	case *sourcev1.GitRepository:
		return &sourcev1.GitRepository{}, nil
	case *sourcev1.HelmRepository:
		return &sourcev1.HelmRepository{}, nil
	case *kustomizev1.Kustomization:
		return &kustomizev1.Kustomization{}, nil
	case *helmv2.HelmRelease:
		return &helmv2.HelmRelease{}, nil
	}

	return nil, errors.New("invalid resource")
}

func getLastHandledReconcileRequest(resource kube.Resource) (string, error) {
	switch r := resource.(type) {
	case *sourcev1.GitRepository:
		return r.Status.GetLastHandledReconcileRequest(), nil
	case *sourcev1.HelmRepository:
		return r.Status.GetLastHandledReconcileRequest(), nil
	case *kustomizev1.Kustomization:
		return r.Status.GetLastHandledReconcileRequest(), nil
	case *helmv2.HelmRelease:
		return r.Status.GetLastHandledReconcileRequest(), nil
	}

	return "", errors.New("invalid resource")
}
