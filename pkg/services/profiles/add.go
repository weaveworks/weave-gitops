package profiles

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
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
	validatedOps, err := ValidateAddOptions(opts)
	if err != nil {
		return err
	}

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

	_, err = doKubeProfilesGetRequest(ctx, opts.Namespace, wegoServiceName, opts.Port, getProfilesPath, s.ClientSet)
	if err != nil {
		return err
	}

	printAddSummary(validatedOps)
	return nil
}

func ValidateAddOptions(opts AddOptions) (AddOptions, error) {
	if opts.ConfigRepo == "" {
		return opts, errors.New("--config-repo should be provided")
	}

	if opts.Name == "" {
		return opts, errors.New("--name should be provided")
	}

	if opts.Cluster == "" {
		return opts, errors.New("--cluster should be provided")
	}

	if automation.ApplicationNameTooLong(opts.Name) {
		return opts, fmt.Errorf("--name value is too long: %s; must be <= %d characters",
			opts.Name, automation.MaxKubernetesResourceNameLength)
	}

	if strings.HasPrefix(opts.Name, "wego") {
		return opts, fmt.Errorf("the prefix 'wego' is used by weave gitops and is not allowed for a profile name")
	}

	return opts, nil
}

func printAddSummary(opts AddOptions) {
	opts.Logger.Println("Adding profile:\n")
	opts.Logger.Println("Name: %s", opts.Name)
	opts.Logger.Println("Version: %s", opts.Version)
	opts.Logger.Printf("Cluster: %s", opts.Cluster)

	opts.Logger.Println("")
}
