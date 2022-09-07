package gitproviders

import (
	"context"
	"errors"
	"fmt"

	"github.com/fluxcd/go-git-providers/gitprovider"
)

type orgGitProvider struct {
	domain   string
	provider gitprovider.Client
}

var _ GitProvider = orgGitProvider{}

func (p orgGitProvider) RepositoryExists(ctx context.Context, repoURL RepoURL) (bool, error) {
	orgRef := gitprovider.OrgRepositoryRef{
		OrganizationRef: gitprovider.OrganizationRef{Domain: p.domain, Organization: repoURL.Owner()},
		RepositoryName:  repoURL.RepositoryName(),
	}
	if _, err := p.provider.OrgRepositories().Get(ctx, orgRef); err != nil {
		if errors.Is(err, gitprovider.ErrNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("could not get verify repository exists  %w", err)
	}

	return true, nil
}

func (p orgGitProvider) DeployKeyExists(ctx context.Context, repoURL RepoURL) (bool, error) {
	orgRepo, err := p.getOrgRepo(ctx, repoURL)
	if err != nil {
		return false, fmt.Errorf("error getting org repo reference for owner %s, repo %s, %s", repoURL.Owner(), repoURL.RepositoryName(), err)
	}

	return deployKeyExists(ctx, orgRepo)
}

func (p orgGitProvider) UploadDeployKey(ctx context.Context, repoURL RepoURL, deployKey []byte) error {
	orgRepo, err := p.getOrgRepo(ctx, repoURL)
	if err != nil {
		return fmt.Errorf("error getting org repo reference for owner %s, repo %s, %w", repoURL.Owner(), repoURL.RepositoryName(), err)
	}

	deployKeyInfo := gitprovider.DeployKeyInfo{
		Name:     DeployKeyName,
		Key:      deployKey,
		ReadOnly: gitprovider.BoolVar(false),
	}

	return uploadDeployKey(ctx, orgRepo, deployKeyInfo)
}

func (p orgGitProvider) GetDefaultBranch(ctx context.Context, repoURL RepoURL) (string, error) {
	repoInfoRef, err := p.getRepoInfoFromURL(ctx, repoURL)
	if err != nil {
		return "main", err
	}

	return *repoInfoRef.DefaultBranch, nil
}

func (p orgGitProvider) GetRepoVisibility(ctx context.Context, repoURL RepoURL) (*gitprovider.RepositoryVisibility, error) {
	repoInfoRef, err := p.getRepoInfoFromURL(ctx, repoURL)
	if err != nil {
		return nil, err
	}

	return repoInfoRef.Visibility, nil
}

func (p orgGitProvider) getRepoInfoFromURL(ctx context.Context, repoURL RepoURL) (*gitprovider.RepositoryInfo, error) {
	repoInfo, err := p.getRepoInfo(ctx, repoURL)
	if err != nil {
		return nil, err
	}

	return repoInfo, nil
}

func (p orgGitProvider) getRepoInfo(ctx context.Context, repoURL RepoURL) (*gitprovider.RepositoryInfo, error) {
	repo, err := p.getOrgRepo(ctx, repoURL)
	if err != nil {
		return nil, err
	}

	info := repo.Get()

	return &info, nil
}

func (p orgGitProvider) getOrgRepo(ctx context.Context, repoURL RepoURL) (gitprovider.OrgRepository, error) {
	orgRepoRef := NewOrgRepositoryRef(p.domain, repoURL.Owner(), repoURL.RepositoryName())

	repo, err := p.provider.OrgRepositories().Get(context.Background(), orgRepoRef)
	if err != nil {
		return nil, fmt.Errorf("error getting org repository %w", err)
	}

	return repo, nil
}

func (p orgGitProvider) CreatePullRequest(ctx context.Context, repoURL RepoURL, prInfo PullRequestInfo) (gitprovider.PullRequest, error) {
	orgRepo, err := p.getOrgRepo(ctx, repoURL)
	if err != nil {
		return nil, fmt.Errorf("error getting org repo for owner %s, repo %s, %w", repoURL.Owner(), repoURL.RepositoryName(), err)
	}

	return createPullRequest(ctx, orgRepo, prInfo)
}

func (p orgGitProvider) GetCommits(ctx context.Context, repoURL RepoURL, targetBranch string, pageSize int, pageToken int) ([]gitprovider.Commit, error) {
	orgRepo, err := p.getOrgRepo(ctx, repoURL)
	if err != nil {
		return nil, fmt.Errorf("error getting repo for owner %s, repo %s, %w", repoURL.Owner(), repoURL.RepositoryName(), err)
	}

	return getCommits(ctx, orgRepo, targetBranch, pageSize, pageToken)
}

func (p orgGitProvider) GetProviderDomain() string {
	return getProviderDomain(p.provider.ProviderID())
}

// GetRepoDirFiles returns the files found in the subdirectory of a repository.
// Note that the current implementation only gets an end subdirectory. It does not get multiple directories recursively. See https://github.com/fluxcd/go-git-providers/issues/143.
func (p orgGitProvider) GetRepoDirFiles(ctx context.Context, repoURL RepoURL, dirPath, targetBranch string) ([]*gitprovider.CommitFile, error) {
	repo, err := p.getOrgRepo(ctx, repoURL)
	if err != nil {
		return nil, err
	}

	files, err := repo.Files().Get(ctx, dirPath, targetBranch)
	if err != nil {
		return nil, err
	}

	return files, nil
}

// MergePullRequest merges a pull request given the repository's URL and the PR's number with a commit message.
func (p orgGitProvider) MergePullRequest(ctx context.Context, repoURL RepoURL, pullRequestNumber int, commitMesage string) error {
	repo, err := p.getOrgRepo(ctx, repoURL)
	if err != nil {
		return err
	}

	return repo.PullRequests().Merge(ctx, pullRequestNumber, gitprovider.MergeMethodMerge, commitMesage)
}
