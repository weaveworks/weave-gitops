package gitproviders

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/utils"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/override"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type ProviderAccountType string

const (
	AccountTypeUser ProviderAccountType = "user"
	AccountTypeOrg  ProviderAccountType = "organization"
)

// GitProvider Handler
//counterfeiter:generate . GitProviderHandler
type GitProviderHandler interface {
	CreateRepository(name string, owner string, private bool) error
	RepositoryExists(name string, owner string) (bool, error)
	UploadDeployKey(owner, repoName string, deployKey []byte) error
}

// making sure it implements the interface
var _ GitProviderHandler = defaultGitProviderHandler{}

func New() defaultGitProviderHandler {
	return defaultGitProviderHandler{}
}

var gitProviderHandler interface{} = defaultGitProviderHandler{}

// TODO: implement the New method and inject dependencies in the struct
type defaultGitProviderHandler struct{}

func (h defaultGitProviderHandler) RepositoryExists(name string, owner string) (bool, error) {
	provider, err := GithubProvider()
	if err != nil {
		return false, err
	}

	ownerType, err := GetAccountType(provider, owner)
	if err != nil {
		return false, err
	}

	ctx := context.Background()

	if ownerType == AccountTypeOrg {
		orgRef := gitprovider.OrgRepositoryRef{
			OrganizationRef: gitprovider.OrganizationRef{Domain: github.DefaultDomain, Organization: owner},
			RepositoryName:  name,
		}
		if _, err := provider.OrgRepositories().Get(ctx, orgRef); err != nil {
			return false, err
		}

		return true, nil
	}

	userRepoRef := gitprovider.UserRepositoryRef{
		UserRef:        gitprovider.UserRef{Domain: github.DefaultDomain, UserLogin: owner},
		RepositoryName: name,
	}
	if _, err := provider.UserRepositories().Get(ctx, userRepoRef); err != nil {
		return false, err
	}

	return true, nil
}

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

	repoCreateOpts := &gitprovider.RepositoryCreateOptions{
		AutoInit:        gitprovider.BoolVar(true),
		LicenseTemplate: gitprovider.LicenseTemplateVar(gitprovider.LicenseTemplateApache2),
	}

	ownerType, err := GetAccountType(provider, owner)
	if err != nil {
		return err
	}

	if ownerType == AccountTypeOrg {
		orgRef := NewOrgRepositoryRef(github.DefaultDomain, owner, name)
		if err = CreateOrgRepository(provider, orgRef, repoInfo, repoCreateOpts); err != nil {
			return err
		}
	} else {
		userRef := NewUserRepositoryRef(github.DefaultDomain, owner, name)
		if err = CreateUserRepository(provider, userRef, repoInfo, repoCreateOpts); err != nil {
			return err
		}
	}

	if err = utils.WaitUntil(os.Stdout, time.Second, time.Second*30, func() error {
		return GetRepoInfo(githubProvider, ownerType, owner, name)
	}); err != nil {
		return fmt.Errorf("could not verify repo existence %s", err)
	}

	return nil
}

func (h defaultGitProviderHandler) UploadDeployKey(owner, repoName string, deployKey []byte) error {
	provider, err := GithubProvider()
	if err != nil {
		return err
	}

	deployKeyInfo := gitprovider.DeployKeyInfo{
		Name: "weave-gitops-deploy-key",
		Key:  deployKey,
	}

	ownerType, err := GetAccountType(provider, owner)
	if err != nil {
		return err
	}

	switch ownerType {
	case AccountTypeOrg:
		ctx := context.Background()
		defer ctx.Done()
		orgRef := NewOrgRepositoryRef(github.DefaultDomain, owner, repoName)
		orgRepo, err := provider.OrgRepositories().Get(ctx, orgRef)
		if err != nil {
			return fmt.Errorf("error getting org repo reference for owner %s, repo %s, %s ", owner, repoName, err)
		}
		fmt.Printf("Uploading deploy key to %s\n", repoName)
		_, err = orgRepo.DeployKeys().Create(ctx, deployKeyInfo)
		if err != nil {
			return fmt.Errorf("error uploading deploy key %s", err)
		}
	case AccountTypeUser:
		ctx := context.Background()
		defer ctx.Done()
		userRef := NewUserRepositoryRef(github.DefaultDomain, owner, repoName)
		userRepo, err := provider.UserRepositories().Get(ctx, userRef)
		if err != nil {
			return fmt.Errorf("error getting user repo reference for owner %s, repo %s, %s ", owner, repoName, err)
		}
		fmt.Println("Uploading deploy key")
		_, err = userRepo.DeployKeys().Create(ctx, deployKeyInfo)
		if err != nil {
			return fmt.Errorf("error uploading deploy key %s", err)
		}
	default:
		return fmt.Errorf("account type not supported %s", ownerType)
	}

	return nil
}

func CreateRepository(name string, owner string, private bool) error {
	return gitProviderHandler.(GitProviderHandler).CreateRepository(name, owner, private)
}

func RepositoryExists(name string, owner string) (bool, error) {
	return gitProviderHandler.(GitProviderHandler).RepositoryExists(name, owner)
}

func UploadDeployKey(owner, repoName string, deployKey []byte) error {
	return gitProviderHandler.(GitProviderHandler).UploadDeployKey(owner, repoName, deployKey)
}

func GetAccountType(provider gitprovider.Client, owner string) (ProviderAccountType, error) {
	ctx := context.Background()
	defer ctx.Done()

	_, err := provider.Organizations().Get(ctx, gitprovider.OrganizationRef{
		Domain:       github.DefaultDomain,
		Organization: owner,
	})

	if err != nil {
		if errors.Is(err, gitprovider.ErrNotFound) {
			return AccountTypeUser, nil
		}

		return "", fmt.Errorf("could not get account type %s", err)
	}

	return AccountTypeOrg, nil
}

func GetRepoInfo(provider gitprovider.Client, accountType ProviderAccountType, owner string, repoName string) error {
	ctx := context.Background()
	defer ctx.Done()

	switch accountType {
	case AccountTypeOrg:
		if err := GetOrgRepo(provider, owner, repoName); err != nil {
			return err
		}
	case AccountTypeUser:
		if err := GetUserRepo(provider, owner, repoName); err != nil {
			return err
		}
	default:
		return fmt.Errorf("unexpected account type %s", accountType)
	}

	return nil
}

func GetOrgRepo(provider gitprovider.Client, org string, repoName string) error {
	ctx := context.Background()
	defer ctx.Done()

	orgRepoRef := NewOrgRepositoryRef(github.DefaultDomain, org, repoName)

	_, err := provider.OrgRepositories().Get(ctx, orgRepoRef)
	if err != nil {
		return fmt.Errorf("error getting org repository %s", err)
	}

	return nil
}

func GetUserRepo(provider gitprovider.Client, user string, repoName string) error {
	ctx := context.Background()
	defer ctx.Done()

	userRepoRef := NewUserRepositoryRef(github.DefaultDomain, user, repoName)

	_, err := provider.UserRepositories().Get(ctx, userRepoRef)
	if err != nil {
		return err
	}

	return nil
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
