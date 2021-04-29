package git_providers

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func Test_CreatePullRequestToOrgRepo(t *testing.T) {

	randomRepoName := fmt.Sprintf("test-repo-%04d", rand.Intn(10000))
	randomBranchName := fmt.Sprintf("test-branch-%04d", rand.Intn(10000))

	orgName := "weaveworks"
	if testOrgName := os.Getenv("GITHUB_TEST_ORG_NAME"); testOrgName != "" {
		orgName = testOrgName
	}

	ghClient, err := GithubProvider()
	assert.NoError(t, err)
	orgRepoRef := NewOrgRepositoryRef(GITHUB_DOMAIN, orgName, randomRepoName)
	repoInfo := NewRepositoryInfo("test org repository", gitprovider.RepositoryVisibilityPrivate)
	opts := &gitprovider.RepositoryCreateOptions{
		AutoInit: gitprovider.BoolVar(true),
	}

	err = CreateOrgRepository(ghClient, orgRepoRef, repoInfo, opts)
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

	ctx := context.Background()
	org, err := ghClient.OrgRepositories().Get(ctx, orgRepoRef)
	assert.NoError(t, err)
	err = org.Delete(ctx)
	assert.NoError(t, err)
}

func Test_CreatePullRequestToUserRepo(t *testing.T) {

	randomRepoName := fmt.Sprintf("test-repo-%04d", rand.Intn(10000))
	randomBranchName := fmt.Sprintf("test-branch-%04d", rand.Intn(10000))

	userAccount := "bot"
	if user := os.Getenv("GITHUB_TEST_USER"); user != "" {
		userAccount = user
	}

	ghClient, err := GithubProvider()
	assert.NoError(t, err)
	userRepoRef := NewUserRepositoryRef(GITHUB_DOMAIN, userAccount, randomRepoName)
	repoInfo := NewRepositoryInfo("test user repository", gitprovider.RepositoryVisibilityPrivate)
	opts := &gitprovider.RepositoryCreateOptions{
		AutoInit: gitprovider.BoolVar(true),
	}

	err = CreateUserRepository(ghClient, userRepoRef, repoInfo, opts)
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

	ctx := context.Background()
	org, err := ghClient.UserRepositories().Get(ctx, userRepoRef)
	assert.NoError(t, err)
	err = org.Delete(ctx)
	assert.NoError(t, err)
}
