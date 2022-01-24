package profiles

import (
	"context"
	"fmt"
	"sort"

	"github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

type AddOptions struct {
	Name       string
	Cluster    string
	ConfigRepo string
	Version    string
	Port       string
	Namespace  string
	AutoMerge  bool
	Logger     logger.Logger
}

// Add adds a new profile to the cluster
func (s *ProfilesSvc) Add(ctx context.Context, gitProvider gitproviders.GitProvider, opts AddOptions) error {
	configRepoUrl, err := gitproviders.NewRepoURL(opts.ConfigRepo)
	if err != nil {
		return err
	}

	repoExists, err := gitProvider.RepositoryExists(ctx, configRepoUrl)
	if err != nil {
		return err
	} else if !repoExists {
		return fmt.Errorf("repository '%v' could not be found", configRepoUrl.String())
	}

	_, err = gitProvider.GetRepoFiles(ctx, configRepoUrl, git.GetSystemPath(opts.Cluster), "")
	if err != nil {
		return fmt.Errorf("failed to get files in '%s' for config repository '%s': %s", git.GetSystemPath(opts.Cluster), configRepoUrl.String(), err)
	}

	profilesList, err := doKubeProfilesGetRequest(ctx, opts.Namespace, wegoServiceName, opts.Port, getProfilesPath, s.ClientSet)
	if err != nil {
		return err
	}

	availableProfile, err := getAvailableProfile(profilesList, opts)
	if err != nil {
		return err
	}

	_ = helm.MakeHelmRelease(availableProfile, opts.Cluster, opts.Namespace, opts.Version)

	printAddSummary(opts)
	return nil
}

func printAddSummary(opts AddOptions) {
	opts.Logger.Println("Adding profile:\n")
	opts.Logger.Println("Name: %s", opts.Name)
	opts.Logger.Println("Version: %s", opts.Version)
	opts.Logger.Printf("Cluster: %s", opts.Cluster)

	opts.Logger.Println("")
}

func getAvailableProfile(profilesList *profiles.GetProfilesResponse, opts AddOptions) (*profiles.Profile, error) {
	for _, p := range profilesList.Profiles {
		if p.Name == opts.Name {
			if len(p.AvailableVersions) == 0 {
				return nil, fmt.Errorf("no version found for profile '%s' in %s/%s", p.Name, opts.Cluster, opts.Namespace)
			}
			switch {
			case opts.Version == "latest":
				sort.Strings(p.AvailableVersions)
				opts.Version = p.AvailableVersions[len(p.AvailableVersions)-1]
			default:
				if !found(p.AvailableVersions, opts.Version) {
					return nil, fmt.Errorf("version '%s' not found for profile '%s' in %s/%s", opts.Version, opts.Name, opts.Cluster, opts.Namespace)
				}
			}
			return p, nil
		}
	}
	return nil, fmt.Errorf("no available profile '%s' found in %s/%s", opts.Name, opts.Cluster, opts.Namespace)
}

func found(availableVersions []string, version string) bool {
	for _, v := range availableVersions {
		if v == version {
			return true
		}
	}
	return false
}
