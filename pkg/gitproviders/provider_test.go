package gitproviders

import (
	"bytes"
	"context"
	"fmt"
	"github.com/fluxcd/go-git-providers/gitlab"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/utils"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
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
	_ = (*r).Close()
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

	var (
		resp         *http.Response
		err          error
		responseBody string
		requestBody  string
	)

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
			strings.Contains(responseBody, ProjectStillBeingDeleted) {
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

func NewRecorder(cassetteID string, accounts *accounts) (*recorder.Recorder, error) {
	r, err := recorder.New(fmt.Sprintf("./cache/%s", cassetteID))
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
		if accounts.GithubOrgName != GithubOrgTestName ||
			accounts.GithubUserName != GithubUserTestName ||
			accounts.GitlabOrgName != GitlabOrgTestName ||
			accounts.GitlabUserName != GitlabUserTestName {
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

func getTestClientWithCassette(cassetteID string) (gitprovider.Client, *recorder.Recorder, error) {
	t := customTransport{}

	var err error

	cacheGithubRecorder, err := NewRecorder(cassetteID, getAccounts())
	if err != nil {
		return nil, nil, err
	}

	cacheGithubRecorder.SetTransport(&t)

	githubTestClient, err := newGithubTestClient(SetRecorder(cacheGithubRecorder))
	if err != nil {
		return nil, nil, err
	}

	return githubTestClient, cacheGithubRecorder, nil
}

func newGithubTestClient(customTransportFactory gitprovider.ChainableRoundTripperFunc) (gitprovider.Client, error) {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" { // This is the case when the tests run in the ci/cd tool. No need to have a value as everything is cached
		token = " "
	}

	return github.NewClient(
		gitprovider.WithOAuth2Token(token),
		gitprovider.WithPreChainTransportHook(customTransportFactory),
		gitprovider.WithDestructiveAPICalls(true),
	)
}

var gitProvider defaultGitProvider

var _ = Describe("Initialization", func() {
	It("it validates token presence", func() {
		_, err := New(Config{})
		Expect(err).Should(MatchError("failed to build git provider: no git provider token present"))
	})

	It("builds a github client", func() {
		client, err := New(Config{Token: "bla", Provider: GitProviderGitHub})
		Expect(err).ToNot(HaveOccurred())
		Expect(client.(defaultGitProvider).domain).To(Equal(github.DefaultDomain))
	})

	It("builds a gitlab client", func() {
		client, err := New(Config{Token: "bla", Provider: GitProviderGitLab})
		Expect(err).ToNot(HaveOccurred())
		Expect(client.(defaultGitProvider).domain).To(Equal(gitlab.DefaultDomain))
	})
})

var _ = Describe("pull requests", func() {
	accounts := getAccounts()

	var client gitprovider.Client
	var recorder *recorder.Recorder
	var err error

	type tier struct {
		client   gitprovider.Client
		domain   string
		orgName  string
		userName string
	}

	var providers []tier

	var _ = BeforeEach(func() {
		client, recorder, err = getTestClientWithCassette("pull_requests")
		Expect(err).NotTo(HaveOccurred())
		gitProvider = defaultGitProvider{
			domain:   github.DefaultDomain,
			provider: client,
		}

		providers = []tier{
			{client, gitProvider.domain, accounts.GithubOrgName, accounts.GithubUserName},
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

var _ = Describe("commits", func() {
	accounts := getAccounts()

	var client gitprovider.Client
	var recorder *recorder.Recorder
	var err error

	type tier struct {
		client   gitprovider.Client
		domain   string
		orgName  string
		userName string
	}

	var providers []tier

	var _ = BeforeEach(func() {
		client, recorder, err = getTestClientWithCassette("commits")
		Expect(err).NotTo(HaveOccurred())
		gitProvider = defaultGitProvider{
			domain:   github.DefaultDomain,
			provider: client,
		}

		providers = []tier{
			{client, gitProvider.domain, accounts.GithubOrgName, accounts.GithubUserName},
			// Remove this for now as we dont support it yet.
			// {"gitlab", gitlabTestClient, gitlab.DefaultDomain, accounts.GitlabOrgName, accounts.GitlabUserName},
		}

	})

	It("should get commits for user and org accounts in github", func() {
		for _, p := range providers {
			GetCommitToUserRepo(p.client, p.domain, p.userName)
			GetCommitToOrgRepo(p.client, p.domain, p.orgName)
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
		client, recorder, err = getTestClientWithCassette("repo_org_exists")
		Expect(err).NotTo(HaveOccurred())
		gitProvider = defaultGitProvider{
			domain:   github.DefaultDomain,
			provider: client,
		}
	})

	It("succeed on validating repo existence", func() {
		err = gitProvider.CreateRepository(repoName, accounts.GithubOrgName, true)
		Expect(err).NotTo(HaveOccurred())

		exists, err := gitProvider.RepositoryExists(repoName, accounts.GithubOrgName)
		Expect(err).NotTo(HaveOccurred())
		Expect(true).To(Equal(exists))
	})

	AfterEach(func() {
		ctx := context.Background()
		orgRepoRef := NewOrgRepositoryRef(gitProvider.domain, accounts.GithubOrgName, repoName)
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
		client, recorder, err = getTestClientWithCassette("repo_personal_exists")
		Expect(err).NotTo(HaveOccurred())
		gitProvider = defaultGitProvider{
			domain:   github.DefaultDomain,
			provider: client,
		}
	})

	It("succeed on validating repo existence", func() {
		accounts := getAccounts()

		err := gitProvider.CreateRepository(repoName, accounts.GithubUserName, true)
		Expect(err).NotTo(HaveOccurred())

		exists, err := gitProvider.RepositoryExists(repoName, accounts.GithubUserName)
		Expect(err).NotTo(HaveOccurred())
		Expect(true).To(Equal(exists))
	})

	AfterEach(func() {
		ctx := context.Background()
		userRepoRef := NewUserRepositoryRef(gitProvider.domain, accounts.GithubUserName, repoName)
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

	err := gitProvider.CreateOrgRepository(orgRepoRef, repoInfo, opts)
	Expect(err).NotTo(HaveOccurred())

	err = gitProvider.CreateOrgRepository(doesNotExistOrgRepoRef, repoInfo, opts)
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

	prLink, err := gitProvider.CreatePullRequestToOrgRepo(orgRepoRef, "", branchName, files, commitMessage, prTitle, prDescription)
	Expect(err).ToNot(HaveOccurred())
	Expect("https://github.com/weaveworks/test-org-repo/pull/1", prLink.Get().WebURL)

	_, err = gitProvider.CreatePullRequestToOrgRepo(orgRepoRef, "branchdoesnotexists", branchName, files, commitMessage, prTitle, prDescription)
	Expect(err).To(HaveOccurred())

	_, err = gitProvider.CreatePullRequestToOrgRepo(doesNotExistOrgRepoRef, "", branchName, files, commitMessage, prTitle, prDescription)
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

	err := gitProvider.CreateUserRepository(userRepoRef, repoInfo, opts)
	Expect(err).NotTo(HaveOccurred())

	err = gitProvider.CreateUserRepository(doesNotExistsUserRepoRef, repoInfo, opts)
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

	prLink, err := gitProvider.CreatePullRequestToUserRepo(userRepoRef, "", branchName, files, commitMessage, prTitle, prDescription)
	Expect(err).NotTo(HaveOccurred())
	Expect("https://github.com/bot/test-user-repo/pull/1", prLink.Get().WebURL)

	_, err = gitProvider.CreatePullRequestToUserRepo(userRepoRef, "branchdoesnotexists", branchName, files, commitMessage, prTitle, prDescription)
	Expect(err).To(HaveOccurred())

	_, err = gitProvider.CreatePullRequestToUserRepo(doesNotExistsUserRepoRef, "", branchName, files, commitMessage, prTitle, prDescription)
	Expect(err).To(HaveOccurred())

	ctx := context.Background()
	user, err := client.UserRepositories().Get(ctx, userRepoRef)
	Expect(err).NotTo(HaveOccurred())
	err = user.Delete(ctx)
	Expect(err).NotTo(HaveOccurred())
}

func GetCommitToUserRepo(client gitprovider.Client, domain string, userAccount string) {
	repoName := "test-user-commit"

	userRepoRef := NewUserRepositoryRef(domain, userAccount, repoName)
	repoInfo := NewRepositoryInfo("test user commit", gitprovider.RepositoryVisibilityPrivate)
	opts := &gitprovider.RepositoryCreateOptions{
		AutoInit: gitprovider.BoolVar(true),
	}

	err := gitProvider.CreateUserRepository(userRepoRef, repoInfo, opts)
	Expect(err).NotTo(HaveOccurred())

	commits, err := gitProvider.GetCommitsFromUserRepo(userRepoRef, "main", 10, 0)
	Expect(err).NotTo(HaveOccurred())
	Expect(commits[0].Get().Message).To(Equal("Initial commit"))
	Expect(commits[0].Get().Author).To(Equal("bot"))

	ctx := context.Background()
	user, err := client.UserRepositories().Get(ctx, userRepoRef)
	Expect(err).NotTo(HaveOccurred())
	err = user.Delete(ctx)
	Expect(err).NotTo(HaveOccurred())
}

func GetCommitToOrgRepo(client gitprovider.Client, domain string, orgName string) {
	repoName := "test-org-commit"

	orgRepoRef := NewOrgRepositoryRef(domain, orgName, repoName)
	repoInfo := NewRepositoryInfo("test org commit", gitprovider.RepositoryVisibilityPrivate)
	opts := &gitprovider.RepositoryCreateOptions{
		AutoInit: gitprovider.BoolVar(true),
	}

	err := gitProvider.CreateOrgRepository(orgRepoRef, repoInfo, opts)
	Expect(err).NotTo(HaveOccurred())

	commits, err := gitProvider.GetCommitsFromOrgRepo(orgRepoRef, "main", 10, 0)
	Expect(err).NotTo(HaveOccurred())
	Expect(commits[0].Get().Message).To(Equal("Initial commit"))
	Expect(commits[0].Get().Author).To(Equal("bot"))

	ctx := context.Background()
	user, err := client.OrgRepositories().Get(ctx, orgRepoRef)
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
		client, recorder, err = getTestClientWithCassette("get_user_repo_info")
		Expect(err).NotTo(HaveOccurred())
		gitProvider = defaultGitProvider{
			domain:   github.DefaultDomain,
			provider: client,
		}

		userRepoRef = NewUserRepositoryRef(gitProvider.domain, accounts.GithubUserName, repoName)

		repoInfo := NewRepositoryInfo("test user repository", gitprovider.RepositoryVisibilityPrivate)
		opts := &gitprovider.RepositoryCreateOptions{
			AutoInit: gitprovider.BoolVar(true),
		}
		err := gitProvider.CreateUserRepository(userRepoRef, repoInfo, opts)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("Succeed on getting user repo info", func() {
		_, err = gitProvider.GetRepoInfo(AccountTypeUser, accounts.GithubUserName, repoName)
		Expect(err).ShouldNot(HaveOccurred())

		_, err = gitProvider.GetRepoInfo(AccountTypeUser, accounts.GithubUserName, "repoNotExisted")
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
		client, recorder, err = getTestClientWithCassette("deploy_key_user")
		Expect(err).NotTo(HaveOccurred())
		gitProvider = defaultGitProvider{
			domain:   github.DefaultDomain,
			provider: client,
		}

		userRepoRef = NewUserRepositoryRef(gitProvider.domain, accounts.GithubUserName, repoName)
		repoInfo := NewRepositoryInfo("test user repository", gitprovider.RepositoryVisibilityPrivate)
		opts := &gitprovider.RepositoryCreateOptions{
			AutoInit: gitprovider.BoolVar(true),
		}

		err = gitProvider.CreateUserRepository(userRepoRef, repoInfo, opts)
		Expect(err).ShouldNot(HaveOccurred())
		err = utils.WaitUntil(os.Stdout, time.Second, time.Second*30, func() error {
			_, err := gitProvider.GetUserRepo(accounts.GithubUserName, repoName)
			return err
		})
		Expect(err).ShouldNot(HaveOccurred())

		deployKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDBmym4XOiTj4rY3AcJKoJ8QupfgpFWtgNzDxzL0TrzfnurUQm+snozKLHGtOtS7PjMQsMaW9phyhhXv2KxadVI1uweFkC1TK4rPNWrqYX2g0JLXEScvaafSiv+SqozWLN/zhQ0e0jrtrYphtkd+H72RYsdq3mngY4WPJXM7z+HSjHSKilxj7XsxENt0dxT08LArxDC4OQXv9EYFgCyZ7SuLPBgA9160Co46Jm27enB/oBPx5zWd1MlkI+RtUi+XV2pLMzIpvYi2r2iWwOfDqE0N2cfpD0bY7cIOlv0iS7v6Qkmf7pBD+tRGTIZFcD5tGmZl1DOaeCZZ/VAN66aX+rN"
	})

	It("Uploads a new deploy key for a brand new user repo, checks for presence of the key, and shows proper message if trying to re-add it", func() {
		exists, err := gitProvider.DeployKeyExists(accounts.GithubUserName, repoName)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(exists).To(BeFalse())

		stdout := utils.CaptureStdout(func() {
			err = gitProvider.UploadDeployKey(accounts.GithubUserName, repoName, []byte(deployKey))
			Expect(err).ShouldNot(HaveOccurred())
		})
		Expect(stdout).To(Equal("uploading deploy key\n"))

		exists, err = gitProvider.DeployKeyExists(accounts.GithubUserName, repoName)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(exists).To(BeTrue())

		stdout = utils.CaptureStdout(func() {
			err = gitProvider.UploadDeployKey(accounts.GithubUserName, repoName, []byte(deployKey))
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
		client, recorder, err = getTestClientWithCassette("deploy_key_org")
		Expect(err).NotTo(HaveOccurred())
		gitProvider = defaultGitProvider{
			domain:   github.DefaultDomain,
			provider: client,
		}

		orgRepoRef = NewOrgRepositoryRef(gitProvider.domain, accounts.GithubOrgName, repoName)
		repoInfo := NewRepositoryInfo("test user repository", gitprovider.RepositoryVisibilityPrivate)
		opts := &gitprovider.RepositoryCreateOptions{
			AutoInit: gitprovider.BoolVar(true),
		}

		err = gitProvider.CreateOrgRepository(orgRepoRef, repoInfo, opts)
		Expect(err).ShouldNot(HaveOccurred())
		err = utils.WaitUntil(os.Stdout, time.Second, time.Second*30, func() error {
			_, err := gitProvider.GetOrgRepo(accounts.GithubOrgName, repoName)
			return err
		})
		Expect(err).ShouldNot(HaveOccurred())

		deployKey = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDorjCI1Ai7xhZx4e2dYImHbjzbEc0gH1mjnkcb3Tqc5Zs/tQVxo282YIMeXq8IABt2AcwTzDHAviajbPqC05GNRwCmEFrYOnYKhMrdrKtYuCtmEhgnhPQlItXJlF00XwHfYetjfIzFSk8vdLJcwmGp6PPemDW2Xv6CPBAN23OGqTbYYsFuO7+hdU3CgGcR9WPDdzN7/4q1aq4Tk7qhNl5Yxw1DQ0OVgiAQnBJHeViOar14Dw1olhtzL2s88e/TE9t47p9iLXFXwN4irER25A4NUa7DYGpNfUEGQdlf1k81ctegQeA8fOZ4uT4zYSja7mG6QYRgPwN4ZB8ywTcHeON6EzWucSWKM4TcJgASmvJtJn5RifbuzMJTtqpCtIFmpo5/ItQFKYjI18Omqh0ZJe/P9YtYtM+Ac3FIOC0yKU7Ozsx/N7wq3uSIOTv8KCxkEgq2fBi9gF/+kE0BGSVao0RfY/fAUjS/ScuNvo30+MrW+8NmWeWRdhMJkJ25kLGuWBE="
	})

	It("Uploads a new deploy key for a brand new user repo, checks for presence of the key, and shows proper message if trying to re-add it", func() {
		exists, err := gitProvider.DeployKeyExists(accounts.GithubOrgName, repoName)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(exists).To(BeFalse())

		stdout := utils.CaptureStdout(func() {
			err = gitProvider.UploadDeployKey(accounts.GithubOrgName, repoName, []byte(deployKey))
			Expect(err).ShouldNot(HaveOccurred())
		})
		Expect(stdout).To(Equal("uploading deploy key\n"))

		exists, err = gitProvider.DeployKeyExists(accounts.GithubOrgName, repoName)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(exists).To(BeTrue())

		stdout = utils.CaptureStdout(func() {
			err = gitProvider.UploadDeployKey(accounts.GithubOrgName, repoName, []byte(deployKey))
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

var _ = Describe("helpers", func() {
	DescribeTable("DetectGitProviderFromUrl", func(input string, expected GitProviderName) {
		result, err := DetectGitProviderFromUrl(input)
		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(expected))
	},
		Entry("ssh+github", "ssh://git@github.com/weaveworks/weave-gitops.git", GitProviderGitHub),
		Entry("ssh+gitlab", "ssh://git@gitlab.com/weaveworks/weave-gitops.git", GitProviderGitLab),
	)

})

type expectedRepoURL struct {
	s        string
	owner    string
	name     string
	provider GitProviderName
	protocol RepositoryURLProtocol
}

var _ = DescribeTable("NormalizedRepoURL", func(input string, expected expectedRepoURL) {
	result, err := NewNormalizedRepoURL(input)
	Expect(err).NotTo(HaveOccurred())

	Expect(result.String()).To(Equal(expected.s))
	u, err := url.Parse(expected.s)
	Expect(err).NotTo(HaveOccurred())
	Expect(result.URL()).To(Equal(u))
	Expect(result.Owner()).To(Equal(expected.owner))
	Expect(result.Provider()).To(Equal(expected.provider))
	Expect(result.Protocol()).To(Equal(expected.protocol))
},
	Entry("github git clone style", "git@github.com:someuser/podinfo.git", expectedRepoURL{
		s:        "ssh://git@github.com/someuser/podinfo.git",
		owner:    "someuser",
		name:     "podinfo",
		provider: GitProviderGitHub,
		protocol: RepositoryURLProtocolSSH,
	}),
	Entry("github url style", "ssh://git@github.com/someuser/podinfo.git", expectedRepoURL{
		s:        "ssh://git@github.com/someuser/podinfo.git",
		owner:    "someuser",
		name:     "podinfo",
		provider: GitProviderGitHub,
		protocol: RepositoryURLProtocolSSH,
	}),
	Entry("github https", "https://github.com/someuser/podinfo.git", expectedRepoURL{
		s:        "https://github.com/someuser/podinfo.git",
		owner:    "someuser",
		name:     "podinfo",
		provider: GitProviderGitHub,
		protocol: RepositoryURLProtocolHTTPS,
	}),
	Entry("gitlab git clone style", "git@gitlab.com:someuser/podinfo.git", expectedRepoURL{
		s:        "ssh://git@gitlab.com/someuser/podinfo.git",
		owner:    "someuser",
		name:     "podinfo",
		provider: GitProviderGitLab,
		protocol: RepositoryURLProtocolSSH,
	}),
	Entry("gitlab https", "https://gitlab.com/someuser/podinfo.git", expectedRepoURL{
		s:        "https://gitlab.com/someuser/podinfo.git",
		owner:    "someuser",
		name:     "podinfo",
		provider: GitProviderGitLab,
		protocol: RepositoryURLProtocolHTTPS,
	}),
)

var _ = Describe("Test GetRepoVisiblity", func() {
	url := "ssh://git@github.com/foo/bar"
	It("tests that a nil info generates the appropriate error", func() {
		result, underlyingError := getVisibilityFromRepoInfo(url, nil)
		Expect(result).To(BeNil())
		Expect(underlyingError.Error()).To(Equal(fmt.Sprintf("unable to obtain repository visibility for: %s", url)))
	})

	It("tests that a nil visibility reference generates the appropriate error", func() {
		result, underlyingError := getVisibilityFromRepoInfo(url, &gitprovider.RepositoryInfo{Visibility: nil})
		Expect(result).To(BeNil())
		Expect(underlyingError.Error()).To(Equal(fmt.Sprintf("unable to obtain repository visibility for: %s", url)))
	})

	It("tests that a non-nil visibility reference is successful", func() {
		public := gitprovider.RepositoryVisibilityPublic
		result, underlyingError := getVisibilityFromRepoInfo(url, &gitprovider.RepositoryInfo{Visibility: &public})
		Expect(underlyingError).To(BeNil())
		Expect(result).To(Equal(&public))
	})
})
