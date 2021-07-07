package gitproviders

import (
	"bytes"
	"context"
	"fmt"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/fluxcd/go-git-providers/gitprovider"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	GithubOrgTestName  = "weaveworks"
	GithubUserTestName = "bot"
	GitlabOrgTestName  = "weaveworks"
	GitlabUserTestName = "bot"
)

type customTransport struct {
	transport http.RoundTripper
	mux       *sync.Mutex
}

func getBodyFromReaderWithoutConsuming(r *io.ReadCloser) string {
	body, _ := ioutil.ReadAll(*r)
	(*r).Close()
	*r = ioutil.NopCloser(bytes.NewBuffer(body))
	return string(body)
}

const (
	ConnectionResetByPeer    = "connection reset by peer"
	ProjectStillBeingDeleted = "The project is still being deleted"
)

func (t *customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t.mux.Lock()
	defer t.mux.Unlock()

	var resp *http.Response
	var err error
	var responseBody string
	var requestBody string
	retryCount := 15
	for retryCount != 0 {
		responseBody = ""
		requestBody = ""
		if req != nil && req.Body != nil {
			requestBody = getBodyFromReaderWithoutConsuming(&req.Body)
		}
		resp, err = t.transport.RoundTrip(req)
		if resp != nil && resp.Body != nil {
			responseBody = getBodyFromReaderWithoutConsuming(&resp.Body)
		}
		if (err != nil && (strings.Contains(err.Error(), ConnectionResetByPeer))) ||
			strings.Contains(string(responseBody), ProjectStillBeingDeleted) {
			time.Sleep(4 * time.Second)
			if req != nil && req.Body != nil {
				req.Body = ioutil.NopCloser(strings.NewReader(requestBody))
			}
			retryCount--
			continue
		}
		break
	}

	return resp, err
}

func SetRecorder(recorder *recorder.Recorder) gitprovider.ChainableRoundTripperFunc {
	return func(transport http.RoundTripper) http.RoundTripper {
		recorder.SetTransport(transport)
		return recorder
	}
}

type accounts struct {
	GithubOrgName  string
	GithubUserName string
	GitlabOrgName  string
	GitlabUserName string
}

func NewRecorder(provider string, accounts *accounts) (*recorder.Recorder, error) {
	r, err := recorder.New(fmt.Sprintf("./cache/%s", provider))
	if err != nil {
		return nil, err
	}

	r.SetMatcher(func(r *http.Request, i cassette.Request) bool {
		if accounts.GithubOrgName != GithubOrgTestName ||
			accounts.GithubUserName != GithubUserTestName ||
			accounts.GitlabOrgName != GitlabOrgTestName ||
			accounts.GitlabUserName != GitlabUserTestName {
			r.URL, _ = url.Parse(strings.Replace(r.URL.String(), accounts.GithubOrgName, GithubOrgTestName, -1))
			r.URL, _ = url.Parse(strings.Replace(r.URL.String(), accounts.GithubUserName, GithubUserTestName, -1))
			r.URL, _ = url.Parse(strings.Replace(r.URL.String(), accounts.GitlabOrgName, GitlabOrgTestName, -1))
			r.URL, _ = url.Parse(strings.Replace(r.URL.String(), accounts.GitlabUserName, GitlabUserTestName, -1))
		}

		return r.Method == i.Method && (r.URL.String() == i.URL)
	})

	r.AddSaveFilter(func(i *cassette.Interaction) error {
		if accounts.GithubOrgName != GithubOrgTestName {
			i.Response.Body = strings.Replace(i.Response.Body, accounts.GithubOrgName, GithubOrgTestName, -1)
			i.Response.Body = strings.Replace(i.Response.Body, accounts.GithubUserName, GithubUserTestName, -1)
			i.Response.Body = strings.Replace(i.Response.Body, accounts.GitlabOrgName, GitlabOrgTestName, -1)
			i.Response.Body = strings.Replace(i.Response.Body, accounts.GitlabUserName, GitlabUserTestName, -1)

			i.Request.URL = strings.Replace(i.Request.URL, accounts.GithubOrgName, GithubOrgTestName, -1)
			i.Request.URL = strings.Replace(i.Request.URL, accounts.GithubUserName, GithubUserTestName, -1)
			i.Request.URL = strings.Replace(i.Request.URL, accounts.GitlabOrgName, GitlabOrgTestName, -1)
			i.Request.URL = strings.Replace(i.Request.URL, accounts.GitlabUserName, GitlabUserTestName, -1)

			i.Response.Body = strings.Replace(i.Response.Body, accounts.GithubOrgName, GithubOrgTestName, -1)
			i.Response.Body = strings.Replace(i.Response.Body, accounts.GithubUserName, GithubUserTestName, -1)
			i.Response.Body = strings.Replace(i.Response.Body, accounts.GitlabOrgName, GitlabOrgTestName, -1)
			i.Response.Body = strings.Replace(i.Response.Body, accounts.GitlabUserName, GitlabUserTestName, -1)

			for headerKey, h := range i.Response.Headers {
				for ind, header := range h {
					header = strings.Replace(header, accounts.GithubOrgName, GithubOrgTestName, -1)
					header = strings.Replace(header, accounts.GithubUserName, GithubUserTestName, -1)
					header = strings.Replace(header, accounts.GitlabOrgName, GitlabOrgTestName, -1)
					header = strings.Replace(header, accounts.GitlabUserName, GitlabUserTestName, -1)
					if ind == 0 {
						i.Response.Headers.Set(headerKey, header)
					} else {
						i.Response.Headers.Add(headerKey, header)
					}
				}
			}
		}
		return nil
	})

	return r, nil
}

