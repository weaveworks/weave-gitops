package gitproviders

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/utils"
	"golang.org/x/oauth2"

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

type ProviderName string

const (
	ProviderNameGithub ProviderName = "github"
	ProviderNameGitlab ProviderName = "gitlab"
)

// GitProvider Handler
//counterfeiter:generate . GitProviderHandler
type GitProviderHandler interface {
	CreateRepository(name string, owner string, private bool) error
	RepositoryExists(name string, owner string) (bool, error)
	DeployKeyExists(owner, repoName string) (bool, error)
	UploadDeployKey(owner, repoName string, deployKey []byte) error
	CreatePullRequestToUserRepo(userRepRef gitprovider.UserRepositoryRef, targetBranch string, newBranch string, files []gitprovider.CommitFile, commitMessage string, prTitle string, prDescription string) (gitprovider.PullRequest, error)
	CreatePullRequestToOrgRepo(orgRepRef gitprovider.OrgRepositoryRef, targetBranch string, newBranch string, files []gitprovider.CommitFile, commitMessage string, prTitle string, prDescription string) (gitprovider.PullRequest, error)
	GetAccountType(owner string) (ProviderAccountType, error)
	OauthConfig() OauthProviderConfig
	GetUser(ctx context.Context, token *oauth2.Token) (*User, error)
}

// OauthProviderConfig is an abstraction around the oauth2.Config to allow for mocking.
// This represents the subset of funcs the API server needs from the oauth2.Config.
//counterfeiter:generate . OauthProviderConfig
type OauthProviderConfig interface {
	AuthCodeURL(state string, opts ...oauth2.AuthCodeOption) string
	Exchange(ctx context.Context, code string, opts ...oauth2.AuthCodeOption) (*oauth2.Token, error)
}

// making sure it implements the interface
var _ GitProviderHandler = defaultGitProviderHandler{}

type User struct {
	Email string
}

// TODO: currently, this always returns the github implementation.
// This eventually needs to return the different git provider implementations.
func New(providerName ProviderName) (defaultGitProviderHandler, error) {
	switch providerName {
	case ProviderNameGithub:
		return defaultGitProviderHandler{}, nil
	}
	return defaultGitProviderHandler{}, fmt.Errorf("provider name %s is not supported", providerName)
}

var gitProviderHandler interface{} = defaultGitProviderHandler{}

func GetSupportedProviders() []ProviderName {
	return []ProviderName{ProviderNameGithub}
}

// TODO: implement the New method and inject dependencies in the struct
type defaultGitProviderHandler struct{}

func (h defaultGitProviderHandler) RepositoryExists(name string, owner string) (bool, error) {
	provider, err := GithubProvider()
	if err != nil {
		return false, err
	}

	ownerType, err := h.GetAccountType(owner)
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

	ownerType, err := h.GetAccountType(owner)
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

	return nil
}

func (h defaultGitProviderHandler) DeployKeyExists(owner, repoName string) (bool, error) {

	provider, err := GithubProvider()
	if err != nil {
		return false, err
	}

	deployKeyName := "weave-gitops-deploy-key"

	ownerType, err := h.GetAccountType(owner)
	if err != nil {
		return false, err
	}

	ctx := context.Background()
	defer ctx.Done()
	switch ownerType {
	case AccountTypeOrg:
		orgRef := NewOrgRepositoryRef(github.DefaultDomain, owner, repoName)
		orgRepo, err := provider.OrgRepositories().Get(ctx, orgRef)
		if err != nil {
			return false, fmt.Errorf("error getting org repo reference for owner %s, repo %s, %s ", owner, repoName, err)
		}
		_, err = orgRepo.DeployKeys().Get(ctx, deployKeyName)
		if err != nil && !strings.Contains(err.Error(), "key is already in use") {
			if errors.Is(err, gitprovider.ErrNotFound) {
				return false, nil
			} else {
				return false, fmt.Errorf("error getting deploy key %s for repo %s. %s", deployKeyName, repoName, err)
			}
		} else {
			return true, nil
		}

	case AccountTypeUser:
		userRef := NewUserRepositoryRef(github.DefaultDomain, owner, repoName)
		userRepo, err := provider.UserRepositories().Get(ctx, userRef)
		if err != nil {
			return false, fmt.Errorf("error getting user repo reference for owner %s, repo %s, %s ", owner, repoName, err)
		}
		_, err = userRepo.DeployKeys().Get(ctx, deployKeyName)
		if err != nil && !strings.Contains(err.Error(), "key is already in use") {
			if errors.Is(err, gitprovider.ErrNotFound) {
				return false, nil
			} else {
				return false, fmt.Errorf("error getting deploy key %s for repo %s. %s", deployKeyName, repoName, err)
			}
		} else {
			return true, nil
		}
	default:
		return false, fmt.Errorf("account type not supported %s", ownerType)
	}
}

