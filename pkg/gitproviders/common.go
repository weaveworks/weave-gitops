package gitproviders

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/override"
)

const (
	OwnerTypeUser = "user"
	OwnerTypeOrg  = "organization"
)

// GitProvider Handler
type GitProviderHandler interface {
	CreateRepository(name string, owner string, private bool) error
}

var gitProviderHandler interface{} = defaultGitProviderHandler{}

type defaultGitProviderHandler struct{}

func (h defaultGitProviderHandler) CreateRepository(name string, owner string, private bool) error {
	// TODO: detect or receive the provider when necessary
	provider, err := GithubProvider()
	if err != nil {
		return err
	}

	visibility := gitprovider.RepositoryVisibilityPrivate
	if !private {
		visibility = gitprovider.RepositoryVisibilityPublic
	}
	repoInfo := NewRepositoryInfo("Weave Gitops repo", visibility)

	bts, err := json.Marshal(repoInfo)
	fmt.Println("VISIBILITY-repoInfo", string(bts))
	fmt.Println("VISIBILITY-err", err)
	fmt.Println("VISIBILITY-NAME", name)
	fmt.Println("VISIBILITY-OWNER", owner)

	repoCreateOpts := &gitprovider.RepositoryCreateOptions{
		AutoInit:        gitprovider.BoolVar(true),
		LicenseTemplate: gitprovider.LicenseTemplateVar(gitprovider.LicenseTemplateApache2),
	}

	ownerType, err := getOwnerType(provider, owner)
	if err != nil {
		return err
	}

	if ownerType == OwnerTypeOrg {
		orgRef := NewOrgRepositoryRef(github.DefaultDomain, owner, name)

		return CreateOrgRepository(provider, orgRef, repoInfo, repoCreateOpts)
	}

	userRef := NewUserRepositoryRef(github.DefaultDomain, owner, name)
	return CreateUserRepository(provider, userRef, repoInfo, repoCreateOpts)
}

func CreateRepository(name string, owner string, private bool) error {
	return gitProviderHandler.(GitProviderHandler).CreateRepository(name, owner, private)
}

func getOwnerType(provider gitprovider.Client, owner string) (string, error) {
	ctx := context.Background()
	defer ctx.Done()

	_, err := provider.Organizations().Get(ctx, gitprovider.OrganizationRef{
		Domain:       github.DefaultDomain,
		Organization: owner,
	})

	if err != nil {
		if errors.Is(err, gitprovider.ErrNotFound) {
			return OwnerTypeUser, nil
		}

		return "", err
	}

	return OwnerTypeOrg, nil
}

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

func Override(handler GitProviderHandler) override.Override {
	return override.Override{Handler: &gitProviderHandler, Mock: handler, Original: gitProviderHandler}
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

	if len(commits) == 0 {
		return fmt.Errorf("targetBranch[%s] does not exists", targetBranch)
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

	if len(commits) == 0 {
		return fmt.Errorf("targetBranch[%s] does not exists", targetBranch)
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
