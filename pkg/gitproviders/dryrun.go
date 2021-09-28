package gitproviders

import "github.com/fluxcd/go-git-providers/gitprovider"

type dryrunProvider struct {
	provider GitProvider
}

func NewDryRun() (GitProvider, error) {
	provider, err := New(Config{
		Provider: GitProviderGitHub,
		Token:    "dummy",
	})
	if err != nil {
		return nil, err
	}

	return &dryrunProvider{
		provider: provider,
	}, nil
}

func (p *dryrunProvider) CreateRepository(name string, owner string, private bool) error {
	return nil
}

func (p *dryrunProvider) RepositoryExists(name string, owner string) (bool, error) {
	return true, nil
}

func (p *dryrunProvider) DeployKeyExists(owner, repoName string) (bool, error) {
	return true, nil
}

func (p *dryrunProvider) GetRepoInfo(_ ProviderAccountType, owner string, repoName string) (*gitprovider.RepositoryInfo, error) {
	return &gitprovider.RepositoryInfo{
		DefaultBranch: gitprovider.StringVar("<default-branch>"),
		Visibility:    gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityPrivate),
	}, nil
}

func (p *dryrunProvider) GetRepoInfoFromUrl(url string) (*gitprovider.RepositoryInfo, error) {
	return &gitprovider.RepositoryInfo{
		DefaultBranch: gitprovider.StringVar("<default-branch>"),
		Visibility:    gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityPrivate),
	}, nil
}

func (p *dryrunProvider) GetDefaultBranch(url string) (string, error) {
	return "<default-branch>", nil
}

func (p *dryrunProvider) GetRepoVisibility(url string) (*gitprovider.RepositoryVisibility, error) {
	return gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityPrivate), nil
}

func (p *dryrunProvider) UploadDeployKey(owner, repoName string, deployKey []byte) error {
	return nil
}

func (p *dryrunProvider) CreatePullRequestToUserRepo(userRepRef gitprovider.UserRepositoryRef, targetBranch string, newBranch string, files []gitprovider.CommitFile, commitMessage string, prTitle string, prDescription string) (gitprovider.PullRequest, error) {
	return nil, nil
}

func (p *dryrunProvider) CreatePullRequestToOrgRepo(orgRepRef gitprovider.OrgRepositoryRef, targetBranch string, newBranch string, files []gitprovider.CommitFile, commitMessage string, prTitle string, prDescription string) (gitprovider.PullRequest, error) {
	return nil, nil
}

func (p *dryrunProvider) GetCommitsFromUserRepo(userRepRef gitprovider.UserRepositoryRef, targetBranch string, pageSize int, pageToken int) ([]gitprovider.Commit, error) {
	return []gitprovider.Commit{}, nil
}

func (p *dryrunProvider) GetCommitsFromOrgRepo(orgRepRef gitprovider.OrgRepositoryRef, targetBranch string, pageSize int, pageToken int) ([]gitprovider.Commit, error) {
	return []gitprovider.Commit{}, nil
}

func (p *dryrunProvider) GetAccountType(owner string) (ProviderAccountType, error) {
	return AccountTypeUser, nil
}

func (p *dryrunProvider) GetProviderDomain() string {
	return p.provider.GetProviderDomain()
}
