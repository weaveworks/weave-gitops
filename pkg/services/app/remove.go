package app

import (
	"context"
	"fmt"
	"github.com/weaveworks/weave-gitops/pkg/services/app/internal"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

type RemoveParams struct {
	Name             string
	Namespace        string
	PrivateKey       string
	DryRun           bool
	GitProviderToken string
}

// Remove removes the Weave GitOps automation for an application
func (a *App) Remove(params RemoveParams) error {
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
	info := internal.NewResourceInfo(*application, clusterName)
	resources := info.ClusterResources()

	if info.ConfigMode() == internal.ConfigModeClusterOnly {
		gvrApp, err := internal.ResourceKindApplication.ToGVR()
		if err != nil {
			return err
		}

		if err := a.Kube.DeleteByName(ctx, info.AppResourceName(), gvrApp, info.Namespace); err != nil {
			return clusterDeleteError(info.AppResourceName(), err)
		}

		gvrSource, err := info.SourceKind().ToGVR()
		if err != nil {
			return err
		}

		if err := a.Kube.DeleteByName(ctx, info.AppSourceName(), gvrSource, info.Namespace); err != nil {
			return clusterDeleteError(info.AppResourceName(), err)
		}

		gvrDeployKind, err := info.DeployKind().ToGVR()
		if err != nil {
			return err
		}

		if err := a.Kube.DeleteByName(ctx, info.AppDeployName(), gvrDeployKind, info.Namespace); err != nil {
			return clusterDeleteError(info.AppResourceName(), err)
		}

		return nil
	}

	cloneURL, branch, err := a.getConfigUrlAndBranch(info)
	if err != nil {
		return fmt.Errorf("failed to obtain config URL and branch: %w", err)
	}

	remover, err := a.cloneRepo(a.ConfigGit, cloneURL, branch, params.DryRun)

	if err != nil {
		return fmt.Errorf("failed to clone configuration repo: %w", err)
	}

	defer remover()

	a.Logger.Actionf("Removing application from cluster and repository")

	if !params.DryRun {
		for _, resourceRef := range resources {
			if resourceRef.RepositoryPath != "" { // Some of the automation doesn't get stored
				if err := a.ConfigGit.Remove(resourceRef.RepositoryPath); err != nil {
					return err
				}
			} else if resourceRef.Kind == internal.ResourceKindKustomization ||
				resourceRef.Kind == internal.ResourceKindHelmRelease {
				gvrDeployKind, err := resourceRef.Kind.ToGVR()
				if err != nil {
					return err
				}
				if err := a.Kube.DeleteByName(ctx, resourceRef.Name, gvrDeployKind, info.Namespace); err != nil {
					return clusterDeleteError(info.AppResourceName(), err)
				}
			}
		}

		if err := a.commitAndPush(a.ConfigGit); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) getConfigUrlAndBranch(info *internal.AppResourceInfo) (string, string, error) {
	cloneURL := info.Spec.ConfigURL
	branch := info.Spec.Branch

	if cloneURL == string(internal.ConfigTypeUserRepo) {
		cloneURL = info.Spec.URL
	} else {
		localBranch, err := a.GitProvider.GetDefaultBranch(cloneURL)
		if err != nil {
			return "", "", err
		}

		branch = localBranch
	}

	return cloneURL, branch, nil
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
