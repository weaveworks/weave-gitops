package app

import (
	"context"
	"fmt"
	"github.com/fluxcd/go-git-providers/gitprovider"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
)

type CommitParams struct {
	Name             string
	Namespace        string
	GitProviderToken string
	PageSize         int
	PageToken        int
}

// GetCommits gets a list of commits from the repo/branch saved in the app manifest
func (a *AppSvc) GetCommits(gitProvider gitproviders.GitProvider, params CommitParams, application *wego.Application) ([]gitprovider.Commit, error) {
	if application.Spec.SourceType == wego.SourceTypeHelm {
		return nil, fmt.Errorf("unable to get commits for a helm chart")
	}

	repoUrl, err := gitproviders.NewRepoURL(application.Spec.URL)
	if err != nil {
		return nil, fmt.Errorf("error creating normalized url: %w", err)
	}

	commits, err := gitProvider.GetCommits(context.Background(), repoUrl, application.Spec.Branch, params.PageSize, params.PageToken)
	if err != nil {
		return nil, fmt.Errorf("unable to get commits for repo: %w", err)
	}

	return commits, nil
}
