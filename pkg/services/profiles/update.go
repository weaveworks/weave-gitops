package profiles

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
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

	_, _, err = s.GetProfile(ctx, GetOptions{
		Name:      opts.Name,
		Version:   opts.Version,
		Cluster:   opts.Cluster,
		Namespace: opts.Namespace,
		Port:      opts.ProfilesPort,
	})
	if err != nil {
		return fmt.Errorf("failed to get profiles from cluster: %w", err)
	}

	return nil
}
