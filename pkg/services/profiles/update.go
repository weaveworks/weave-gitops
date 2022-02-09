package profiles

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/models"

	"k8s.io/apimachinery/pkg/types"
)

type UpdateOptions struct {
	Name         string
	Cluster      string
	ConfigRepo   string
	Version      string
	ProfilesPort string
	Namespace    string
	Kubeconfig   string
	AutoMerge    bool
}

// Update updates an installed profile
func (s *ProfilesSvc) Update(ctx context.Context, gitProvider gitproviders.GitProvider, opts UpdateOptions) error {
	configRepoURL, err := gitproviders.NewRepoURL(opts.ConfigRepo)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}

	repoExists, err := gitProvider.RepositoryExists(ctx, configRepoURL)
	if err != nil {
		return fmt.Errorf("failed to check whether repository exists: %w", err)
	} else if !repoExists {
		return fmt.Errorf("repository %q could not be found", configRepoURL)
	}

	defaultBranch, err := gitProvider.GetDefaultBranch(ctx, configRepoURL)
	if err != nil {
		return fmt.Errorf("failed to get default branch: %w", err)
	}

	helmRepo, version, err := s.discoverHelmRepository(ctx, GetOptions{
		Name:      opts.Name,
		Version:   opts.Version,
		Cluster:   opts.Cluster,
		Namespace: opts.Namespace,
		Port:      opts.ProfilesPort,
	})
	if err != nil {
		return fmt.Errorf("failed to discover HelmRepository: %w", err)
	}

	files, err := gitProvider.GetRepoDirFiles(ctx, configRepoURL, git.GetSystemPath(opts.Cluster), defaultBranch)
	if err != nil {
		return fmt.Errorf("failed to get files in '%s' of config repository %q: %s", git.GetSystemPath(opts.Cluster), configRepoURL, err)
	}

	fileContent := getGitCommitFileContent(files, git.GetProfilesPath(opts.Cluster, models.WegoProfilesPath))
	if fileContent == "" {
		return fmt.Errorf("failed to find installed profiles in '%s' of config repo %q", git.GetProfilesPath(opts.Cluster, models.WegoProfilesPath), configRepoURL)
	}

	_, err = updateHelmRelease(helmRepo, fileContent, version, opts)
	if err != nil {
		return fmt.Errorf("failed to update HelmRelease for profile '%s' in %s: %w", opts.Name, models.WegoProfilesPath, err)
	}

	return nil
}

func updateHelmRelease(helmRepo types.NamespacedName, fileContent, version string, opts UpdateOptions) (string, error) {
	newRelease := helm.MakeHelmRelease(opts.Name, version, opts.Cluster, opts.Namespace, helmRepo)

	matchingHelmReleases, err := helm.FindHelmReleaseInString(fileContent, newRelease)
	if len(matchingHelmReleases) == 0 {
		return "", fmt.Errorf("profile '%s' could not be found in %s/%s", opts.Name, opts.Namespace, opts.Cluster)
	} else if err != nil {
		return "", fmt.Errorf("error reading from %s: %w", models.WegoProfilesPath, err)
	}

	return helm.AppendHelmReleaseToString(fileContent, newRelease)
}
