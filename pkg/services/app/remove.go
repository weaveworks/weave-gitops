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
func (a *App) Remove(configGit git.Git, gitProvider gitproviders.GitProvider, params RemoveParams) error {
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
	info, err := getAppResourceInfo(*application, clusterName)
	if err != nil {
		return err
	}

	resources := info.clusterResources()

	if info.configMode() == ConfigModeClusterOnly {
		gvrApp, err := ResourceKindApplication.ToGVR()
		if err != nil {
			return err
		}

		if err := a.Kube.DeleteByName(ctx, info.appResourceName(), gvrApp, info.Namespace); err != nil {
			return clusterDeleteError(info.appResourceName(), err)
		}

		gvrSource, err := info.sourceKind().ToGVR()
		if err != nil {
			return err
		}

		if err := a.Kube.DeleteByName(ctx, info.appSourceName(), gvrSource, info.Namespace); err != nil {
			return clusterDeleteError(info.appResourceName(), err)
		}

		gvrDeployKind, err := info.deployKind().ToGVR()
		if err != nil {
			return err
		}

		if err := a.Kube.DeleteByName(ctx, info.appDeployName(), gvrDeployKind, info.Namespace); err != nil {
			return clusterDeleteError(info.appResourceName(), err)
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
			} else if resourceRef.kind == ResourceKindKustomization ||
				resourceRef.kind == ResourceKindHelmRelease {
				gvrDeployKind, err := resourceRef.kind.ToGVR()
				if err != nil {
					return err
				}
				if err := a.Kube.DeleteByName(ctx, resourceRef.name, gvrDeployKind, info.Namespace); err != nil {
					return clusterDeleteError(info.appResourceName(), err)
				}
			}
		}

		return a.commitAndPush(configGit, RemoveCommitMessage, params.DryRun)
	}

	return nil
}

func (a *App) getConfigUrlAndBranch(ctx context.Context, gitProvider gitproviders.GitProvider, info *AppResourceInfo) (gitproviders.RepoURL, string, error) {
	configUrl := info.Spec.ConfigURL
	if configUrl == string(ConfigTypeUserRepo) {
		configUrl = info.Spec.URL
	}

	repoUrl, err := gitproviders.NewRepoURL(configUrl)
	if err != nil {
		return gitproviders.RepoURL{}, "", err
	}

	branch := info.Spec.Branch

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
