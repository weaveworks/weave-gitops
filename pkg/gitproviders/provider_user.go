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

func (p userGitProvider) RepositoryExists(ctx context.Context, repoUrl RepoURL) (bool, error) {
	userRepoRef := gitprovider.UserRepositoryRef{
		UserRef:        gitprovider.UserRef{Domain: p.domain, UserLogin: repoUrl.Owner()},
		RepositoryName: repoUrl.RepositoryName(),
	}
	if _, err := p.provider.UserRepositories().Get(ctx, userRepoRef); err != nil {
		if errors.Is(err, gitprovider.ErrNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("could not get verify repository exists  %w", err)
	}

	return true, nil
}

func (p userGitProvider) DeployKeyExists(ctx context.Context, repoUrl RepoURL) (bool, error) {
	userRepo, err := p.getUserRepo(ctx, repoUrl)
	if err != nil {
		return false, fmt.Errorf("error getting user repo reference for owner %s, repo %s, %w", repoUrl.Owner(), repoUrl.RepositoryName(), err)
	}

	return deployKeyExists(ctx, userRepo)
}

func (p userGitProvider) UploadDeployKey(ctx context.Context, repoUrl RepoURL, deployKey []byte) error {
	userRepo, err := p.getUserRepo(ctx, repoUrl)
	if err != nil {
		return fmt.Errorf("error getting user repo reference for owner %s, repo %s, %w", repoUrl.Owner(), repoUrl.RepositoryName(), err)
	}

	deployKeyInfo := gitprovider.DeployKeyInfo{
		Name:     DeployKeyName,
		Key:      deployKey,
		ReadOnly: gitprovider.BoolVar(false),
	}

	return uploadDeployKey(ctx, userRepo, deployKeyInfo)
}

func (p userGitProvider) GetDefaultBranch(ctx context.Context, repoUrl RepoURL) (string, error) {
	repoInfoRef, err := p.getRepoInfoFromUrl(ctx, repoUrl)
	if err != nil {
		return "main", err
	}

	return *repoInfoRef.DefaultBranch, nil
}

func (p userGitProvider) GetRepoVisibility(ctx context.Context, repoUrl RepoURL) (*gitprovider.RepositoryVisibility, error) {
	repoInfoRef, err := p.getRepoInfoFromUrl(ctx, repoUrl)
	if err != nil {
		return nil, err
	}

	return repoInfoRef.Visibility, nil
}

func (p userGitProvider) getRepoInfoFromUrl(ctx context.Context, repoUrl RepoURL) (*gitprovider.RepositoryInfo, error) {
	repo, err := p.getUserRepo(ctx, repoUrl)
	if err != nil {
		return nil, err
	}

	repoInfo := repo.Get()

	return &repoInfo, nil
}

func (p userGitProvider) getUserRepo(ctx context.Context, repoUrl RepoURL) (gitprovider.UserRepository, error) {
	repo, err := p.provider.UserRepositories().Get(ctx, newUserRepositoryRef(p.domain, repoUrl.Owner(), repoUrl.RepositoryName()))
	if err != nil {
		return nil, fmt.Errorf("error getting user repository %w", err)
	}

	return repo, nil
}

func (p userGitProvider) CreatePullRequest(ctx context.Context, repoUrl RepoURL, prInfo PullRequestInfo) (gitprovider.PullRequest, error) {
	userRepo, err := p.getUserRepo(ctx, repoUrl)
	if err != nil {
		return nil, fmt.Errorf("error getting user repo for owner %s, repo %s, %w", repoUrl.Owner(), repoUrl.RepositoryName(), err)
	}

	return createPullRequest(ctx, userRepo, prInfo)
}

func (p userGitProvider) GetCommits(ctx context.Context, repoUrl RepoURL, targetBranch string, pageSize int, pageToken int) ([]gitprovider.Commit, error) {
	userRepo, err := p.getUserRepo(ctx, repoUrl)
	if err != nil {
		return nil, fmt.Errorf("error getting repo for owner %s, repo %s, %w", repoUrl.Owner(), repoUrl.RepositoryName(), err)
	}

	return getCommits(ctx, userRepo, targetBranch, pageSize, pageToken)
}

func (p userGitProvider) GetProviderDomain() string {
	return getProviderDomain(p.provider.ProviderID())
}

// GetRepoFiles returns the files found in a directory. The targetPath must point to a directory, not a file.
// Note that the current implementation only gets an end subdirectory. It does not get multiple directories recursively. See https://github.com/fluxcd/go-git-providers/issues/143.
func (p userGitProvider) GetRepoFiles(ctx context.Context, repoUrl RepoURL, targetPath, targetBranch string) ([]*gitprovider.CommitFile, error) {
	repo, err := p.getUserRepo(ctx, repoUrl)
	if err != nil {
		return nil, err
	}
	files, err := repo.Files().Get(ctx, targetPath, targetBranch)
	if err != nil {
		return nil, err
	}
	return files, nil
}

// MergePullRequest merges a pull request given the repository's URL and the PR's number with a merge method, and a commit message.
func (p userGitProvider) MergePullRequest(ctx context.Context, repoUrl RepoURL, pullRequestNumber int, mergeMethod gitprovider.MergeMethod, commitMesage string) error {
	repo, err := p.getUserRepo(ctx, repoUrl)
	if err != nil {
		return err
	}
	return repo.PullRequests().Merge(ctx, pullRequestNumber, mergeMethod, commitMesage)
}
