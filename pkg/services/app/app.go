package app

import (
	"context"
	"fmt"

	"github.com/fluxcd/go-git-providers/gitprovider"
	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta1"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// type DeploymentType string
// type SourceType string

// const (
//  DeployTypeKustomize DeploymentType = "kustomize"
//  DeployTypeHelm      DeploymentType = "helm"

//  SourceTypeGit  SourceType = "git"
//  SourceTypeHelm SourceType = "helm"
// )

// AppService entity that manages applications
type AppService interface {
	// Add adds a new application to the cluster
	Add(params AddParams) error
	// Get returns a given applicaiton
	Get(name types.NamespacedName) (*wego.Application, error)
	// GetCommits returns a list of commits for an application
	GetCommits(params CommitParams, application *wego.Application) ([]gitprovider.Commit, error)
	// Remove removes an application from the cluster
	Remove(params RemoveParams) error
	// Status returns flux resources status and the last successful reconciliation time
	Status(params StatusParams) (string, string, error)
	// Pause pauses the gitops automation for an app
	Pause(params PauseParams) error
	// Unpause resumes the gitops automation for an app
	Unpause(params UnpauseParams) error
}

type App struct {
	Context     context.Context
	Osys        osys.Osys
	AppGit      git.Git
	ConfigGit   git.Git
	Flux        flux.Flux
	Kube        kube.Kube
	Logger      logger.Logger
	GitProvider gitproviders.GitProvider
}

func New(ctx context.Context, logger logger.Logger, appGit, configGit git.Git, gitProvider gitproviders.GitProvider, flux flux.Flux, kube kube.Kube, osys osys.Osys) AppService {
	return &App{
		Context:     ctx,
		AppGit:      appGit,
		ConfigGit:   configGit,
		Flux:        flux,
		Kube:        kube,
		Logger:      logger,
		Osys:        osys,
		GitProvider: gitProvider,
	}
}

// Make sure App implements all the required methods.
var _ AppService = &App{}

func (a *App) getDeploymentType(ctx context.Context, name string, namespace string) (wego.DeploymentType, error) {
	app, err := a.Kube.GetApplication(ctx, types.NamespacedName{Name: name, Namespace: namespace})
	if err != nil {
		return wego.DeploymentTypeKustomize, err
	}

	return wego.DeploymentType(app.Spec.DeploymentType), nil
}

func (a *App) getSuspendedStatus(ctx context.Context, name, namespace string, deploymentType wego.DeploymentType) (bool, error) {
	var automation client.Object

	switch deploymentType {
	case wego.DeploymentTypeKustomize:
		automation = &kustomizev1.Kustomization{}
	case wego.DeploymentTypeHelm:
		automation = &helmv2.HelmRelease{}
	default:
		return false, fmt.Errorf("invalid deployment type: %v", deploymentType)
	}

	if err := a.Kube.GetResource(ctx, types.NamespacedName{Namespace: namespace, Name: name}, automation); err != nil {
		return false, err
	}

	suspendStatus := false

	switch at := automation.(type) {
	case *kustomizev1.Kustomization:
		suspendStatus = at.Spec.Suspend
	case *helmv2.HelmRelease:
		suspendStatus = at.Spec.Suspend
	}

	return suspendStatus, nil
}

func (a *App) pauseOrUnpause(suspendAction wego.SuspendActionType, name, namespace string) error {
	ctx := context.Background()

	deploymentType, err := a.getDeploymentType(ctx, name, namespace)
	if err != nil {
		return fmt.Errorf("unable to determine deployment type for %s: %s", name, err)
	}

	suspendStatus, err := a.getSuspendedStatus(ctx, name, namespace, deploymentType)
	if err != nil {
		return fmt.Errorf("failed to get suspended status: %s", err)
	}

	switch deploymentType {
	case wego.DeploymentTypeKustomize:
		deploymentType = "kustomization"
	case wego.DeploymentTypeHelm:
		deploymentType = "helmrelease"
	default:
		return fmt.Errorf("invalid deployment type: %v", deploymentType)
	}

	switch suspendAction {
	case wego.SuspendAction:
		if suspendStatus {
			a.Logger.Printf("app %s is already paused\n", name)
			return nil
		}

		out, err := a.Flux.SuspendOrResumeApp(suspendAction, name, namespace, string(deploymentType))
		if err != nil {
			return fmt.Errorf("unable to pause %s err: %s", name, err)
		}

		a.Logger.Printf("%s\n gitops automation paused for %s\n", string(out), name)

		return nil
	case wego.ResumeAction:
		if !suspendStatus {
			a.Logger.Printf("app %s is already reconciling\n", name)
			return nil
		}

		out, err := a.Flux.SuspendOrResumeApp(suspendAction, name, namespace, string(deploymentType))
		if err != nil {
			return fmt.Errorf("unable to unpause %s err: %s", name, err)
		}

		a.Logger.Printf("%s\n gitops automation unpaused for %s\n", string(out), name)

		return nil
	}

	return fmt.Errorf("invalid suspend action")
}

func IsClusterReady(l logger.Logger, k kube.Kube) error {
	l.Waitingf("Checking cluster status")

	clusterStatus := k.GetClusterStatus(context.Background())

	switch clusterStatus {
	case kube.Unmodified:
		return fmt.Errorf("gitops not installed... exiting")
	case kube.Unknown:
		return fmt.Errorf("can not determine cluster status... exiting")
	}

	l.Successf(clusterStatus.String())

	return nil
}
