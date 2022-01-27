package profiles

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	goyaml "github.com/go-yaml/yaml"
	"sigs.k8s.io/yaml"
)

const AddCommitMessage = "Add Profile manifests"

type AddOptions struct {
	Name       string
	Cluster    string
	ConfigRepo string
	Version    string
	Port       string
	Namespace  string
	AutoMerge  bool
}

// Add installs an available profile in a cluster's namespace by appending a HelmRelease to the profile manifest in the config repo,
// provided that such a HelmRelease does not exist with the same profile name and version in the same namespace and cluster.
func (s *ProfilesSvc) Add(ctx context.Context, gitProvider gitproviders.GitProvider, opts AddOptions) error {
	configRepoUrl, err := gitproviders.NewRepoURL(opts.ConfigRepo, true)
	if err != nil {
		return err
	}

	// TODO: refactor this into 1 func that returns repoRef
	repoExists, err := gitProvider.RepositoryExists(ctx, configRepoUrl)
	if err != nil {
		return err
	} else if !repoExists {
		return fmt.Errorf("repository '%v' could not be found", configRepoUrl.String())
	}
	defaultBranch, err := gitProvider.GetDefaultBranch(ctx, configRepoUrl)
	if err != nil {
		return err
	}

	// TODO: We need a ticket for when the user auto-completes the chart name, helm repository name, and helm repository namespace,
	// so that a user doesn't only rely on hitting the Kube API if they don't have access to the cluster.
	availableProfile, err := s.GetAvailableProfile(ctx, GetOptions{
		Name:      opts.Name,
		Version:   opts.Version,
		Cluster:   opts.Cluster,
		Namespace: opts.Namespace,
	})
	if err != nil {
		return err
	}

	newRelease := helm.MakeHelmRelease(availableProfile, opts.Cluster, opts.Namespace)

	files, err := gitProvider.GetRepoFiles(ctx, configRepoUrl, git.GetSystemPath(opts.Cluster), defaultBranch)
	if err != nil {
		return fmt.Errorf("failed to get files in '%s' for config repository '%s': %s", git.GetSystemPath(opts.Cluster), configRepoUrl.String(), err)
	}

	file, err := MakeManifestFile(files, newRelease, git.GetProfilesPath(opts.Cluster))
	if err != nil {
		return err
	}

	pr, err := gitProvider.CreatePullRequest(ctx, configRepoUrl, gitproviders.PullRequestInfo{
		Title:         fmt.Sprintf("GitOps add %s", opts.Name),
		Description:   fmt.Sprintf("Add manifest for %s profile", opts.Name),
		CommitMessage: AddCommitMessage,
		TargetBranch:  defaultBranch,
		NewBranch:     automation.GetRandomName("wego-"),
		Files:         file,
	})
	if err != nil {
		return fmt.Errorf("failed to create the pull request: %s", err)
	}
	s.Logger.Actionf("Pull Request created: %s", pr.Get().WebURL)

	if opts.AutoMerge {
		s.Logger.Actionf("auto-merge=true; merging PR number %s", pr.Get().Number)
		if err := gitProvider.MergePullRequest(ctx, configRepoUrl, pr.Get().Number, gitprovider.MergeMethodMerge, AddCommitMessage); err != nil {
			return err
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
	s.Logger.Println("Namespace: %s", opts.Namespace)
	s.Logger.Println("")
}

func MakeManifestFile(files []*gitprovider.CommitFile, newRelease *v2beta1.HelmRelease, path string) ([]gitprovider.CommitFile, error) {
	var content string
	for _, f := range files {
		if *f.Path == path {
			if *f.Content == "" {
				break
			}

			manifestByteSlice, err := splitYAML([]byte(*f.Content))
			if err != nil {
				return nil, err
			}

			for _, manifestBytes := range manifestByteSlice {
				var r v2beta1.HelmRelease
				if err := yaml.Unmarshal(manifestBytes, &r); err != nil {
					return nil, err
				}
				if profileIsInstalled(r, *newRelease) {
					return nil, fmt.Errorf("version %s of profile '%s' already exists in namespace %s", r.Spec.Chart.Spec.Version, r.Name, r.Namespace)
				}
			}
			content = *f.Content + "\n---\n"
		}
	}
	helmReleaseManifest, err := yaml.Marshal(newRelease)
	if err != nil {
		return nil, err
	}
	content = content + string(helmReleaseManifest)

	return []gitprovider.CommitFile{
		{
			Path:    &path,
			Content: &content,
		},
	}, nil
}

// splitYAML splits a manifest file that contains multiple YAML resources separated by '---'.
func splitYAML(resources []byte) ([][]byte, error) {
	decoder := goyaml.NewDecoder(bytes.NewReader(resources))
	var splitResources [][]byte
	for {
		var value interface{}
		err := decoder.Decode(&value)
		if err == io.EOF {
			break
		}
		if err != nil {
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
