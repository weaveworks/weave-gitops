package git_providers

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/dnaeon/go-vcr/recorder"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/stretchr/testify/assert"

	"github.com/fluxcd/go-git-providers/gitprovider"
)

var githubClient, gitlabClient gitprovider.Client

func SetRecorder(recorder *recorder.Recorder) gitprovider.ChainableRoundTripperFunc {
	return func(transport http.RoundTripper) http.RoundTripper {
		recorder.SetTransport(transport)
		return recorder
	}
}

func NewRecorder(provider string) (*recorder.Recorder, error) {
	return recorder.New(fmt.Sprintf("./cache/%s", provider))
}

func TestMain(m *testing.M) {

	cacheGithubRecorder, err := NewRecorder("github")
	if err != nil {
		panic(err)
	}

	cacheGitlabRecorder, err := NewRecorder("gitlab")
	if err != nil {
		panic(err)
	}

	githubClient, err = newGithubTestClient(SetRecorder(cacheGithubRecorder))
	if err != nil {
		panic(err)
	}
	gitlabClient, err = newGitlabTestClient(SetRecorder(cacheGitlabRecorder))
	if err != nil {
		panic(err)
	}

	rand.Seed(time.Now().UnixNano())

	exitCode := m.Run()

	err = cacheGithubRecorder.Stop()
	if err != nil {
		panic(err)
	}

	err = cacheGitlabRecorder.Stop()
	if err != nil {
		panic(err)
	}

	os.Exit(exitCode)

}

func newGithubTestClient(customTransportFactory gitprovider.ChainableRoundTripperFunc) (gitprovider.Client, error) {

	token := os.Getenv("GITHUB_TOKEN")
	if token == "" { // This is the case when the tests run in the ci/cd tool. No need to have a value as everything is cached
		token = " "
	}

	return github.NewClient(
		github.WithOAuth2Token(token),
		github.WithPreChainTransportHook(customTransportFactory),
		github.WithDestructiveAPICalls(true),
	)
}

func newGitlabTestClient(customTransportFactory gitprovider.ChainableRoundTripperFunc) (gitprovider.Client, error) {

	token := os.Getenv("GITLAB_TOKEN")
	if token == "" { // This is the case when the tests run in the ci/cd tool. No need to have a value as everything is cached
		token = " "
	}

	return gitlab.NewClient(
		"",
		"",
		gitlab.WithOAuth2Token(token),
		gitlab.WithPreChainTransportHook(customTransportFactory),
		gitlab.WithDestructiveAPICalls(true),
	)
}

func Test_CreatePullRequestToOrgRepo(t *testing.T) {

	githubOrgName := os.Getenv("GITHUB_ORG_NAME")
	githubUserName := os.Getenv("GITHUB_USER_NAME")

	gitlabOrgName := os.Getenv("GITLAB_ORG_NAME")
	gitlabUserName := os.Getenv("GITLAB_USER_NAME")

	providers := []struct {
		provider string
		client   gitprovider.Client
		domain   string
		orgName  string
		userName string
	}{
		{"github", githubClient, GITHUB_DOMAIN, githubOrgName, githubUserName},
		{"gitlab", gitlabClient, GITLAB_DOMAIN, gitlabOrgName, gitlabUserName},
	}

	testNameFormat := "create pr for %s account [%s]"
	for _, p := range providers {
		testName := fmt.Sprintf(testNameFormat, "org", p.provider)
		t.Run(testName, func(t *testing.T) {
			CreateTestPullRequestToOrgRepo(t, p.client, p.domain, p.orgName)

		})
		testName = fmt.Sprintf(testNameFormat, "user", p.provider)
		t.Run(testName, func(t *testing.T) {
			CreateTestPullRequestToUserRepo(t, p.client, p.domain, p.userName)
		})

	}

}

func CreateTestPullRequestToOrgRepo(t *testing.T, client gitprovider.Client, domain string, orgName string) {

	repoName := "test-org-repo"
	branchName := "test-org-branch"

	doesNotExistOrg := "doesnotexists"

	orgRepoRef := NewOrgRepositoryRef(domain, orgName, repoName)
	doesNotExistOrgRepoRef := NewOrgRepositoryRef(domain, doesNotExistOrg, repoName)
	repoInfo := NewRepositoryInfo("test org repository", gitprovider.RepositoryVisibilityPrivate)
	opts := &gitprovider.RepositoryCreateOptions{
		AutoInit: gitprovider.BoolVar(true),
	}

	err := CreateOrgRepository(client, orgRepoRef, repoInfo, opts)
	assert.NoError(t, err)

	err = CreateOrgRepository(client, doesNotExistOrgRepoRef, repoInfo, opts)
	assert.Error(t, err)

	path := "setup/config.yaml"
	content := "init content"
	files := []gitprovider.CommitFile{
		{
			Path:    &path,
			Content: &content,
		},
	}

	commitMessage := "added config files"
	prTitle := "config files"
	prDescription := "test description"

	err = CreatePullRequestToOrgRepo(client, orgRepoRef, "", branchName, files, commitMessage, prTitle, prDescription)
	assert.NoError(t, err)

	err = CreatePullRequestToOrgRepo(client, orgRepoRef, "branchdoesnotexists", branchName, files, commitMessage, prTitle, prDescription)
	assert.Error(t, err)

	err = CreatePullRequestToOrgRepo(client, doesNotExistOrgRepoRef, "", branchName, files, commitMessage, prTitle, prDescription)
	assert.Error(t, err)

	t.Cleanup(func() {
		ctx := context.Background()
		org, err := client.OrgRepositories().Get(ctx, orgRepoRef)
		assert.NoError(t, err)
		err = org.Delete(ctx)
		assert.NoError(t, err)
	})

}

func CreateTestPullRequestToUserRepo(t *testing.T, client gitprovider.Client, domain string, userAccount string) {

	repoName := "test-user-repo"
	branchName := "test-user-branch"

	doesnotExistUserAccount := "doesnotexists"

	userRepoRef := NewUserRepositoryRef(domain, userAccount, repoName)
	doesNotExistsUserRepoRef := NewUserRepositoryRef(domain, doesnotExistUserAccount, repoName)
	repoInfo := NewRepositoryInfo("test user repository", gitprovider.RepositoryVisibilityPrivate)
	opts := &gitprovider.RepositoryCreateOptions{
		AutoInit: gitprovider.BoolVar(true),
	}

	err := CreateUserRepository(client, userRepoRef, repoInfo, opts)
	assert.NoError(t, err)

	err = CreateUserRepository(client, doesNotExistsUserRepoRef, repoInfo, opts)
	assert.Error(t, err)

	path := "setup/config.yaml"
	content := "init content"
	files := []gitprovider.CommitFile{
		{
			Path:    &path,
			Content: &content,
		},
	}

	commitMessage := "added config files"
	prTitle := "config files"
	prDescription := "test description"

	err = CreatePullRequestToUserRepo(client, userRepoRef, "", branchName, files, commitMessage, prTitle, prDescription)
	assert.NoError(t, err)

	err = CreatePullRequestToUserRepo(client, userRepoRef, "branchdoesnotexists", branchName, files, commitMessage, prTitle, prDescription)
	assert.Error(t, err)

	err = CreatePullRequestToUserRepo(client, doesNotExistsUserRepoRef, "", branchName, files, commitMessage, prTitle, prDescription)
	assert.Error(t, err)

	t.Cleanup(func() {
		ctx := context.Background()
		user, err := client.UserRepositories().Get(ctx, userRepoRef)
		assert.NoError(t, err)
		err = user.Delete(ctx)
		assert.NoError(t, err)
	})

}
