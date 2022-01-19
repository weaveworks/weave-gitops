package profile

import (
	"context"
	"errors"
	"fmt"
	"path"
	"strings"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
)

type AddParams struct {
	Name       string
	Cluster    string
	ConfigRepo string
	Version    string
	AutoMerge  bool
}

// Add adds a new profile to the cluster
func (s *ProfileSvc) Add(gitProvider gitproviders.GitProvider, params AddParams) error {
	validatedParams, err := s.ValidateAddParams(params)
	if err != nil {
		return err
	}

	ctx := context.Background()

	configRepoUrl, err := gitproviders.NewRepoURL(params.ConfigRepo)
	if err != nil {
		return err
	}

	repoExists, err := gitProvider.RepositoryExists(ctx, configRepoUrl)
	if err != nil {
		return err
	} else if !repoExists {
		return fmt.Errorf("repository '%v' could not be found", configRepoUrl.String())
	}

	files, err := gitProvider.GetRepoFiles(ctx, configRepoUrl, git.WegoRoot, "main")
	if err != nil {
		return fmt.Errorf("failed to get files of config repository '%s': %s", configRepoUrl.String(), err)
	}

	found := getClusterName(files, params)
	if !found {
		return fmt.Errorf("failed to find cluster in '/%s/%s/'", git.WegoRoot, git.WegoClusterDir)
	}

	s.printAddSummary(validatedParams)
	return nil
}

func (s *ProfileSvc) ValidateAddParams(params AddParams) (AddParams, error) {
	if params.ConfigRepo == "" {
		return params, errors.New("--config-repo should be provided")
	}

	if params.Name == "" {
		return params, errors.New("--name should be provided")
	}

	if automation.ApplicationNameTooLong(params.Name) {
		return params, fmt.Errorf("--name value is too long: %s; must be <= %d characters",
			params.Name, automation.MaxKubernetesResourceNameLength)
	}

	if strings.HasPrefix(params.Name, "wego") {
		return params, fmt.Errorf("the prefix 'wego' is used by weave gitops and is not allowed for a profile name")
	}

	return params, nil
}

func (s *ProfileSvc) printAddSummary(params AddParams) {
	s.Logger.Println("Adding profile:\n")
	s.Logger.Println("Name: %s", params.Name)
	s.Logger.Println("Version: %s", params.Version)
	s.Logger.Printf("Cluster: %s", params.Cluster)

	s.Logger.Println("")
}

func getClusterName(files []*gitprovider.CommitFile, params AddParams) bool {
	for _, f := range files {
		if strings.Contains(*f.Path, path.Join(git.WegoClusterDir, params.Cluster)) {
			return true
		}
	}
	return false
}
