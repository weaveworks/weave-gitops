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

func (p orgGitProvider) RepositoryExists(name string, owner string) (bool, error) {
	orgRef := gitprovider.OrgRepositoryRef{
		OrganizationRef: gitprovider.OrganizationRef{Domain: p.domain, Organization: owner},
		RepositoryName:  name,
	}
	if _, err := p.provider.OrgRepositories().Get(context.Background(), orgRef); err != nil {
		if errors.Is(err, gitprovider.ErrNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("could not get verify repository exists  %w", err)
	}

	return true, nil
}

func (p orgGitProvider) DeployKeyExists(owner, repoName string) (bool, error) {
	ctx := context.Background()

	orgRepo, err := p.getOrgRepo(owner, repoName)
	if err != nil {
		return false, fmt.Errorf("error getting org repo reference for owner %s, repo %s, %s", owner, repoName, err)
	}

	return deployKeyExists(ctx, orgRepo)
}

func (p orgGitProvider) UploadDeployKey(owner, repoName string, deployKey []byte) error {
	ctx := context.Background()

	orgRepo, err := p.getOrgRepo(owner, repoName)
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

func (p orgGitProvider) GetDefaultBranch(url string) (string, error) {
	repoInfoRef, err := p.getRepoInfoFromUrl(url)
	if err != nil {
		return "main", err
	}

	return *repoInfoRef.DefaultBranch, nil
}

func (p orgGitProvider) GetRepoVisibility(url string) (*gitprovider.RepositoryVisibility, error) {
	repoInfoRef, err := p.getRepoInfoFromUrl(url)
	if err != nil {
		return nil, err
	}

	return repoInfoRef.Visibility, nil
}

func (p orgGitProvider) getRepoInfoFromUrl(repoUrl string) (*gitprovider.RepositoryInfo, error) {
	normalizedUrl, err := NewNormalizedRepoURL(repoUrl)
	if err != nil {
		return nil, err
	}

	repoInfo, err := p.getRepoInfo(normalizedUrl.Owner(), normalizedUrl.RepositoryName())
	if err != nil {
		return nil, err
	}

	return repoInfo, nil
}

func (p orgGitProvider) getRepoInfo(owner string, repoName string) (*gitprovider.RepositoryInfo, error) {
	repo, err := p.getOrgRepo(owner, repoName)
	if err != nil {
		return nil, err
	}

	info := repo.Get()

	return &info, nil
}

func (p orgGitProvider) getOrgRepo(org string, repoName string) (gitprovider.OrgRepository, error) {
	orgRepoRef := newOrgRepositoryRef(p.domain, org, repoName)

	repo, err := p.provider.OrgRepositories().Get(context.Background(), orgRepoRef)
	if err != nil {
		return nil, fmt.Errorf("error getting org repository %w", err)
	}

	return repo, nil
}

func (p orgGitProvider) CreatePullRequest(owner string, repoName string, targetBranch string, newBranch string, files []gitprovider.CommitFile, commitMsg string, prTitle string, prDescription string) (gitprovider.PullRequest, error) {
	ctx := context.Background()

	orgRepo, err := p.getOrgRepo(owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("error getting org repo for owner %s, repo %s, %s ", owner, repoName, err)
	}

	prInfo := PullRequestInfo{
		Title:         prTitle,
		Description:   prDescription,
		CommitMessage: commitMsg,
		TargetBranch:  targetBranch,
		NewBranch:     newBranch,
		Files:         files,
	}

	return createPullRequest(ctx, orgRepo, prInfo)
}

func (p orgGitProvider) GetCommits(owner string, repoName string, targetBranch string, pageSize int, pageToken int) ([]gitprovider.Commit, error) {
	ctx := context.Background()

	orgRepo, err := p.getOrgRepo(owner, repoName)
	if err != nil {
		return nil, fmt.Errorf("error getting repo for owner %s, repo %s, %s ", owner, repoName, err)
	}

	return getCommits(ctx, orgRepo, targetBranch, pageSize, pageToken)
}

func (p orgGitProvider) GetProviderDomain() string {
	return getProviderDomain(p.provider.ProviderID())
}
