package app

import (
	"context"
	"fmt"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"k8s.io/apimachinery/pkg/types"
)

const (
	RemoveCommitMessage = "Remove App manifests"
)

type RemoveParams struct {
	Name             string
	Namespace        string
	DryRun           bool
	GitProviderToken string
}

// Remove removes the Weave GitOps automation for an application
func (a *AppSvc) Remove(configGit git.Git, gitProvider gitproviders.GitProvider, params RemoveParams) error {
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
	app, err := models.NewApplication(*application)
	if err != nil {
		return err
	}

	resources := app.ClusterResources(clusterName)

	if app.ConfigMode() == models.ConfigModeClusterOnly {
		gvrApp, err := models.ResourceKindApplication.ToGVR()
		if err != nil {
			return err
		}

		if err := a.Kube.DeleteByName(ctx, app.AppResourceName(), gvrApp, app.Namespace); err != nil {
			return clusterDeleteError(app.AppResourceName(), err)
		}

		gvrSource, err := app.SourceKind().ToGVR()
		if err != nil {
			return err
		}

		if err := a.Kube.DeleteByName(ctx, app.AppSourceName(), gvrSource, app.Namespace); err != nil {
			return clusterDeleteError(app.AppResourceName(), err)
		}

		gvrDeployKind, err := app.DeployKind().ToGVR()
		if err != nil {
			return err
		}

		if err := a.Kube.DeleteByName(ctx, app.AppDeployName(), gvrDeployKind, app.Namespace); err != nil {
			return clusterDeleteError(app.AppResourceName(), err)
		}

		return nil
	}

	cloneURL, branch, err := a.getConfigUrlAndBranch(ctx, gitProvider, info)
	if err != nil {
		return fmt.Errorf("failed to obtain config URL and branch: %w", err)
	}

	remover, _, err := a.cloneRepo(configGit, cloneURL.String(), branch, params.DryRun)
	if err != nil {
		return fmt.Errorf("failed to clone configuration repo: %w", err)
	}

	defer remover()

	a.Logger.Actionf("Removing application from cluster and repository")

	if !params.DryRun {
		for _, resourceRef := range resources {
			if resourceRef.repositoryPath != "" { // Some of the automation doesn't get stored
				if err := configGit.Remove(resourceRef.repositoryPath); err != nil {
					return err
				}
			} else if resourceRef.Kind == models.ResourceKindKustomization ||
				resourceRef.Kind == models.ResourceKindHelmRelease {
				gvrDeployKind, err := resourceRef.Kind.ToGVR()
				if err != nil {
					return err
				}
				if err := a.Kube.DeleteByName(ctx, resourceRef.Name, gvrDeployKind, app.Namespace); err != nil {
					return clusterDeleteError(app.AppResourceName(), err)
				}
			}
		}

		return a.commitAndPush(configGit, RemoveCommitMessage, params.DryRun)
	}

	return nil
}

func (a *App) getConfigUrlAndBranch(ctx context.Context, gitProvider gitproviders.GitProvider, info *AppResourceInfo) (gitproviders.RepoURL, string, error) {
	configUrl := info.Spec.ConfigURL
	if configUrl == string(models.ConfigTypeUserRepo) {
		configUrl = info.Spec.URL
	}

	repoUrl, err := gitproviders.NewRepoURL(configUrl)
	if err != nil {
		return gitproviders.RepoURL{}, "", err
	}

	branch := app.Spec.Branch

	if branch != "" {
		branch, err = gitProvider.GetDefaultBranch(ctx, repoUrl)
		if err != nil {
			return gitproviders.RepoURL{}, "", err
		}
	}

	return repoUrl, branch, nil
}

func clusterDeleteError(name string, err error) error {
	return fmt.Errorf("failed to delete resource: %s with error: %w", name, err)
}

func dirExists(d string) bool {
	info, err := os.Stat(d)
	if os.IsNotExist(err) {
		return false
	}

	return info.IsDir()
}

// findAppManifests locates all manifests associated with a specific application within a repository;
// only used when "RemoveWorkload" is specified
func findAppManifests(application wego.Application, repoDir string) ([][]byte, error) {
	root := filepath.Join(repoDir, application.Spec.Path)
	if !dirExists(root) {
		return nil, fmt.Errorf("application path '%s' not found", application.Spec.Path)
	}

	manifests := [][]byte{}

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.IsDir() && filepath.Ext(path) == ".yaml" {
			manifest, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			manifests = append(manifests, manifest)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return manifests, nil
}
