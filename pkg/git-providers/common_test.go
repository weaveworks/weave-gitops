package git_providers

import (
	"context"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/dnaeon/go-vcr/recorder"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/stretchr/testify/assert"

	"github.com/fluxcd/go-git-providers/gitprovider"
)

var ghClient gitprovider.Client

func TestMain(m *testing.M) {

	cacheRecorder, err := recorder.New("./cache/git-providers")
	if err != nil {
		panic(err)
	}
	customTransportFactory := func(transport http.RoundTripper) http.RoundTripper {
		cacheRecorder.SetTransport(transport)
		return cacheRecorder
	}

	ghClient, err = GetGHTestClient(customTransportFactory)
	if err != nil {
		panic(err)
	}

	rand.Seed(time.Now().UnixNano())

	m.Run()

	err = cacheRecorder.Stop()
	if err != nil {
		panic(err)
	}

}

func GetGHTestClient(customTransportFactory gitprovider.ChainableRoundTripperFunc) (gitprovider.Client, error) {
	return github.NewClient(
		github.WithOAuth2Token(os.Getenv("GITHUB_TOKEN")),
		github.WithPreChainTransportHook(customTransportFactory),
		github.WithDestructiveAPICalls(true),
	)
}

func Test_CreatePullRequestToOrgRepo(t *testing.T) {

	randomRepoName := "test-org-repo"
	randomBranchName := "test-org-branch"

	orgName := "weaveworks"
	if testOrgName := os.Getenv("GITHUB_TEST_ORG_NAME"); testOrgName != "" {
		orgName = testOrgName
	}

	orgRepoRef := NewOrgRepositoryRef(GITHUB_DOMAIN, orgName, randomRepoName)
	repoInfo := NewRepositoryInfo("test org repository", gitprovider.RepositoryVisibilityPrivate)
	opts := &gitprovider.RepositoryCreateOptions{
		AutoInit: gitprovider.BoolVar(true),
	}

	err := CreateOrgRepository(ghClient, orgRepoRef, repoInfo, opts)
	assert.NoError(t, err)

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

	err = CreatePullRequestToOrgRepo(ghClient, orgRepoRef, "", randomBranchName, files, commitMessage, prTitle, prDescription)
	assert.NoError(t, err)

	t.Cleanup(func() {
		ctx := context.Background()
		org, err := ghClient.OrgRepositories().Get(ctx, orgRepoRef)
		assert.NoError(t, err)
		err = org.Delete(ctx)
		assert.NoError(t, err)
	})
}

func Test_CreatePullRequestToUserRepo(t *testing.T) {

	randomRepoName := "test-user-repo"
	randomBranchName := "test-user-branch"

	userAccount := "bot"
	if user := os.Getenv("GITHUB_TEST_USER"); user != "" {
		userAccount = user
	}

	userRepoRef := NewUserRepositoryRef(GITHUB_DOMAIN, userAccount, randomRepoName)
	repoInfo := NewRepositoryInfo("test user repository", gitprovider.RepositoryVisibilityPrivate)
	opts := &gitprovider.RepositoryCreateOptions{
		AutoInit: gitprovider.BoolVar(true),
	}

	err := CreateUserRepository(ghClient, userRepoRef, repoInfo, opts)
	assert.NoError(t, err)

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

	err = CreatePullRequestToUserRepo(ghClient, userRepoRef, "", randomBranchName, files, commitMessage, prTitle, prDescription)
	assert.NoError(t, err)

	t.Cleanup(func() {
		ctx := context.Background()
		user, err := ghClient.UserRepositories().Get(ctx, userRepoRef)
		assert.NoError(t, err)
		err = user.Delete(ctx)
		assert.NoError(t, err)
	})

}
