package app

import (
	"context"
	"fmt"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
)

type CommitParams struct {
	Name             string
	Namespace        string
	GitProviderToken string
	PageSize         int
	PageToken        int
}

// GetCommits gets a list of commits from the repo/branch saved in the app manifest
func (a *App) GetCommits(params CommitParams) ([]gitprovider.Commit, error) {
	ctx := context.Background()

	app, err := a.kube.GetApplication(ctx, types.NamespacedName{Name: params.Name, Namespace: params.Namespace})
	if err != nil {
		return nil, fmt.Errorf("unable to get application for %s %w", params.Name, err)
	}

	if app.Spec.SourceType == "helm" {
		return nil, fmt.Errorf("unable to get commits for a helm chart")
	}

	owner, err := utils.GetOwnerFromUrl(app.Spec.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve owner: %w", err)
	}

	gitProvider, err := a.gitProviderFactory(params.GitProviderToken)
	if err != nil {
		return nil, err
	}

	accountType, err := gitProvider.GetAccountType(owner)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve account type: %w", err)
	}

	var commits []gitprovider.Commit
	repoName := utils.UrlToRepoName(app.Spec.URL)

	if accountType == gitproviders.AccountTypeUser {
		userRepoRef := gitproviders.NewUserRepositoryRef(github.DefaultDomain, owner, repoName)
		commits, err = gitProvider.GetCommitsFromUserRepo(userRepoRef, app.Spec.Branch, params.PageSize, params.PageToken)
		if err != nil {
			return nil, fmt.Errorf("unable to get Commits for user repo: %w", err)
		}
	} else {
		orgRepoRef := gitproviders.NewOrgRepositoryRef(github.DefaultDomain, owner, repoName)
		commits, err = gitProvider.GetCommitsFromOrgRepo(orgRepoRef, app.Spec.Branch, params.PageSize, params.PageToken)
		if err != nil {
			return nil, fmt.Errorf("unable to get Commits for org repo: %w", err)
		}
	}

	return commits, nil
}
