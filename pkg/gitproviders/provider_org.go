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

func (p orgGitProvider) RepositoryExists(ctx context.Context, repoUrl RepoURL) (bool, error) {
	orgRef := gitprovider.OrgRepositoryRef{
		OrganizationRef: gitprovider.OrganizationRef{Domain: p.domain, Organization: repoUrl.Owner()},
		RepositoryName:  repoUrl.RepositoryName(),
	}
	if _, err := p.provider.OrgRepositories().Get(ctx, orgRef); err != nil {
		if errors.Is(err, gitprovider.ErrNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("could not get verify repository exists  %w", err)
	}

	return true, nil
}

func (p orgGitProvider) DeployKeyExists(ctx context.Context, repoUrl RepoURL) (bool, error) {
	orgRepo, err := p.getOrgRepo(ctx, repoUrl)
	if err != nil {
		return false, fmt.Errorf("error getting org repo reference for owner %s, repo %s, %s", repoUrl.Owner(), repoUrl.RepositoryName(), err)
	}

	return deployKeyExists(ctx, orgRepo)
}

func (p orgGitProvider) UploadDeployKey(ctx context.Context, repoUrl RepoURL, deployKey []byte) error {
	orgRepo, err := p.getOrgRepo(ctx, repoUrl)
	if err != nil {
		return fmt.Errorf("error getting org repo reference for owner %s, repo %s, %w", repoUrl.Owner(), repoUrl.RepositoryName(), err)
	}

	deployKeyInfo := gitprovider.DeployKeyInfo{
		Name:     DeployKeyName,
		Key:      deployKey,
		ReadOnly: gitprovider.BoolVar(false),
	}

	return uploadDeployKey(ctx, orgRepo, deployKeyInfo)
}

func (p orgGitProvider) GetDefaultBranch(ctx context.Context, repoUrl RepoURL) (string, error) {
	repoInfoRef, err := p.getRepoInfoFromUrl(ctx, repoUrl)
	if err != nil {
		return "main", err
	}

	return *repoInfoRef.DefaultBranch, nil
}

func (p orgGitProvider) GetRepoVisibility(ctx context.Context, repoUrl RepoURL) (*gitprovider.RepositoryVisibility, error) {
	repoInfoRef, err := p.getRepoInfoFromUrl(ctx, repoUrl)
	if err != nil {
		return nil, err
	}

	return repoInfoRef.Visibility, nil
}

func (p orgGitProvider) getRepoInfoFromUrl(ctx context.Context, repoUrl RepoURL) (*gitprovider.RepositoryInfo, error) {
	repoInfo, err := p.getRepoInfo(ctx, repoUrl)
	if err != nil {
		return nil, err
	}

	return repoInfo, nil
}

func (p orgGitProvider) getRepoInfo(ctx context.Context, repoUrl RepoURL) (*gitprovider.RepositoryInfo, error) {
	repo, err := p.getOrgRepo(ctx, repoUrl)
	if err != nil {
		return nil, err
	}

	info := repo.Get()

	return &info, nil
}

func (p orgGitProvider) getOrgRepo(ctx context.Context, repoUrl RepoURL) (gitprovider.OrgRepository, error) {
	orgRepoRef := NewOrgRepositoryRef(p.domain, repoUrl.Owner(), repoUrl.RepositoryName())

	repo, err := p.provider.OrgRepositories().Get(context.Background(), orgRepoRef)
	if err != nil {
		return nil, fmt.Errorf("error getting org repository %w", err)
	}

	return repo, nil
}

func (p orgGitProvider) CreatePullRequest(ctx context.Context, repoUrl RepoURL, prInfo PullRequestInfo) (gitprovider.PullRequest, error) {
	orgRepo, err := p.getOrgRepo(ctx, repoUrl)
	if err != nil {
		return nil, fmt.Errorf("error getting org repo for owner %s, repo %s, %w", repoUrl.Owner(), repoUrl.RepositoryName(), err)
	}

	return createPullRequest(ctx, orgRepo, prInfo)
}

func (p orgGitProvider) GetCommits(ctx context.Context, repoUrl RepoURL, targetBranch string, pageSize int, pageToken int) ([]gitprovider.Commit, error) {
	orgRepo, err := p.getOrgRepo(ctx, repoUrl)
	if err != nil {
		return nil, fmt.Errorf("error getting repo for owner %s, repo %s, %w", repoUrl.Owner(), repoUrl.RepositoryName(), err)
	}

	return getCommits(ctx, orgRepo, targetBranch, pageSize, pageToken)
}

func (p orgGitProvider) GetProviderDomain() string {
	return getProviderDomain(p.provider.ProviderID())
}

func (p orgGitProvider) GetRepoFiles(ctx context.Context, repoUrl RepoURL, targetPath, targetBranch string) ([]*gitprovider.CommitFile, error) {
	repo, err := p.getOrgRepo(ctx, repoUrl)
	if err != nil {
		return nil, err
	}
	files, err := repo.Files().Get(ctx, targetPath, targetBranch)
	if err != nil {
		return nil, err
	}
	return files, nil
}
