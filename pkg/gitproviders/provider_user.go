package gitproviders

import (
	"context"
	"errors"
	"fmt"

	"github.com/fluxcd/go-git-providers/gitprovider"
)

type userGitProvider struct {
	domain   string
	provider gitprovider.Client
}

var _ GitProvider = userGitProvider{}

func (p userGitProvider) RepositoryExists(name string, owner string) (bool, error) {
	userRepoRef := gitprovider.UserRepositoryRef{
		UserRef:        gitprovider.UserRef{Domain: p.domain, UserLogin: owner},
		RepositoryName: name,
	}
	if _, err := p.provider.UserRepositories().Get(context.Background(), userRepoRef); err != nil {
		if errors.Is(err, gitprovider.ErrNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("could not get verify repository exists  %w", err)
	}

	return true, nil
}

func (p userGitProvider) DeployKeyExists(owner, repoName string) (bool, error) {
	ctx := context.Background()

	userRepo, err := p.getUserRepo(owner, repoName)
	if err != nil {
		return false, fmt.Errorf("error getting user repo reference for owner %s, repo %s, %s ", owner, repoName, err)
	}

	return deployKeyExists(ctx, userRepo)
}

func (p userGitProvider) UploadDeployKey(owner, repoName string, deployKey []byte) error {
	ctx := context.Background()

	userRepo, err := p.getUserRepo(owner, repoName)
	if err != nil {
		return fmt.Errorf("error getting user repo reference for owner %s, repo %s, %s ", owner, repoName, err)
	}

	fmt.Println("uploading deploy key")

	deployKeyInfo := gitprovider.DeployKeyInfo{
		Name:     deployKeyName,
		Key:      deployKey,
		ReadOnly: gitprovider.BoolVar(false),
	}

	return uploadDeployKey(ctx, userRepo, deployKeyInfo)
}

func (p userGitProvider) GetDefaultBranch(url string) (string, error) {
	repoInfoRef, err := p.getRepoInfoFromUrl(url)
	if err != nil {
		return "main", err
	}

	return *repoInfoRef.DefaultBranch, nil
}

func (p userGitProvider) GetRepoVisibility(url string) (*gitprovider.RepositoryVisibility, error) {
	repoInfoRef, err := p.getRepoInfoFromUrl(url)
	if err != nil {
		return nil, err
	}

	return repoInfoRef.Visibility, nil
}

func (p userGitProvider) getRepoInfoFromUrl(repoUrl string) (*gitprovider.RepositoryInfo, error) {
	normalizedUrl, err := NewNormalizedRepoURL(repoUrl)
	if err != nil {
		return nil, err
	}

	repo, err := p.getUserRepo(normalizedUrl.Owner(), normalizedUrl.RepositoryName())
	if err != nil {
		return nil, err
	}

	repoInfo := repo.Get()

	return &repoInfo, nil
}

func (p userGitProvider) getUserRepo(user string, repoName string) (gitprovider.UserRepository, error) {
	ctx := context.Background()

	repo, err := p.provider.UserRepositories().Get(ctx, newUserRepositoryRef(p.domain, user, repoName))
	if err != nil {
		return nil, fmt.Errorf("error getting user repository %w", err)
	}

	return repo, nil
}

func (p userGitProvider) CreatePullRequest(owner string, repoName string, targetBranch string, newBranch string, files []gitprovider.CommitFile, commitMsg string, prTitle string, prDescription string) (gitprovider.PullRequest, error) {
	ctx := context.Background()

	userRepo, err := p.getUserRepo(owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("error getting user repo for owner %s, repo %s, %s ", owner, repoName, err)
	}

	prInfo := PullRequestInfo{
		Title:         prTitle,
		Description:   prDescription,
		CommitMessage: commitMsg,
		TargetBranch:  targetBranch,
		NewBranch:     newBranch,
		Files:         files,
	}

	return createPullRequest(ctx, userRepo, prInfo)
}

func (p userGitProvider) GetCommits(owner string, repoName string, targetBranch string, pageSize int, pageToken int) ([]gitprovider.Commit, error) {
	ctx := context.Background()

	userRepo, err := p.getUserRepo(owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("error getting repo for owner %s, repo %s, %s ", owner, repoName, err)
	}

	return getCommits(ctx, userRepo, targetBranch, pageSize, pageToken)
}

func (p userGitProvider) GetProviderDomain() string {
	return getProviderDomain(p.provider.ProviderID())
}
