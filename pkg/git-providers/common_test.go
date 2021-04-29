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
	"github.com/thanhpk/randstr"
)

func init(){
	rand.Seed(time.Now().UnixNano())
}

func CreatePullRequestFromScratch(provider gitprovider.Client, userRepRef gitprovider.UserRepositoryRef) {

	ctx := context.Background()

	ur, err := provider.UserRepositories().Get(ctx, userRepRef)
	if err != nil {
		panic(err)
	}

	defaultBranch := ur.Get().DefaultBranch

	commits, err := ur.Commits().ListPage(ctx, *defaultBranch, 1, 0)
	if err != nil {
		panic(err)
	}

	latestCommit := commits[0]

	branchName := "test-branch-" + randstr.Hex(3)

	if err := ur.Branches().Create(ctx, branchName, latestCommit.Get().Sha); err != nil {
		panic(err)
	}

	path := "setup/config.txt"
	content := "yaml content"
	files := []gitprovider.CommitFile{
		gitprovider.CommitFile{
			Path:    &path,
			Content: &content,
		},
	}

	if _, err := ur.Commits().Create(ctx, branchName, "added config files", files); err != nil {
		panic(err)
	}

	if err := ur.PullRequests().Create(ctx, "Added wego config files", branchName, *defaultBranch, "added config file"); err != nil {
		panic(err)
	}

}

func Test_CreateOrgRepository(t *testing.T) {

	randomRepoName := fmt.Sprintf("test-repo-%04d", rand.Intn(10000))

	orgName := "weaveworks"
	if testOrgName := os.Getenv("GITHUB_TEST_ORG_NAME"); testOrgName != ""{
		orgName = testOrgName
	}

	ghClient,err := GithubProvider()
	assert.NoError(t, err)
	orgRepoRef := NewOrgRepositoryRef(GITHUB_DOMAIN,orgName,randomRepoName)
	repoInfo := NewRepositoryInfo("test org repository",gitprovider.RepositoryVisibilityPrivate)

	err = CreateOrgRepository(ghClient,orgRepoRef,repoInfo)
	assert.NoError(t, err)

}

func Test_CreateUserRepository(t *testing.T) {

	randomRepoName := fmt.Sprintf("test-repo-%04d", rand.Intn(10000))

	orgName := "weaveworks"
	if testOrgName := os.Getenv("GITHUB_TEST_ORG_NAME"); testOrgName != ""{
		orgName = testOrgName
	}

	ghClient,err := GithubProvider()
	assert.NoError(t, err)
	orgRepoRef := NewUserRepositoryRef(GITHUB_DOMAIN,orgName,randomRepoName)
	repoInfo := NewRepositoryInfo("test user repository",gitprovider.RepositoryVisibilityPrivate)

	err = CreateUserRepository(ghClient,orgRepoRef,repoInfo)
	assert.NoError(t, err)

}

func Test_CreatePullRequestToOrgRepo(t *testing.T) {

	randomRepoName := fmt.Sprintf("test-repo-%04d", rand.Intn(10000))
	randomBranchName := fmt.Sprintf("test-branch-%04d", rand.Intn(10000))

	orgName := os.Getenv("GITHUB_TEST_USER_NAME")

	ghClient,err := GithubProvider()
	assert.NoError(t, err)
	orgRepoRef := NewOrgRepositoryRef(GITHUB_DOMAIN,orgName,randomRepoName)

	path := "setup/config.yaml"
	content := "init content"
	files := []gitprovider.CommitFile{
		{
			Path: &path,
			Content: &content,
		},
	}

	commitMessage := "added config files"
	prTitle := "config files"
	prDescription := "test description"

	err = CreatePullRequestToOrgRepo(ghClient,orgRepoRef,"",randomBranchName,files,commitMessage,prTitle,prDescription)
	assert.NoError(t, err)

}

func Test_CreatePullRequestToUserRepo(t *testing.T) {

	randomRepoName := fmt.Sprintf("test-repo-%04d", rand.Intn(10000))
	randomBranchName := fmt.Sprintf("test-branch-%04d", rand.Intn(10000))

	userName := os.Getenv("GITHUB_TEST_USER_NAME")

	ghClient,err := GithubProvider()
	assert.NoError(t, err)
	userRepoRef := NewUserRepositoryRef(GITHUB_DOMAIN,userName,randomRepoName)

	path := "setup/config.yaml"
	content := "init content"
	files := []gitprovider.CommitFile{
		{
			Path: &path,
			Content: &content,
		},
	}

	commitMessage := "added config files"
	prTitle := "config files"
	prDescription := "test description"

	err = CreatePullRequestToUserRepo(ghClient,userRepoRef,"",randomBranchName,files,commitMessage,prTitle,prDescription)
	assert.NoError(t, err)

}