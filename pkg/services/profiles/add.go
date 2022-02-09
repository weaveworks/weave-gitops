package profiles

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/models"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/google/uuid"
	"k8s.io/apimachinery/pkg/types"
)

const AddCommitMessage = "Add Profile manifests"

type AddOptions struct {
	Name         string
	Cluster      string
	ConfigRepo   string
	Version      string
	ProfilesPort string
	Namespace    string
	Kubeconfig   string
	AutoMerge    bool
}

// Add installs an available profile in a cluster's namespace by appending a HelmRelease to the profile manifest in the config repo,
// provided that such a HelmRelease does not exist with the same profile name and version in the same namespace and cluster.
func (s *ProfilesSvc) Add(ctx context.Context, gitProvider gitproviders.GitProvider, opts AddOptions) error {
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
		return fmt.Errorf("failed to get files in '%s' for config repository %q: %s", git.GetSystemPath(opts.Cluster), configRepoURL, err)
	}

	fileContent := getGitCommitFileContent(files, git.GetProfilesPath(opts.Cluster, models.WegoProfilesPath))

	content, err := addHelmRelease(helmRepo, fileContent, version, opts)
	if err != nil {
		return fmt.Errorf("failed to add HelmRelease for profile '%s' to %s: %w", opts.Name, models.WegoProfilesPath, err)
	}

	path := git.GetProfilesPath(opts.Cluster, models.WegoProfilesPath)

	pr, err := gitProvider.CreatePullRequest(ctx, configRepoURL, gitproviders.PullRequestInfo{
		Title:         fmt.Sprintf("GitOps add %s", opts.Name),
		Description:   fmt.Sprintf("Add manifest for %s profile", opts.Name),
		CommitMessage: AddCommitMessage,
		TargetBranch:  defaultBranch,
		NewBranch:     uuid.New().String(),
		Files: []gitprovider.CommitFile{{
			Path:    &path,
			Content: &content,
		}},
	})
	if err != nil {
		return fmt.Errorf("failed to create pull request: %s", err)
	}

	s.Logger.Actionf("Pull Request created: %s", pr.Get().WebURL)

	if opts.AutoMerge {
		s.Logger.Actionf("auto-merge=true; merging PR number %v", pr.Get().Number)

		if err := gitProvider.MergePullRequest(ctx, configRepoURL, pr.Get().Number, AddCommitMessage); err != nil {
			return fmt.Errorf("error auto-merging PR: %w", err)
		}
	}

	s.printAddSummary(opts)

	return nil
}

func (s *ProfilesSvc) printAddSummary(opts AddOptions) {
	s.Logger.Println("Adding profile:\n")
	s.Logger.Println("Name: %s", opts.Name)
	s.Logger.Println("Version: %s", opts.Version)
	s.Logger.Println("Cluster: %s", opts.Cluster)
	s.Logger.Println("Namespace: %s\n", opts.Namespace)
}

func addHelmRelease(helmRepo types.NamespacedName, fileContent, version string, opts AddOptions) (string, error) {
	newRelease := helm.MakeHelmRelease(opts.Name, version, opts.Cluster, opts.Namespace, helmRepo)

	matchingHelmReleases, err := helm.FindHelmReleaseInString(fileContent, newRelease)
	if len(matchingHelmReleases) >= 1 {
		return "", fmt.Errorf("profile '%s' is already installed in %s/%s", opts.Name, opts.Namespace, opts.Cluster)
	} else if err != nil {
		return "", fmt.Errorf("error reading from %s: %w", models.WegoProfilesPath, err)
	}

	return helm.AppendHelmReleaseToString(fileContent, newRelease)
}
