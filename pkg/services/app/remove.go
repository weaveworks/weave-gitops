package app

import (
	"context"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"

	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"k8s.io/apimachinery/pkg/types"
)

type RemoveParams struct {
	Name      string
	Namespace string
	//    RemoveWorkload bool
	PrivateKey       string
	DryRun           bool
	AutoMerge        bool
	GitProviderToken string
}

// Remove removes the Weave GitOps automation for an application
func (a *App) Remove(params RemoveParams) error {
	ctx := context.Background()

	clusterName, err := a.kube.GetClusterName(ctx)
	if err != nil {
		return err
	}

	application, err := a.kube.GetApplication(ctx, types.NamespacedName{Namespace: params.Namespace, Name: params.Name})
	if err != nil {
		return err
	}

	// Find all resources created when adding this app
	info := getAppResourceInfo(*application, clusterName)
	resources := info.clusterResources()

	cloneURL := info.Spec.ConfigURL

	if cloneURL == string(ConfigTypeUserRepo) {
		cloneURL = info.Spec.URL
	}

	remover, err := a.cloneRepo(cloneURL, info.Spec.Branch, params.DryRun)
	if err != nil {
		return fmt.Errorf("failed to clone configuration repo: %w", err)
	}

	defer remover()

	if !params.DryRun {
		if params.AutoMerge {
			a.logger.Actionf("Removing manifests from disk")

			for _, resourceRef := range resources {
				if resourceRef.repositoryPath != "" { // Some of the automation doesn't get stored
					if err := a.git.Remove(resourceRef.repositoryPath); err != nil {
						return err
					}
				} else if resourceRef.kind == ResourceKindKustomization ||
					resourceRef.kind == ResourceKindHelmRelease {
					out, err := a.kube.DeleteByName(resourceRef.name, string(resourceRef.kind), info.Namespace)
					if err != nil {
						return fmt.Errorf("failed to delete resource: %s with error: %w", out, err)
					}
				}
			}

			if err := a.commitAndPush(); err != nil {
				return err
			}
		}
	}

	return nil
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