func (h defaultGitProviderHandler) UploadDeployKey(owner, repoName string, deployKey []byte) error {

	provider, err := GithubProvider()
	if err != nil {
		return err
	}

	deployKeyName := "weave-gitops-deploy-key"
	deployKeyInfo := gitprovider.DeployKeyInfo{
		Name: deployKeyName,
		Key:  deployKey,
	}

	ownerType, err := h.GetAccountType(owner)
	if err != nil {
		return err
	}

	ctx := context.Background()
	defer ctx.Done()
	switch ownerType {
	case AccountTypeOrg:
		orgRef := NewOrgRepositoryRef(github.DefaultDomain, owner, repoName)
		orgRepo, err := provider.OrgRepositories().Get(ctx, orgRef)
		if err != nil {
			return fmt.Errorf("error getting org repo reference for owner %s, repo %s, %s ", owner, repoName, err)
		}
		fmt.Println("uploading deploy key")
		_, err = orgRepo.DeployKeys().Create(ctx, deployKeyInfo)
		if err != nil {
			return fmt.Errorf("error uploading deploy key %s", err)
		}
		if err = utils.WaitUntil(os.Stdout, time.Second, time.Second*30, func() error {
			_, err = orgRepo.DeployKeys().Get(ctx, deployKeyName)
			return err
		}); err != nil {
			return fmt.Errorf("error verifying deploy key %s existance for repo %s. %s", deployKeyName, repoName, err)
		}
	case AccountTypeUser:
		userRef := NewUserRepositoryRef(github.DefaultDomain, owner, repoName)
		userRepo, err := provider.UserRepositories().Get(ctx, userRef)
		if err != nil {
			return fmt.Errorf("error getting user repo reference for owner %s, repo %s, %s ", owner, repoName, err)
		}
		fmt.Println("uploading deploy key")
		_, err = userRepo.DeployKeys().Create(ctx, deployKeyInfo)
		if err != nil {
			return fmt.Errorf("error uploading deploy key %s", err)
		}
		if err = utils.WaitUntil(os.Stdout, time.Second, time.Second*30, func() error {
			_, err = userRepo.DeployKeys().Get(ctx, deployKeyName)
			return err
		}); err != nil {
			return fmt.Errorf("error verifying deploy key %s existance for repo %s. %s", deployKeyName, repoName, err)
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

func DeployKeyExists(owner, repoName string) (bool, error) {
	return gitProviderHandler.(GitProviderHandler).DeployKeyExists(owner, repoName)
}
func CreatePullRequestToUserRepo(userRepRef gitprovider.UserRepositoryRef, targetBranch string, newBranch string, files []gitprovider.CommitFile, commitMessage string, prTitle string, prDescription string) (gitprovider.PullRequest, error) {
	return gitProviderHandler.(GitProviderHandler).CreatePullRequestToUserRepo(userRepRef, targetBranch, newBranch, files, commitMessage, prTitle, prDescription)
}

func CreatePullRequestToOrgRepo(orgRepRef gitprovider.OrgRepositoryRef, targetBranch string, newBranch string, files []gitprovider.CommitFile, commitMessage string, prTitle string, prDescription string) (gitprovider.PullRequest, error) {
	return gitProviderHandler.(GitProviderHandler).CreatePullRequestToOrgRepo(orgRepRef, targetBranch, newBranch, files, commitMessage, prTitle, prDescription)
}

func (h defaultGitProviderHandler) GetAccountType(owner string) (ProviderAccountType, error) {
	provider, err := GithubProvider()
	if err != nil {
		return "", err
	}
	ctx := context.Background()
	defer ctx.Done()

	_, err = provider.Organizations().Get(ctx, gitprovider.OrganizationRef{
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

	return waitUntilRepoCreated(AccountTypeOrg, orgRepoRef.Organization, orgRepoRef.RepositoryName)
}

func CreateUserRepository(provider gitprovider.Client, userRepoRef gitprovider.UserRepositoryRef, repoInfo gitprovider.RepositoryInfo, opts ...gitprovider.RepositoryCreateOption) error {
	ctx := context.Background()
	defer ctx.Done()

	_, err := provider.UserRepositories().Create(ctx, userRepoRef, repoInfo, opts...)
	if err != nil {
		return fmt.Errorf("error creating repo %s", err)
	}

	return waitUntilRepoCreated(AccountTypeUser, userRepoRef.UserLogin, userRepoRef.RepositoryName)
}

func Override(handler GitProviderHandler) override.Override {
	return override.Override{Handler: &gitProviderHandler, Mock: handler, Original: gitProviderHandler}
}

func (h defaultGitProviderHandler) CreatePullRequestToUserRepo(userRepRef gitprovider.UserRepositoryRef, targetBranch string, newBranch string, files []gitprovider.CommitFile, commitMessage string, prTitle string, prDescription string) (gitprovider.PullRequest, error) {
	provider, err := GithubProvider()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	ur, err := provider.UserRepositories().Get(ctx, userRepRef)
	if err != nil {
		return nil, fmt.Errorf("error getting info for repo [%s] err [%s]", userRepRef.String(), err)
	}

	if targetBranch == "" {
		targetBranch = *ur.Get().DefaultBranch
	}

	commits, err := ur.Commits().ListPage(ctx, targetBranch, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("error getting commits for repo[%s] err [%s]", userRepRef.String(), err)
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("targetBranch[%s] does not exists", targetBranch)
	}

	latestCommit := commits[0]

	if err := ur.Branches().Create(ctx, newBranch, latestCommit.Get().Sha); err != nil {
		return nil, fmt.Errorf("error creating branch[%s] for repo[%s] err [%s]", newBranch, userRepRef.String(), err)
	}

	if _, err := ur.Commits().Create(ctx, newBranch, commitMessage, files); err != nil {
		return nil, fmt.Errorf("error creating commit for branch[%s] for repo[%s] err [%s]", newBranch, userRepRef.String(), err)
	}

	pr, err := ur.PullRequests().Create(ctx, prTitle, newBranch, targetBranch, prDescription)
	if err != nil {
		return nil, fmt.Errorf("error creating pull request[%s] for branch[%s] for repo[%s] err [%s]", prTitle, newBranch, userRepRef.String(), err)
	}

	return pr, nil
}

func (h defaultGitProviderHandler) CreatePullRequestToOrgRepo(orgRepRef gitprovider.OrgRepositoryRef, targetBranch string, newBranch string, files []gitprovider.CommitFile, commitMessage string, prTitle string, prDescription string) (gitprovider.PullRequest, error) {
	provider, err := GithubProvider()
	if err != nil {
		return nil, err
	}
	ctx := context.Background()

	ur, err := provider.OrgRepositories().Get(ctx, orgRepRef)
	if err != nil {
		return nil, fmt.Errorf("error getting info for repo [%s] err [%s]", orgRepRef.String(), err)
	}

	if targetBranch == "" {
		targetBranch = *ur.Get().DefaultBranch
	}

	commits, err := ur.Commits().ListPage(ctx, targetBranch, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("error getting commits for repo[%s] err [%s]", orgRepRef.String(), err)
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("targetBranch[%s] does not exists", targetBranch)
	}

	latestCommit := commits[0]

	if err := ur.Branches().Create(ctx, newBranch, latestCommit.Get().Sha); err != nil {
		return nil, fmt.Errorf("error creating branch[%s] for repo[%s] err [%s]", newBranch, orgRepRef.String(), err)
	}

	if _, err := ur.Commits().Create(ctx, newBranch, commitMessage, files); err != nil {
		return nil, fmt.Errorf("error creating commit for branch[%s] for repo[%s] err [%s]", newBranch, orgRepRef.String(), err)
	}

	pr, err := ur.PullRequests().Create(ctx, prTitle, newBranch, targetBranch, prDescription)
	if err != nil {
		return nil, fmt.Errorf("error creating pull request[%s] for branch[%s] for repo[%s] err [%s]", prTitle, newBranch, orgRepRef.String(), err)
	}

	return pr, nil
}

func (h defaultGitProviderHandler) OauthConfig() OauthProviderConfig {
	gh := DefaultGithubProviderHandler{}

	return gh.OauthConfig()
}

func (h defaultGitProviderHandler) GetUser(ctx context.Context, token *oauth2.Token) (*User, error) {
	gh := DefaultGithubProviderHandler{}

	user, err := gh.GetUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("could not get github user: %w", err)
	}

	return &User{Email: user.Email}, nil
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

func waitUntilRepoCreated(ownerType ProviderAccountType, owner, name string) error {
	if err := utils.WaitUntil(os.Stdout, time.Second, time.Second*30, func() error {
		return GetRepoInfo(githubProvider, ownerType, owner, name)
	}); err != nil {
		return fmt.Errorf("could not verify repo existence %s", err)
	}
	return nil
}
