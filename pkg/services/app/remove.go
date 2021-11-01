package app

import (
	"context"

	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"github.com/weaveworks/weave-gitops/pkg/services/gitopswriter"
	"github.com/weaveworks/weave-gitops/pkg/services/gitrepo"
	"k8s.io/apimachinery/pkg/types"
)

type RemoveParams struct {
	Name             string
	Namespace        string
	DryRun           bool
	GitProviderToken string
}

// Remove removes the Weave GitOps automation for an application
func (a *AppSvc) Remove(configGit git.Git, gitProvider gitproviders.GitProvider, params RemoveParams) error {
	if params.DryRun {
		return nil
	}

	ctx := context.Background()

	clusterName, err := a.Kube.GetClusterName(ctx)
	if err != nil {
		return err
	}

	application, err := a.Kube.GetApplication(ctx, types.NamespacedName{Namespace: params.Namespace, Name: params.Name})
	if err != nil {
		return err
	}

	// Find all resources created when adding this app
	app, err := automation.WegoAppToApp(*application)
	if err != nil {
		return err
	}

	return a.removeApp(ctx, app, clusterName, true)
}

func (a *AppSvc) removeApp(ctx context.Context, app models.Application, clusterName string, autoMerge bool) error {
	repoWriter := gitrepo.NewRepoWriter(app.ConfigURL, a.GitProvider, a.ConfigGit, a.Logger)
	automationSvc := automation.NewAutomationService(a.GitProvider, a.Flux, a.Logger)
	gitOpsDirWriter := gitopswriter.NewGitOpsDirectoryWriter(automationSvc, repoWriter, a.Logger)

	return gitOpsDirWriter.RemoveApplication(ctx, app, clusterName, autoMerge)
}
