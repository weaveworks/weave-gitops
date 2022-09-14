package gitproviders

import (
	"context"

	"github.com/fluxcd/go-git-providers/gitprovider"
)

type dryrunProvider struct {
	provider GitProvider
}

func NewDryRun() (GitProvider, error) {
	provider, err := New(Config{
		Provider: GitProviderGitHub,
		Token:    "dummy",
	}, "", func(provider gitprovider.Client, domain, owner string) (ProviderAccountType, error) {
		return ProviderAccountType(GitProviderGitHub), nil
	})
	if err != nil {
		return nil, err
	}

	return &dryrunProvider{
		provider: provider,
	}, nil
}

func (p *dryrunProvider) RepositoryExists(_ context.Context, repoURL RepoURL) (bool, error) {
	return true, nil
}

func (p *dryrunProvider) DeployKeyExists(_ context.Context, repoURL RepoURL) (bool, error) {
	return true, nil
}

func (p *dryrunProvider) GetDefaultBranch(_ context.Context, repoURL RepoURL) (string, error) {
	return "<default-branch>", nil
}

func (p *dryrunProvider) GetRepoVisibility(_ context.Context, repoURL RepoURL) (*gitprovider.RepositoryVisibility, error) {
	return gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityPrivate), nil
}

func (p *dryrunProvider) UploadDeployKey(_ context.Context, repoURL RepoURL, deployKey []byte) error {
	return nil
}

func (p *dryrunProvider) CreatePullRequest(_ context.Context, repoURL RepoURL, prInfo PullRequestInfo) (gitprovider.PullRequest, error) {
	return nil, nil
}

func (p *dryrunProvider) GetCommits(_ context.Context, repoURL RepoURL, targetBranch string, pageSize int, pageToken int) ([]gitprovider.Commit, error) {
	return []gitprovider.Commit{}, nil
}

func (p *dryrunProvider) GetProviderDomain() string {
	return p.provider.GetProviderDomain()
}

func (p *dryrunProvider) GetRepoDirFiles(_ context.Context, repoURL RepoURL, dirPath, targetBranch string) ([]*gitprovider.CommitFile, error) {
	return nil, nil
}

func (p *dryrunProvider) MergePullRequest(ctx context.Context, repoURL RepoURL, pullRequestNumber int, commitMesage string) error {
	return nil
}
