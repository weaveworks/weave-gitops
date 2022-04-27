package profiles

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/gitops/pkg/gitproviders"

	"github.com/fluxcd/go-git-providers/gitprovider"
	helmv2beta1 "github.com/fluxcd/helm-controller/api/v2beta1"
)

const UpdateCommitMessage = "Update profile manifests"

// Update updates an installed profile
func (s *ProfilesSvc) Update(ctx context.Context, gitProvider gitproviders.GitProvider, opts Options) error {
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

	_, version, err := s.discoverHelmRepository(ctx, GetOptions{
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

	content, err := updateHelmRelease(files, opts.Name, opts.Version, opts.Cluster, opts.Namespace)
	if err != nil {
		return fmt.Errorf("failed to update HelmRelease for profile '%s' in %s: %w", opts.Name, ManifestFileName, err)
	}

	path := git.GetProfilesPath(opts.Cluster, ManifestFileName)

	pr, err := gitProvider.CreatePullRequest(ctx, configRepoURL, prInfo(opts, "update", defaultBranch, gitprovider.CommitFile{
		Path:    &path,
		Content: &content,
	}))
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

func (s *ProfilesSvc) printUpdateSummary(opts Options) {
	s.Logger.Println("Updating profile:\n")
	s.Logger.Println("Name: %s", opts.Name)
	s.Logger.Println("Version: %s", opts.Version)
	s.Logger.Println("Cluster: %s", opts.Cluster)
	s.Logger.Println("Namespace: %s\n", opts.Namespace)
}

func updateHelmRelease(files []*gitprovider.CommitFile, name, version, cluster, ns string) (string, error) {
	fileContent := getGitCommitFileContent(files, git.GetProfilesPath(cluster, ManifestFileName))
	if fileContent == "" {
		return "", fmt.Errorf("failed to find installed profiles in '%s'", git.GetProfilesPath(cluster, ManifestFileName))
	}

	existingReleases, err := SplitHelmReleaseYAML([]byte(fileContent))
	if err != nil {
		return "", fmt.Errorf("error splitting into YAML: %w", err)
	}

	updatedReleases, err := patchRelease(existingReleases, cluster+"-"+name, ns, version)
	if err != nil {
		return "", err
	}

	return helm.MarshalHelmReleases(updatedReleases)
}

func patchRelease(existingReleases []*helmv2beta1.HelmRelease, name, ns, version string) ([]*helmv2beta1.HelmRelease, error) {
	for _, r := range existingReleases {
		if r.Name == name && r.Namespace == ns {
			if r.Spec.Chart.Spec.Version == version {
				return nil, fmt.Errorf("version %s of HelmRelease '%s' already installed in namespace '%s'", version, name, ns)
			}

			r.Spec.Chart.Spec.Version = version

			return existingReleases, nil
		}
	}

	return nil, fmt.Errorf("failed to find HelmRelease '%s' in namespace '%s'", name, ns)
}
