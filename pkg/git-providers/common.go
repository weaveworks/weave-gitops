package git_providers

import (
	"context"
	"fmt"
	"github.com/fluxcd/go-git-providers/gitprovider"
)

func CreateOrgRepository(provider gitprovider.Client, orgRepoRef gitprovider.OrgRepositoryRef, repoInfo gitprovider.RepositoryInfo, opts ...gitprovider.RepositoryCreateOption) error {

	ctx := context.Background()
	defer ctx.Done()

	_, err := provider.OrgRepositories().Create(ctx, orgRepoRef, repoInfo, opts...)
	if err != nil {
		return fmt.Errorf("error creating repo %s", err)
	}

	return nil
}

func CreateUserRepository(provider gitprovider.Client, userRepoRef gitprovider.UserRepositoryRef, repoInfo gitprovider.RepositoryInfo, opts ...gitprovider.RepositoryCreateOption) error {

	ctx := context.Background()
	defer ctx.Done()

	_, err := provider.UserRepositories().Create(ctx, userRepoRef, repoInfo, opts...)
	if err != nil {
		return fmt.Errorf("error creating repo %s", err)
	}

	return nil
}

func CreatePullRequestToUserRepo(provider gitprovider.Client, userRepRef gitprovider.UserRepositoryRef, targetBranch string, newBranch string, files []gitprovider.CommitFile, commitMessage string, prTitle string, prDescription string) error {

	ctx := context.Background()

	ur, err := provider.UserRepositories().Get(ctx, userRepRef)
	if err != nil {
		return fmt.Errorf("error getting info for repo [%s] err [%s]", userRepRef.String(), err)
	}

	if targetBranch == "" {
		targetBranch = *ur.Get().DefaultBranch
	}

	commits, err := ur.Commits().ListPage(ctx, targetBranch, 1, 0)
	if err != nil {
		return fmt.Errorf("error getting commits for repo[%s] err [%s]", userRepRef.String(), err)
	}

	latestCommit := commits[0]

	if err := ur.Branches().Create(ctx, newBranch, latestCommit.Get().Sha); err != nil {
		return fmt.Errorf("error creating branch[%s] for repo[%s] err [%s]", newBranch, userRepRef.String(), err)
	}

	if _, err := ur.Commits().Create(ctx, newBranch, commitMessage, files); err != nil {
		return fmt.Errorf("error creating commit for branch[%s] for repo[%s] err [%s]", newBranch, userRepRef.String(), err)
	}

	if err := ur.PullRequests().Create(ctx, prTitle, newBranch, targetBranch, prDescription); err != nil {
		return fmt.Errorf("error creating pull request[%s] for branch[%s] for repo[%s] err [%s]", prTitle, newBranch, userRepRef.String(), err)
	}

	return nil
}

func CreatePullRequestToOrgRepo(provider gitprovider.Client, orgRepRef gitprovider.OrgRepositoryRef, targetBranch string, newBranch string, files []gitprovider.CommitFile, commitMessage string, prTitle string, prDescription string) error {

	ctx := context.Background()

	ur, err := provider.OrgRepositories().Get(ctx, orgRepRef)
	if err != nil {
		return fmt.Errorf("error getting info for repo [%s] err [%s]", orgRepRef.String(), err)
	}

	if targetBranch == "" {
		targetBranch = *ur.Get().DefaultBranch
	}

	commits, err := ur.Commits().ListPage(ctx, targetBranch, 1, 0)
	if err != nil {
		return fmt.Errorf("error getting commits for repo[%s] err [%s]", orgRepRef.String(), err)
	}

	latestCommit := commits[0]

	if err := ur.Branches().Create(ctx, newBranch, latestCommit.Get().Sha); err != nil {
		return fmt.Errorf("error creating branch[%s] for repo[%s] err [%s]", newBranch, orgRepRef.String(), err)
	}

	if _, err := ur.Commits().Create(ctx, newBranch, commitMessage, files); err != nil {
		return fmt.Errorf("error creating commit for branch[%s] for repo[%s] err [%s]", newBranch, orgRepRef.String(), err)
	}

	if err := ur.PullRequests().Create(ctx, prTitle, newBranch, targetBranch, prDescription); err != nil {
		return fmt.Errorf("error creating pull request[%s] for branch[%s] for repo[%s] err [%s]", prTitle, newBranch, orgRepRef.String(), err)
	}

	return nil
}

func NewRepositoryInfo(description string, visibility gitprovider.RepositoryVisibility) gitprovider.RepositoryInfo {
	return gitprovider.RepositoryInfo{
		Description: &description,
		Visibility:  &visibility,
	}
}

func NewOrgRepositoryRef(domain, org, repoName string) gitprovider.OrgRepositoryRef {
	return gitprovider.OrgRepositoryRef{
		RepositoryName: repoName,
		OrganizationRef: gitprovider.OrganizationRef{
			Domain:       domain,
			Organization: org,
		},
	}
}

func NewUserRepositoryRef(domain, user, repoName string) gitprovider.UserRepositoryRef {
	return gitprovider.UserRepositoryRef{
		RepositoryName: repoName,
		UserRef: gitprovider.UserRef{
			Domain:    domain,
			UserLogin: user,
		},
	}
}
