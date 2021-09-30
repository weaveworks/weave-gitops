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
		return false, err
	}

	return true, nil
}

func (p userGitProvider) DeployKeyExists(owner, repoName string) (bool, error) {
	ctx := context.Background()
	defer ctx.Done()

	userRef := NewUserRepositoryRef(p.domain, owner, repoName)
	userRepo, err := p.provider.UserRepositories().Get(ctx, userRef)

	if err != nil {
		return false, fmt.Errorf("error getting user repo reference for owner %s, repo %s, %s ", owner, repoName, err)
	}

	_, err = userRepo.DeployKeys().Get(ctx, deployKeyName)
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

func (p userGitProvider) UploadDeployKey(owner, repoName string, deployKey []byte) error {
	deployKeyInfo := gitprovider.DeployKeyInfo{
		Name:     deployKeyName,
		Key:      deployKey,
		ReadOnly: gitprovider.BoolVar(false),
	}

	ctx := context.Background()
	defer ctx.Done()

	userRef := NewUserRepositoryRef(p.domain, owner, repoName)
	userRepo, err := p.provider.UserRepositories().Get(ctx, userRef)

	if err != nil {
		return fmt.Errorf("error getting user repo reference for owner %s, repo %s, %s ", owner, repoName, err)
	}

	fmt.Println("uploading deploy key")

	_, err = userRepo.DeployKeys().Create(ctx, deployKeyInfo)
	if err != nil {
		return fmt.Errorf("error uploading deploy key %s", err)
	}

	if err = utils.WaitUntil(os.Stdout, time.Second, defaultTimeout, func() error {
		_, err = userRepo.DeployKeys().Get(ctx, deployKeyName)
		return err
	}); err != nil {
		return fmt.Errorf("error verifying deploy key %s existance for repo %s. %s", deployKeyName, repoName, err)
	}

	return nil
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
	owner, err := utils.GetOwnerFromUrl(repoUrl)
	if err != nil {
		return nil, err
	}

	repo, err := p.getUserRepo(owner, utils.UrlToRepoName(repoUrl))
	if err != nil {
		return nil, err
	}

	repoInfo := repo.Get()

	return &repoInfo, nil
}

func (p userGitProvider) getUserRepo(user string, repoName string) (gitprovider.UserRepository, error) {
	ctx := context.Background()
	defer ctx.Done()

	userRepoRef := NewUserRepositoryRef(p.domain, user, repoName)

	repo, err := p.provider.UserRepositories().Get(ctx, userRepoRef)
	if err != nil {
		return nil, fmt.Errorf("error getting user repository %w", err)
	}

	return repo, nil
}

func (p userGitProvider) CreatePullRequest(owner string, repoName string, targetBranch string, newBranch string, files []gitprovider.CommitFile, commitMsg string, prTitle string, prDescription string) (gitprovider.PullRequest, error) {
	pr, err := p.createPullRequestToUserRepo(owner, repoName, targetBranch, newBranch, files, commitMsg, prTitle, prDescription)
	if err != nil {
		return nil, fmt.Errorf("unable to create pull request: %w", err)
	}

	return pr, nil
}

func (p userGitProvider) createPullRequestToUserRepo(owner string, repoName string, targetBranch string, newBranch string, files []gitprovider.CommitFile, commitMessage string, prTitle string, prDescription string) (gitprovider.PullRequest, error) {
	ctx := context.Background()
	userRepoRef := NewUserRepositoryRef(p.GetProviderDomain(), owner, repoName)

	ur, err := p.provider.UserRepositories().Get(ctx, userRepoRef)
	if err != nil {
		return nil, fmt.Errorf("error getting info for repo [%s] err [%s]", userRepoRef.String(), err)
	}

	if targetBranch == "" {
		targetBranch = *ur.Get().DefaultBranch
	}

	commits, err := ur.Commits().ListPage(ctx, targetBranch, 1, 0)
	if err != nil {
		return nil, fmt.Errorf("error getting commits for repo[%s] err [%s]", userRepoRef.String(), err)
	}

	if len(commits) == 0 {
		return nil, fmt.Errorf("targetBranch [%s] does not exists", targetBranch)
	}

	latestCommit := commits[0]

	if err := ur.Branches().Create(ctx, newBranch, latestCommit.Get().Sha); err != nil {
		return nil, fmt.Errorf("error creating branch [%s] for repo [%s] err [%s]", newBranch, userRepoRef.String(), err)
	}

	if _, err := ur.Commits().Create(ctx, newBranch, commitMessage, files); err != nil {
		return nil, fmt.Errorf("error creating commit for branch [%s] for repo [%s] err [%s]", newBranch, userRepoRef.String(), err)
	}

	pr, err := ur.PullRequests().Create(ctx, prTitle, newBranch, targetBranch, prDescription)
	if err != nil {
		return nil, fmt.Errorf("error creating pull request [%s] for branch [%s] for repo [%s] err [%s]", prTitle, newBranch, userRepoRef.String(), err)
	}

	return pr, nil
}

func (p userGitProvider) GetCommits(owner string, repoName string, targetBranch string, pageSize int, pageToken int) ([]gitprovider.Commit, error) {
	commits, err := p.getCommitsFromUserRepo(owner, repoName, targetBranch, pageSize, pageToken)
	if err != nil {
		return nil, fmt.Errorf("unable to get commits: %w", err)
	}

	return commits, nil
}

func (p userGitProvider) getCommitsFromUserRepo(owner string, repoName string, targetBranch string, pageSize int, pageToken int) ([]gitprovider.Commit, error) {
	ctx := context.Background()
	userRepoRef := NewUserRepositoryRef(p.GetProviderDomain(), owner, repoName)

	ur, err := p.provider.UserRepositories().Get(ctx, userRepoRef)
	if err != nil {
		return nil, fmt.Errorf("error getting info for repo [%s] err [%s]", userRepoRef.String(), err)
	}

	// currently locking the commit list at 10. May discuss pagination options later.
	commits, err := ur.Commits().ListPage(ctx, targetBranch, pageSize, pageToken)
	if err != nil {
		if isEmptyRepoError(err) {
			return []gitprovider.Commit{}, nil
		}

		return nil, fmt.Errorf("error getting commits for repo [%s] err [%s]", userRepoRef.String(), err)
	}

	return commits, nil
}

func (p userGitProvider) GetProviderDomain() string {
	return string(GitProviderName(p.provider.ProviderID())) + ".com"
}
