package gitproviders

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/utils"

	"github.com/fluxcd/go-git-providers/gitprovider"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type ProviderAccountType string

const (
	AccountTypeUser ProviderAccountType = "user"
	AccountTypeOrg  ProviderAccountType = "organization"
	DeployKeyName                       = "wego-deploy-key"

	defaultTimeout = time.Second * 30
)

var ErrRepositoryNoPermissionsOrDoesNotExist = errors.New("no permissions to access this repository or repository doesn't exists")

// GitProvider Handler
//counterfeiter:generate . GitProvider
type GitProvider interface {
	RepositoryExists(ctx context.Context, repoUrl RepoURL) (bool, error)
	DeployKeyExists(ctx context.Context, repoUrl RepoURL) (bool, error)
	GetDefaultBranch(ctx context.Context, repoUrl RepoURL) (string, error)
	GetRepoVisibility(ctx context.Context, repoUrl RepoURL) (*gitprovider.RepositoryVisibility, error)
	UploadDeployKey(ctx context.Context, repoUrl RepoURL, deployKey []byte) error
	CreatePullRequest(ctx context.Context, repoUrl RepoURL, prInfo PullRequestInfo) (gitprovider.PullRequest, error)
	GetCommits(ctx context.Context, repoUrl RepoURL, targetBranch string, pageSize int, pageToken int) ([]gitprovider.Commit, error)
	GetProviderDomain() string
	GetRepoDirFiles(ctx context.Context, repoUrl RepoURL, dirPath, targetBranch string) ([]*gitprovider.CommitFile, error)
	MergePullRequest(ctx context.Context, repoUrl RepoURL, pullRequestNumber int, commitMesage string) error
}

type PullRequestInfo struct {
	Title                     string
	Description               string
	CommitMessage             string
	TargetBranch              string
	NewBranch                 string
	SkipAddingFilesOnCreation bool
	Files                     []gitprovider.CommitFile
}

type AccountTypeGetter func(provider gitprovider.Client, domain string, owner string) (ProviderAccountType, error)

func New(config Config, owner string, getAccountType AccountTypeGetter) (GitProvider, error) {
	provider, domain, err := buildGitProvider(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build git provider: %w", err)
	}

	accountType, err := getAccountType(provider, domain, owner)
	if err != nil {
		return nil, err
	}

	if accountType == AccountTypeOrg {
		return orgGitProvider{
			domain:   domain,
			provider: provider,
		}, nil
	}

	return userGitProvider{
		domain:   domain,
		provider: provider,
	}, nil
}

func deployKeyExists(ctx context.Context, repo gitprovider.UserRepository) (bool, error) {
	_, err := repo.DeployKeys().Get(ctx, DeployKeyName)
	if err != nil && !strings.Contains(err.Error(), "key is already in use") {
		if errors.Is(err, gitprovider.ErrNotFound) {
			return false, nil
		} else {
			return false, fmt.Errorf("error getting deploy key %s: %s", DeployKeyName, err)
		}
	} else {
		return true, nil
	}
}

func uploadDeployKey(ctx context.Context, repo gitprovider.UserRepository, deployKeyInfo gitprovider.DeployKeyInfo) error {
	_, err := repo.DeployKeys().Create(ctx, deployKeyInfo)
	if err != nil {
		if errors.Is(err, gitprovider.ErrNotFound) {
			return ErrRepositoryNoPermissionsOrDoesNotExist
		}

		return fmt.Errorf("error uploading deploy key %s", err)
	}

	if err = utils.WaitUntil(os.Stdout, time.Second, defaultTimeout, func() error {
		_, err = repo.DeployKeys().Get(ctx, DeployKeyName)
		return err
	}); err != nil {
		return fmt.Errorf("error verifying deploy key %s: %s", DeployKeyName, err)
	}

	return nil
}

func createPullRequest(ctx context.Context, repo gitprovider.UserRepository, prInfo PullRequestInfo) (gitprovider.PullRequest, error) {
	repoInfo := repo.Get()

	if prInfo.TargetBranch == "" {
		prInfo.TargetBranch = *repoInfo.DefaultBranch
	}

	if prInfo.SkipAddingFilesOnCreation {
		pr, err := repo.PullRequests().Create(ctx, prInfo.Title, prInfo.NewBranch, prInfo.TargetBranch, prInfo.Description)
		if err != nil {
			return nil, fmt.Errorf("error creating pull request %s: %w", prInfo.Title, err)
		}

		return pr, nil
	}

	commits, err := repo.Commits().ListPage(ctx, prInfo.TargetBranch, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("error getting commits: %w", err)
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits on the target branch: %s", prInfo.TargetBranch)
	}

	latestCommit := commits[0]

	if err := repo.Branches().Create(ctx, prInfo.NewBranch, latestCommit.Get().Sha); err != nil {
		return nil, fmt.Errorf("error creating branch %s: %w", prInfo.NewBranch, err)
	}

	if _, err := repo.Commits().Create(ctx, prInfo.NewBranch, prInfo.CommitMessage, prInfo.Files); err != nil {
		return nil, fmt.Errorf("error creating commit %s: %w", prInfo.NewBranch, err)
	}

	pr, err := repo.PullRequests().Create(ctx, prInfo.Title, prInfo.NewBranch, prInfo.TargetBranch, prInfo.Description)
	if err != nil {
		return nil, fmt.Errorf("error creating pull request %s: %w", prInfo.Title, err)
	}

	return pr, nil
}

func getCommits(ctx context.Context, repo gitprovider.UserRepository, targetBranch string, pageSize int, pageToken int) ([]gitprovider.Commit, error) {
	// currently locking the commit list at 10. May discuss pagination options later.
	commits, err := repo.Commits().ListPage(ctx, targetBranch, pageSize, pageToken)
	if err != nil {
		if isEmptyRepoError(err) {
			return []gitprovider.Commit{}, nil
		}

		return nil, fmt.Errorf("error getting commits: %s", err)
	}

	return commits, nil
}

func getProviderDomain(providerID gitprovider.ProviderID) string {
	return string(GitProviderName(providerID)) + ".com"
}

func GetAccountType(provider gitprovider.Client, domain string, owner string) (ProviderAccountType, error) {
	_, err := provider.Organizations().Get(context.Background(), gitprovider.OrganizationRef{
		Domain:       domain,
		Organization: owner,
	})
	if err != nil {
		if errors.Is(err, gitprovider.ErrNotFound) || strings.Contains(err.Error(), gitprovider.ErrGroupNotFound.Error()) {
			return AccountTypeUser, nil
		}

		return "", fmt.Errorf("could not get account type %s", err)
	}

	return AccountTypeOrg, nil
}

func isEmptyRepoError(err error) bool {
	return strings.Contains(err.Error(), "409 Git Repository is empty")
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

func newUserRepositoryRef(domain, user, repoName string) gitprovider.UserRepositoryRef {
	return gitprovider.UserRepositoryRef{
		RepositoryName: repoName,
		UserRef: gitprovider.UserRef{
			Domain:    domain,
			UserLogin: user,
		},
	}
}
