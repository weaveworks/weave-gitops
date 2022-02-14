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

const UpdateCommitMessage = "Update Profile manifests"

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

	opts.Version = version

	files, err := gitProvider.GetRepoDirFiles(ctx, configRepoURL, git.GetSystemPath(opts.Cluster), defaultBranch)
	if err != nil {
		return fmt.Errorf("failed to get files in '%s' of config repository %q: %s", git.GetSystemPath(opts.Cluster), configRepoURL, err)
	}

	fileContent := getGitCommitFileContent(files, git.GetProfilesPath(opts.Cluster, models.WegoProfilesPath))
	if fileContent == "" {
		return fmt.Errorf("failed to find installed profiles in '%s' of config repo %q", git.GetProfilesPath(opts.Cluster, models.WegoProfilesPath), configRepoURL)
	}

	content, err := updateHelmRelease(helmRepo, fileContent, opts.Name, opts.Version, opts.Cluster, opts.Namespace)
	if err != nil {
		return fmt.Errorf("failed to update HelmRelease for profile '%s' in %s: %w", opts.Name, models.WegoProfilesPath, err)
	}

	path := git.GetProfilesPath(opts.Cluster, models.WegoProfilesPath)

	pr, err := gitProvider.CreatePullRequest(ctx, configRepoURL, gitproviders.PullRequestInfo{
		Title:         fmt.Sprintf("GitOps update %s", opts.Name),
		Description:   fmt.Sprintf("Update manifest for %s profile", opts.Name),
		CommitMessage: UpdateCommitMessage,
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

	s.Logger.Actionf("created Pull Request: %s", pr.Get().WebURL)

	if opts.AutoMerge {
		s.Logger.Actionf("auto-merge=true; merging PR number %v", pr.Get().Number)

		if err := gitProvider.MergePullRequest(ctx, configRepoURL, pr.Get().Number, AddCommitMessage); err != nil {
			return fmt.Errorf("error auto-merging PR: %w", err)
		}
	}

	s.printUpdateSummary(opts)

	return nil
}

func (s *ProfilesSvc) printUpdateSummary(opts UpdateOptions) {
	s.Logger.Println("Updating profile:\n")
	s.Logger.Println("Name: %s", opts.Name)
	s.Logger.Println("Version: %s", opts.Version)
	s.Logger.Println("Cluster: %s", opts.Cluster)
	s.Logger.Println("Namespace: %s\n", opts.Namespace)
}

func updateHelmRelease(helmRepo types.NamespacedName, fileContent, name, version, cluster, ns string) (string, error) {
	existingReleases, err := helm.SplitHelmReleaseYAML([]byte(fileContent))
	if err != nil {
		return "", fmt.Errorf("error splitting into YAML: %w", err)
	}

	releaseName := cluster + "-" + name

	matchingHelmRelease, index, err := helm.FindReleaseInNamespace(existingReleases, releaseName, ns)
	if matchingHelmRelease == nil {
		return "", fmt.Errorf("failed to find HelmRelease '%s' in namespace %s", releaseName, ns)
	} else if err != nil {
		return "", fmt.Errorf("error reading from %s: %w", models.WegoProfilesPath, err)
	}

	if matchingHelmRelease.Spec.Chart.Spec.Version == version {
		return "", fmt.Errorf("version %s of profile '%s' already installed in %s/%s", version, name, ns, cluster)
	}

	matchingHelmRelease.Spec.Chart.Spec.Version = version

	return helm.PatchHelmRelease(existingReleases, *matchingHelmRelease, index)
}
