package gitproviders

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/utils"
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
	defer ctx.Done()

	orgRef := NewOrgRepositoryRef(p.domain, owner, repoName)

	orgRepo, err := p.provider.OrgRepositories().Get(ctx, orgRef)
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
}

func (p orgGitProvider) UploadDeployKey(owner, repoName string, deployKey []byte) error {
	deployKeyInfo := gitprovider.DeployKeyInfo{
		Name:     deployKeyName,
		Key:      deployKey,
		ReadOnly: gitprovider.BoolVar(false),
	}

	ctx := context.Background()
	defer ctx.Done()

	orgRef := NewOrgRepositoryRef(p.domain, owner, repoName)
	orgRepo, err := p.provider.OrgRepositories().Get(ctx, orgRef)

	if err != nil {
		return fmt.Errorf("error getting org repo reference for owner %s, repo %s, %s ", owner, repoName, err)
	}

	fmt.Println("uploading deploy key")

	_, err = orgRepo.DeployKeys().Create(ctx, deployKeyInfo)
	if err != nil {
		return fmt.Errorf("error uploading deploy key %s", err)
	}

	if err = utils.WaitUntil(os.Stdout, time.Second, defaultTimeout, func() error {
		_, err = orgRepo.DeployKeys().Get(ctx, deployKeyName)
		return err
	}); err != nil {
		return fmt.Errorf("error verifying deploy key %s existance for repo %s. %s", deployKeyName, repoName, err)
	}

	return nil
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
	owner, err := utils.GetOwnerFromUrl(repoUrl)
	if err != nil {
		return nil, err
	}

	repoInfo, err := p.getRepoInfo(owner, utils.UrlToRepoName(repoUrl))
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
	orgRepoRef := NewOrgRepositoryRef(p.domain, org, repoName)

	repo, err := p.provider.OrgRepositories().Get(context.Background(), orgRepoRef)
	if err != nil {
		return nil, fmt.Errorf("error getting org repository %w", err)
	}

	return repo, nil
}

func (p orgGitProvider) CreatePullRequest(owner string, repoName string, targetBranch string, newBranch string, files []gitprovider.CommitFile, commitMsg string, prTitle string, prDescription string) (gitprovider.PullRequest, error) {
	pr, err := p.createPullRequestToOrgRepo(owner, repoName, targetBranch, newBranch, files, commitMsg, prTitle, prDescription)
	if err != nil {
		return nil, fmt.Errorf("unable to create pull request: %w", err)
	}

	return pr, nil
}

func (p orgGitProvider) createPullRequestToOrgRepo(owner string, repoName string, targetBranch string, newBranch string, files []gitprovider.CommitFile, commitMessage string, prTitle string, prDescription string) (gitprovider.PullRequest, error) {
	ctx := context.Background()
	orgRepRef := NewOrgRepositoryRef(p.GetProviderDomain(), owner, repoName)

	ur, err := p.provider.OrgRepositories().Get(ctx, orgRepRef)
	if err != nil {
		return nil, fmt.Errorf("error getting info for repo [%s] err [%s]", orgRepRef.String(), err)
	}

	if targetBranch == "" {
		targetBranch = *ur.Get().DefaultBranch
	}

	commits, err := ur.Commits().ListPage(ctx, targetBranch, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("error getting commits for repo [%s] err [%s]", orgRepRef.String(), err)
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("targetBranch [%s] does not exists", targetBranch)
	}

	latestCommit := commits[0]

	if err := ur.Branches().Create(ctx, newBranch, latestCommit.Get().Sha); err != nil {
		return nil, fmt.Errorf("error creating branch [%s] for repo [%s] err [%s]", newBranch, orgRepRef.String(), err)
	}

	if _, err := ur.Commits().Create(ctx, newBranch, commitMessage, files); err != nil {
		return nil, fmt.Errorf("error creating commit for branch [%s] for repo [%s] err [%s]", newBranch, orgRepRef.String(), err)
	}

	pr, err := ur.PullRequests().Create(ctx, prTitle, newBranch, targetBranch, prDescription)
	if err != nil {
		return nil, fmt.Errorf("error creating pull request [%s] for branch [%s] for repo [%s] err [%s]", prTitle, newBranch, orgRepRef.String(), err)
	}

	return pr, nil
}

func (p orgGitProvider) GetCommits(owner string, repoName string, targetBranch string, pageSize int, pageToken int) ([]gitprovider.Commit, error) {
	commits, err := p.getCommitsFromOrgRepo(owner, repoName, targetBranch, pageSize, pageToken)
	if err != nil {
		return nil, fmt.Errorf("unable to get commits: %w", err)
	}

	return commits, nil
}

func (p orgGitProvider) getCommitsFromOrgRepo(owner string, repoName string, targetBranch string, pageSize int, pageToken int) ([]gitprovider.Commit, error) {
	ctx := context.Background()
	orgRepoRef := NewOrgRepositoryRef(p.GetProviderDomain(), owner, repoName)

	ur, err := p.provider.OrgRepositories().Get(ctx, orgRepoRef)
	if err != nil {
		return nil, fmt.Errorf("error getting info for repo [%s] err [%s]", orgRepoRef.String(), err)
	}

	// currently locking the commit list at 10. May discuss pagination options later.
	commits, err := ur.Commits().ListPage(ctx, targetBranch, pageSize, pageToken)
	if err != nil {
		if isEmptyRepoError(err) {
			return []gitprovider.Commit{}, nil
		}

		return nil, fmt.Errorf("error getting commits for repo [%s] err [%s]", orgRepoRef.String(), err)
	}

	return commits, nil
}

func (p orgGitProvider) GetProviderDomain() string {
	return string(GitProviderName(p.provider.ProviderID())) + ".com"
}