func getAccounts() *accounts {
	accounts := &accounts{}

	ghOrgName := os.Getenv("GITHUB_ORG_NAME")
	if ghOrgName == "" {
		accounts.GithubOrgName = GithubOrgTestName
	} else {
		accounts.GithubOrgName = ghOrgName
	}

	ghUserName := os.Getenv("GITHUB_USER_NAME")
	if ghUserName == "" {
		accounts.GithubUserName = GithubUserTestName
	} else {
		accounts.GithubUserName = ghUserName
	}

	glOrgName := os.Getenv("GITLAB_ORG_NAME")
	if glOrgName == "" {
		accounts.GitlabOrgName = GitlabOrgTestName
	} else {
		accounts.GitlabOrgName = glOrgName
	}

	glUserName := os.Getenv("GITLAB_USER_NAME")
	if glUserName == "" {
		accounts.GitlabUserName = GitlabUserTestName
	} else {
		accounts.GitlabUserName = glUserName
	}

	return accounts
}

func getTestClientWithCassette(cassetteID string) (gitprovider.Client,*recorder.Recorder,error){
	t := customTransport{}

	var err error
	cacheGithubRecorder, err := NewRecorder(cassetteID, getAccounts())
	if err != nil {
		return nil,nil,err
	}

	cacheGithubRecorder.SetTransport(&t)

	githubTestClient, err := newGithubTestClient(SetRecorder(cacheGithubRecorder))
	if err != nil {
		return nil,nil,err
	}

	return githubTestClient,cacheGithubRecorder,nil
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

var _ = Describe("pull requests", func() {
	accounts := getAccounts()

	var client gitprovider.Client
	var recorder *recorder.Recorder
	var err error

	type tier struct {
		provider string
		client   gitprovider.Client
		domain   string
		orgName  string
		userName string
	}

	var providers []tier

	BeforeEach(func() {
		client,recorder,err = getTestClientWithCassette("pull_requests")
		Expect(err).NotTo(HaveOccurred())
		SetGithubProvider(client)

		providers = []tier{
			{"github", client, github.DefaultDomain, accounts.GithubOrgName, accounts.GithubUserName},
			// Remove this for now as we dont support it yet.
			// {"gitlab", gitlabTestClient, gitlab.DefaultDomain, accounts.GitlabOrgName, accounts.GitlabUserName},
		}

	})

	It("should create pr user and org accounts in github", func() {
		for _, p := range providers {
			CreateTestPullRequestToOrgRepo(p.client, p.domain, p.orgName)
			CreateTestPullRequestToUserRepo(p.client, p.domain, p.userName)
		}
	})

	AfterEach(func() {
		err = recorder.Stop()
		Expect(err).ToNot(HaveOccurred())
	})

})

var _ = Describe("test org repo exists", func() {

	accounts := getAccounts()

	var client gitprovider.Client
	var recorder *recorder.Recorder
	var err error
	repoName := "repo-exists-org"
	BeforeEach(func() {
		client,recorder,err = getTestClientWithCassette("repo_org_exists")
		Expect(err).NotTo(HaveOccurred())
		SetGithubProvider(client)
	})

	It("succeed on validating repo existence", func() {
		err = CreateRepository(repoName, accounts.GithubOrgName, true)
		Expect(err).NotTo(HaveOccurred())

		exists, err := RepositoryExists(repoName, accounts.GithubOrgName)
		Expect(err).NotTo(HaveOccurred())
		Expect(true).To(Equal(exists))
	})

	AfterEach(func() {
		ctx := context.Background()
		orgRepoRef := NewOrgRepositoryRef(github.DefaultDomain, accounts.GithubOrgName, repoName)
		org, err := client.OrgRepositories().Get(ctx, orgRepoRef)
		Expect(err).NotTo(HaveOccurred())
		err = org.Delete(ctx)
		Expect(err).NotTo(HaveOccurred())
		err = recorder.Stop()
		Expect(err).NotTo(HaveOccurred())
	})

})

var _ = Describe("test personal repo exists", func() {

	accounts := getAccounts()

	var client gitprovider.Client
	var recorder *recorder.Recorder
	var err error
	repoName := "repo-exists-personal"
	BeforeEach(func() {
		client,recorder,err = getTestClientWithCassette("repo_personal_exists")
		Expect(err).NotTo(HaveOccurred())
		SetGithubProvider(client)
	})

	It("succeed on validating repo existence", func() {

		accounts := getAccounts()

		err := CreateRepository(repoName, accounts.GithubUserName, true)
		Expect(err).NotTo(HaveOccurred())

		exists, err := RepositoryExists(repoName, accounts.GithubUserName)
		Expect(err).NotTo(HaveOccurred())
		Expect(true).To(Equal(exists))
	})

	AfterEach(func() {
		ctx := context.Background()
		userRepoRef := NewUserRepositoryRef(github.DefaultDomain, accounts.GithubUserName, repoName)
		user, err := client.UserRepositories().Get(ctx, userRepoRef)
		Expect(err).NotTo(HaveOccurred())
		err = user.Delete(ctx)
		Expect(err).NotTo(HaveOccurred())
		err = recorder.Stop()
		Expect(err).NotTo(HaveOccurred())
	})

})

func CreateTestPullRequestToOrgRepo(client gitprovider.Client, domain string, orgName string) {
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
	Expect(err).NotTo(HaveOccurred())

	err = CreateOrgRepository(client, doesNotExistOrgRepoRef, repoInfo, opts)
	Expect(err).To(HaveOccurred())

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

	err = CreatePullRequestToOrgRepo(orgRepoRef, "", branchName, files, commitMessage, prTitle, prDescription)
	Expect(err).ToNot(HaveOccurred())

	err = CreatePullRequestToOrgRepo(orgRepoRef, "branchdoesnotexists", branchName, files, commitMessage, prTitle, prDescription)
	Expect(err).To(HaveOccurred())

	err = CreatePullRequestToOrgRepo(doesNotExistOrgRepoRef, "", branchName, files, commitMessage, prTitle, prDescription)
	Expect(err).To(HaveOccurred())

	ctx := context.Background()
	org, err := client.OrgRepositories().Get(ctx, orgRepoRef)
	Expect(err).ToNot(HaveOccurred())
	err = org.Delete(ctx)
	Expect(err).ToNot(HaveOccurred())
}

func CreateTestPullRequestToUserRepo(client gitprovider.Client, domain string, userAccount string) {
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
	Expect(err).NotTo(HaveOccurred())

	err = CreateUserRepository(client, doesNotExistsUserRepoRef, repoInfo, opts)
	Expect(err).To(HaveOccurred())

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

	err = CreatePullRequestToUserRepo(userRepoRef, "", branchName, files, commitMessage, prTitle, prDescription)
	Expect(err).NotTo(HaveOccurred())

	err = CreatePullRequestToUserRepo(userRepoRef, "branchdoesnotexists", branchName, files, commitMessage, prTitle, prDescription)
	Expect(err).To(HaveOccurred())

	err = CreatePullRequestToUserRepo(doesNotExistsUserRepoRef, "", branchName, files, commitMessage, prTitle, prDescription)
	Expect(err).To(HaveOccurred())

	ctx := context.Background()
	user, err := client.UserRepositories().Get(ctx, userRepoRef)
	Expect(err).NotTo(HaveOccurred())
	err = user.Delete(ctx)
	Expect(err).NotTo(HaveOccurred())

}

var _ = Describe("Get User repo info", func() {

	accounts := getAccounts()
	var client gitprovider.Client
	var recorder *recorder.Recorder
	var err error
	var userRepoRef gitprovider.UserRepositoryRef
	repoName := "test-user-repo-info"

	BeforeEach(func() {
		client,recorder,err = getTestClientWithCassette("get_repo_info")
		Expect(err).NotTo(HaveOccurred())
		SetGithubProvider(client)

		userRepoRef = NewUserRepositoryRef(github.DefaultDomain, accounts.GithubUserName, repoName)

		repoInfo := NewRepositoryInfo("test user repository", gitprovider.RepositoryVisibilityPrivate)
		opts := &gitprovider.RepositoryCreateOptions{
			AutoInit: gitprovider.BoolVar(true),
		}
		err := CreateUserRepository(client, userRepoRef, repoInfo, opts)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("Succeed on getting user repo info", func() {

		err = GetRepoInfo(client, AccountTypeUser, accounts.GithubUserName, repoName)
		Expect(err).ShouldNot(HaveOccurred())

		err = GetRepoInfo(client, AccountTypeUser, accounts.GithubUserName, "repoNotExisted")
		Expect(err).Should(HaveOccurred())

	})

	AfterEach(func() {
		ctx := context.Background()
		user, err := client.UserRepositories().Get(ctx, userRepoRef)
		Expect(err).ShouldNot(HaveOccurred())
		err = user.Delete(ctx)
		Expect(err).ShouldNot(HaveOccurred())
		err = recorder.Stop()
		Expect(err).ShouldNot(HaveOccurred())
	})
})

var _ = Describe("Test user deploy keys creation", func() {

	accounts := getAccounts()
	var client gitprovider.Client
	var recorder *recorder.Recorder
	var err error
	var userRepoRef gitprovider.UserRepositoryRef
	repoName := "test-deploy-key-user-repo"
	var deployKey string

	BeforeEach(func() {
		client,recorder,err = getTestClientWithCassette("deploy_key_user")
		Expect(err).NotTo(HaveOccurred())
		SetGithubProvider(client)

		userRepoRef = NewUserRepositoryRef(github.DefaultDomain, accounts.GithubUserName, repoName)
		repoInfo := NewRepositoryInfo("test user repository", gitprovider.RepositoryVisibilityPrivate)
		opts := &gitprovider.RepositoryCreateOptions{
			AutoInit: gitprovider.BoolVar(true),
		}

		err = CreateUserRepository(client, userRepoRef, repoInfo, opts)
		Expect(err).ShouldNot(HaveOccurred())
		err = utils.WaitUntil(os.Stdout, time.Second, time.Second*30, func() error {
			return GetUserRepo(client, accounts.GithubUserName, repoName)
		})
		Expect(err).ShouldNot(HaveOccurred())

		deployKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDBmym4XOiTj4rY3AcJKoJ8QupfgpFWtgNzDxzL0TrzfnurUQm+snozKLHGtOtS7PjMQsMaW9phyhhXv2KxadVI1uweFkC1TK4rPNWrqYX2g0JLXEScvaafSiv+SqozWLN/zhQ0e0jrtrYphtkd+H72RYsdq3mngY4WPJXM7z+HSjHSKilxj7XsxENt0dxT08LArxDC4OQXv9EYFgCyZ7SuLPBgA9160Co46Jm27enB/oBPx5zWd1MlkI+RtUi+XV2pLMzIpvYi2r2iWwOfDqE0N2cfpD0bY7cIOlv0iS7v6Qkmf7pBD+tRGTIZFcD5tGmZl1DOaeCZZ/VAN66aX+rN"
	})


	It("Uploads a new deploy key for a brand new user repo, checks for presence of the key, and shows proper message if trying to re-add it", func() {

		exists, err := DeployKeyExists(accounts.GithubUserName, repoName)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(exists).To(BeFalse())

		stdout := utils.CaptureStdout(func() {
			err = UploadDeployKey(accounts.GithubUserName, repoName, []byte(deployKey))
			Expect(err).ShouldNot(HaveOccurred())
		})
		Expect(stdout).To(Equal("uploading deploy key\n"))

		exists, err = DeployKeyExists(accounts.GithubUserName, repoName)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(exists).To(BeTrue())

		stdout = utils.CaptureStdout(func() {
			err = UploadDeployKey(accounts.GithubUserName, repoName, []byte(deployKey))
			Expect(err).Should(HaveOccurred())
		})
		Expect(stdout).To(Equal("uploading deploy key\n"))

	})

	AfterEach(func() {
		ctx := context.Background()
		user, err := client.UserRepositories().Get(ctx, userRepoRef)
		Expect(err).ShouldNot(HaveOccurred())
		err = user.Delete(ctx)
		Expect(err).ShouldNot(HaveOccurred())
		err = recorder.Stop()
		Expect(err).ShouldNot(HaveOccurred())
	})

})

var _ = Describe("Test org deploy keys creation", func() {

	accounts := getAccounts()
	var client gitprovider.Client
	var recorder *recorder.Recorder
	var err error
	var orgRepoRef gitprovider.OrgRepositoryRef
	repoName := "test-deploy-key-org-repo"
	var deployKey string

	BeforeEach(func() {
		client,recorder,err = getTestClientWithCassette("deploy_key_org")
		Expect(err).NotTo(HaveOccurred())
		SetGithubProvider(client)

		orgRepoRef = NewOrgRepositoryRef(github.DefaultDomain, accounts.GithubOrgName, repoName)
		repoInfo := NewRepositoryInfo("test user repository", gitprovider.RepositoryVisibilityPrivate)
		opts := &gitprovider.RepositoryCreateOptions{
			AutoInit: gitprovider.BoolVar(true),
		}

		err = CreateOrgRepository(client, orgRepoRef, repoInfo, opts)
		Expect(err).ShouldNot(HaveOccurred())
		err = utils.WaitUntil(os.Stdout, time.Second, time.Second*30, func() error {
			return GetOrgRepo(client, accounts.GithubOrgName, repoName)
		})
		Expect(err).ShouldNot(HaveOccurred())

		deployKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDorjCI1Ai7xhZx4e2dYImHbjzbEc0gH1mjnkcb3Tqc5Zs/tQVxo282YIMeXq8IABt2AcwTzDHAviajbPqC05GNRwCmEFrYOnYKhMrdrKtYuCtmEhgnhPQlItXJlF00XwHfYetjfIzFSk8vdLJcwmGp6PPemDW2Xv6CPBAN23OGqTbYYsFuO7+hdU3CgGcR9WPDdzN7/4q1aq4Tk7qhNl5Yxw1DQ0OVgiAQnBJHeViOar14Dw1olhtzL2s88e/TE9t47p9iLXFXwN4irER25A4NUa7DYGpNfUEGQdlf1k81ctegQeA8fOZ4uT4zYSja7mG6QYRgPwN4ZB8ywTcHeON6EzWucSWKM4TcJgASmvJtJn5RifbuzMJTtqpCtIFmpo5/ItQFKYjI18Omqh0ZJe/P9YtYtM+Ac3FIOC0yKU7Ozsx/N7wq3uSIOTv8KCxkEgq2fBi9gF/+kE0BGSVao0RfY/fAUjS/ScuNvo30+MrW+8NmWeWRdhMJkJ25kLGuWBE="
	})


	It("Uploads a new deploy key for a brand new user repo, checks for presence of the key, and shows proper message if trying to re-add it", func() {

		exists, err := DeployKeyExists(accounts.GithubOrgName, repoName)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(exists).To(BeFalse())

		stdout := utils.CaptureStdout(func() {
			err = UploadDeployKey(accounts.GithubOrgName, repoName, []byte(deployKey))
			Expect(err).ShouldNot(HaveOccurred())
		})
		Expect(stdout).To(Equal("uploading deploy key\n"))
		time.Sleep(time.Second*10)

		exists, err = DeployKeyExists(accounts.GithubOrgName, repoName)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(exists).To(BeTrue())

		stdout = utils.CaptureStdout(func() {
			err = UploadDeployKey(accounts.GithubOrgName, repoName, []byte(deployKey))
			Expect(err).Should(HaveOccurred())
		})
		Expect(stdout).To(Equal("uploading deploy key\n"))

	})

	AfterEach(func() {
		ctx := context.Background()
		org, err := client.OrgRepositories().Get(ctx, orgRepoRef)
		Expect(err).ShouldNot(HaveOccurred())
		err = org.Delete(ctx)
		Expect(err).ShouldNot(HaveOccurred())
		err = recorder.Stop()
		Expect(err).ShouldNot(HaveOccurred())
	})
})