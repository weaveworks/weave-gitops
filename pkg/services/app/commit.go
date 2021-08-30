package app

import (
	"fmt"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/utils"
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
	owner, err := utils.GetOwnerFromUrl(application.Spec.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve owner: %w", err)
	}

	gitProvider, err := a.GitProviderFactory(params.GitProviderToken)
	if err != nil {
		return nil, err
	}

	accountType, err := gitProvider.GetAccountType(owner)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account type: %w", err)
	}

	var commits []gitprovider.Commit
	repoName := utils.UrlToRepoName(application.Spec.URL)

	if accountType == gitproviders.AccountTypeUser {
		userRepoRef := gitproviders.NewUserRepositoryRef(github.DefaultDomain, owner, repoName)
		commits, err = gitProvider.GetCommitsFromUserRepo(userRepoRef, application.Spec.Branch, params.PageSize, params.PageToken)
		if err != nil {
			return nil, fmt.Errorf("unable to get Commits for user repo: %w", err)
		}
	} else {
		orgRepoRef := gitproviders.NewOrgRepositoryRef(github.DefaultDomain, owner, repoName)
		commits, err = gitProvider.GetCommitsFromOrgRepo(orgRepoRef, application.Spec.Branch, params.PageSize, params.PageToken)
		if err != nil {
			return nil, fmt.Errorf("unable to get Commits for org repo: %w", err)
		}
	}

	return commits, nil
}
