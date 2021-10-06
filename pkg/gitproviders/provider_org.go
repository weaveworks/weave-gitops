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

func (p orgGitProvider) RepositoryExists(ctx context.Context, name string, owner string) (bool, error) {
	orgRef := gitprovider.OrgRepositoryRef{
		OrganizationRef: gitprovider.OrganizationRef{Domain: p.domain, Organization: owner},
		RepositoryName:  name,
	}
	if _, err := p.provider.OrgRepositories().Get(ctx, orgRef); err != nil {
		if errors.Is(err, gitprovider.ErrNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("could not get verify repository exists  %w", err)
	}

	return true, nil
}

func (p orgGitProvider) DeployKeyExists(ctx context.Context, owner, repoName string) (bool, error) {
	orgRepo, err := p.getOrgRepo(ctx, owner, repoName)
	if err != nil {
		return false, fmt.Errorf("error getting org repo reference for owner %s, repo %s, %s", owner, repoName, err)
	}

	return deployKeyExists(ctx, orgRepo)
}

func (p orgGitProvider) UploadDeployKey(ctx context.Context, owner, repoName string, deployKey []byte) error {
	orgRepo, err := p.getOrgRepo(ctx, owner, repoName)
	if err != nil {
		return fmt.Errorf("error getting org repo reference for owner %s, repo %s, %s ", owner, repoName, err)
	}

	fmt.Println("uploading deploy key")

	deployKeyInfo := gitprovider.DeployKeyInfo{
		Name:     deployKeyName,
		Key:      deployKey,
		ReadOnly: gitprovider.BoolVar(false),
	}

	return uploadDeployKey(ctx, orgRepo, deployKeyInfo)
}

func (p orgGitProvider) GetDefaultBranch(ctx context.Context, url string) (string, error) {
	repoInfoRef, err := p.getRepoInfoFromUrl(ctx, url)
	if err != nil {
		return "main", err
	}

	return *repoInfoRef.DefaultBranch, nil
}

func (p orgGitProvider) GetRepoVisibility(ctx context.Context, url string) (*gitprovider.RepositoryVisibility, error) {
	repoInfoRef, err := p.getRepoInfoFromUrl(ctx, url)
	if err != nil {
		return nil, err
	}

	return repoInfoRef.Visibility, nil
}

func (p orgGitProvider) getRepoInfoFromUrl(ctx context.Context, repoUrl string) (*gitprovider.RepositoryInfo, error) {
	normalizedUrl, err := NewNormalizedRepoURL(repoUrl)
	if err != nil {
		return nil, err
	}

	repoInfo, err := p.getRepoInfo(ctx, normalizedUrl.Owner(), normalizedUrl.RepositoryName())
	if err != nil {
		return nil, err
	}

	return repoInfo, nil
}

func (p orgGitProvider) getRepoInfo(ctx context.Context, owner string, repoName string) (*gitprovider.RepositoryInfo, error) {
	repo, err := p.getOrgRepo(ctx, owner, repoName)
	if err != nil {
		return nil, err
	}

	info := repo.Get()

	return &info, nil
}

func (p orgGitProvider) getOrgRepo(ctx context.Context, org string, repoName string) (gitprovider.OrgRepository, error) {
	orgRepoRef := newOrgRepositoryRef(p.domain, org, repoName)

	repo, err := p.provider.OrgRepositories().Get(ctx, orgRepoRef)
	if err != nil {
		return nil, fmt.Errorf("error getting org repository %w", err)
	}

	return repo, nil
}

func (p orgGitProvider) CreatePullRequest(ctx context.Context, owner string, repoName string, prInfo PullRequestInfo) (gitprovider.PullRequest, error) {
	orgRepo, err := p.getOrgRepo(ctx, owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("error getting org repo for owner %s, repo %s, %s ", owner, repoName, err)
	}

	return createPullRequest(ctx, orgRepo, prInfo)
}

func (p orgGitProvider) GetCommits(ctx context.Context, owner string, repoName string, targetBranch string, pageSize int, pageToken int) ([]gitprovider.Commit, error) {
	orgRepo, err := p.getOrgRepo(ctx, owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("error getting repo for owner %s, repo %s, %s ", owner, repoName, err)
	}

	return getCommits(ctx, orgRepo, targetBranch, pageSize, pageToken)
}

func (p orgGitProvider) GetProviderDomain() string {
	return getProviderDomain(p.provider.ProviderID())
}
