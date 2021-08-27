package app

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
)

type RemoveParams struct {
	Name             string
	Namespace        string
	PrivateKey       string
	DryRun           bool
	GitProviderToken string
}

// Remove removes the Weave GitOps automation for an application
func (a *App) Remove(params RemoveParams, application *wego.Application) error {
	ctx := context.Background()

	clusterName, err := a.Kube.GetClusterName(ctx)
	if err != nil {
		return err
	}

	// Find all resources created when adding this app
	info := getAppResourceInfo(*application, clusterName)
	resources := info.clusterResources()

	if info.configMode() == ConfigModeClusterOnly {
		out, err := a.Kube.DeleteByName(info.appResourceName(), "app", info.Namespace)
		if err != nil {
			return clusterDeleteError(out, err)
		}
		out, err = a.Kube.DeleteByName(info.appSourceName(), string(info.sourceKind()), info.Namespace)
		if err != nil {
			return clusterDeleteError(out, err)
		}
		out, err = a.Kube.DeleteByName(info.appDeployName(), string(info.deployKind()), info.Namespace)
		if err != nil {
			return clusterDeleteError(out, err)
		}
		return nil
	}

	cloneURL, branch, err := a.getConfigUrlAndBranch(info, params.GitProviderToken)
	if err != nil {
		return fmt.Errorf("failed to obtain config URL and branch: %w", err)
	}

	remover, err := a.cloneRepo(cloneURL, branch, params.DryRun)

	if err != nil {
		return fmt.Errorf("failed to clone configuration repo: %w", err)
	}

	defer remover()

	a.Logger.Actionf("Removing application from cluster and repository")

	if !params.DryRun {
		for _, resourceRef := range resources {
			if resourceRef.repositoryPath != "" { // Some of the automation doesn't get stored
				if err := a.Git.Remove(resourceRef.repositoryPath); err != nil {
					return err
				}
			} else if resourceRef.kind == ResourceKindKustomization ||
				resourceRef.kind == ResourceKindHelmRelease {
				out, err := a.Kube.DeleteByName(resourceRef.name, string(resourceRef.kind), info.Namespace)
				if err != nil {
					return clusterDeleteError(out, err)
				}
			}
		}

		if err := a.commitAndPush(); err != nil {
			return err
		}
	}

	return nil
}

func (a *App) getConfigUrlAndBranch(info *AppResourceInfo, token string) (string, string, error) {
	cloneURL := info.Spec.ConfigURL
	branch := info.Spec.Branch

	if cloneURL == string(ConfigTypeUserRepo) {
		cloneURL = info.Spec.URL
	} else {
		gitProvider, err := a.GitProviderFactory(token)
		if err != nil {
			return "", "", err
		}

		branch, err = gitProvider.GetDefaultBranch(cloneURL)
		if err != nil {
			return "", "", err
		}
	}

	return cloneURL, branch, nil
}

func clusterDeleteError(out []byte, err error) error {
	return fmt.Errorf("failed to delete resource: %s with error: %w", out, err)
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
