package profiles

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"k8s.io/apimachinery/pkg/types"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	goyaml "github.com/go-yaml/yaml"
	"sigs.k8s.io/yaml"
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
	configRepoUrl, err := gitproviders.NewRepoURL(opts.ConfigRepo)
	if err != nil {
		return fmt.Errorf("failed to parse url: %w", err)
	}

	repoExists, err := gitProvider.RepositoryExists(ctx, configRepoUrl)
	if err != nil {
		return fmt.Errorf("failed to check whether repository exists: %w", err)
	} else if !repoExists {
		return fmt.Errorf("repository '%v' could not be found", configRepoUrl.String())
	}
	defaultBranch, err := gitProvider.GetDefaultBranch(ctx, configRepoUrl)
	if err != nil {
		return fmt.Errorf("failed to get default branch: %w", err)
	}

	availableProfile, version, err := s.GetProfile(ctx, GetOptions{
		Name:      opts.Name,
		Version:   opts.Version,
		Cluster:   opts.Cluster,
		Namespace: opts.Namespace,
		Port:      opts.ProfilesPort,
	})
	if err != nil {
		return fmt.Errorf("failed to get profiles from cluster: %w", err)
	}
	if availableProfile.GetHelmRepository().GetName() == "" || availableProfile.GetHelmRepository().GetNamespace() == "" {
		return fmt.Errorf("failed to discover HelmRepository's name and namespace")
	}

	helmRepo := types.NamespacedName{
		Name:      availableProfile.HelmRepository.Name,
		Namespace: availableProfile.HelmRepository.Namespace,
	}

	newRelease := helm.MakeHelmRelease(opts.Name, version, opts.Cluster, opts.Namespace, helmRepo)

	files, err := gitProvider.GetRepoDirFiles(ctx, configRepoUrl, git.GetSystemPath(opts.Cluster), defaultBranch)
	if err != nil {
		return fmt.Errorf("failed to get files in '%s' for config repository '%s': %s", git.GetSystemPath(opts.Cluster), configRepoUrl.String(), err)
	}

	file, err := AppendProfileToFile(files, newRelease, git.GetProfilesPath(opts.Cluster, models.WegoProfilesPath))
	if err != nil {
		return fmt.Errorf("failed to append HelmRelease to profiles file: %w", err)
	}

	pr, err := gitProvider.CreatePullRequest(ctx, configRepoUrl, gitproviders.PullRequestInfo{
		Title:         fmt.Sprintf("GitOps add %s", opts.Name),
		Description:   fmt.Sprintf("Add manifest for %s profile", opts.Name),
		CommitMessage: AddCommitMessage,
		TargetBranch:  defaultBranch,
		NewBranch:     automation.GetRandomString("wego-"),
		Files:         []gitprovider.CommitFile{file},
	})
	if err != nil {
		return fmt.Errorf("failed to create pull request: %s", err)
	}
	s.Logger.Actionf("Pull Request created: %s", pr.Get().WebURL)

	if opts.AutoMerge {
		s.Logger.Actionf("auto-merge=true; merging PR number %v", pr.Get().Number)
		if err := gitProvider.MergePullRequest(ctx, configRepoUrl, pr.Get().Number, AddCommitMessage); err != nil {
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

// AppendProfileToFile appends a HelmRelease to profiles.yaml if file does not contain other HelmRelease with the same name and namespace.
func AppendProfileToFile(files []*gitprovider.CommitFile, newRelease *v2beta1.HelmRelease, path string) (gitprovider.CommitFile, error) {
	var content string
	for _, f := range files {
		if f.Path != nil && *f.Path == path {
			if f.Content == nil || *f.Content == "" {
				break
			}

			manifestByteSlice, err := splitYAML([]byte(*f.Content))
			if err != nil {
				return gitprovider.CommitFile{}, fmt.Errorf("error splitting %s: %w", models.WegoProfilesPath, err)
			}

			for _, manifestBytes := range manifestByteSlice {
				var r v2beta1.HelmRelease
				if err := yaml.Unmarshal(manifestBytes, &r); err != nil {
					return gitprovider.CommitFile{}, fmt.Errorf("error unmarshaling %s: %w", models.WegoProfilesPath, err)
				}
				if profileIsInstalled(r, *newRelease) {
					return gitprovider.CommitFile{}, fmt.Errorf("version %s of profile '%s' already exists in namespace %s", r.Spec.Chart.Spec.Version, r.Name, r.Namespace)
				}
			}
			content = *f.Content
		}
	}
	helmReleaseManifest, err := yaml.Marshal(newRelease)
	if err != nil {
		return gitprovider.CommitFile{}, fmt.Errorf("failed to marshal new HelmRelease: %w", err)
	}
	content += "\n---\n" + string(helmReleaseManifest)

	return gitprovider.CommitFile{
		Path:    &path,
		Content: &content,
	}, nil
}

// splitYAML splits a manifest file that may contain multiple YAML resources separated by '---'
// and validates that each element is YAML.
func splitYAML(resources []byte) ([][]byte, error) {
	decoder := goyaml.NewDecoder(bytes.NewReader(resources))
	var splitResources [][]byte
	for {
		var value interface{}
		if err := decoder.Decode(&value); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		valueBytes, err := goyaml.Marshal(value)
		if err != nil {
			return nil, err
		}
		splitResources = append(splitResources, valueBytes)
	}
	return splitResources, nil
}

func profileIsInstalled(r, newRelease v2beta1.HelmRelease) bool {
	return r.Name == newRelease.Name && r.Namespace == newRelease.Namespace && r.Spec.Chart.Spec.Version == newRelease.Spec.Chart.Spec.Version
}
