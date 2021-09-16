package app

import (
	"fmt"

	"github.com/fluxcd/go-git-providers/github"
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
func (a *App) GetCommits(params CommitParams, application *wego.Application) ([]gitprovider.Commit, error) {
	normalizedUrl, err := gitproviders.NewNormalizedRepoURL(application.Spec.URL)
	if err != nil {
		return nil, fmt.Errorf("error creating normalized url: %w", err)
	}

	accountType, err := a.GitProvider.GetAccountType(normalizedUrl.Owner())
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account type: %w", err)
	}

	normalizedUrl.Provider()

	var commits []gitprovider.Commit

	if accountType == gitproviders.AccountTypeUser {
		userRepoRef := gitproviders.NewUserRepositoryRef(github.DefaultDomain, normalizedUrl.Owner(), normalizedUrl.RepositoryName())

		commits, err = a.GitProvider.GetCommitsFromUserRepo(userRepoRef, application.Spec.Branch, params.PageSize, params.PageToken)
		if err != nil {
			return nil, fmt.Errorf("unable to get Commits for user repo: %w", err)
		}
	} else {
		orgRepoRef := gitproviders.NewOrgRepositoryRef(github.DefaultDomain, normalizedUrl.Owner(), normalizedUrl.RepositoryName())
		commits, err = a.GitProvider.GetCommitsFromOrgRepo(orgRepoRef, application.Spec.Branch, params.PageSize, params.PageToken)
		if err != nil {
			return nil, fmt.Errorf("unable to get Commits for org repo: %w", err)
		}
	}

	return commits, nil
}
