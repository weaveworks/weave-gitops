package internal

import (
	"errors"
	"fmt"
	"os"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
)

const (
	githubToken = "github-token-123"
	gitlabToken = "gitlab-token-abc"
)

func fakeAuthHandlerFuncError(_ gitproviders.GitProviderName) (auth.BlockingCLIAuthHandler, error) {
	return nil, errors.New("get auth handler goes ka-boom")
}

func fakeAccountGetterError(_ gitprovider.Client, _ string, _ string) (gitproviders.ProviderAccountType, error) {
	return gitproviders.AccountTypeOrg, errors.New("ka-boom")
}

func fakeAccountGetterSuccess(_ gitprovider.Client, _ string, _ string) (gitproviders.ProviderAccountType, error) {
	return gitproviders.AccountTypeUser, nil
}

func fakeEnvLookupExists(key string) (string, bool) {
	if key == "GITHUB_TOKEN" {
		return githubToken, true
	} else if key == "GITLAB_TOKEN" {
		return gitlabToken, true
	} else {
		return "", false
	}
}

var _ = Describe("Get git provider", func() {
	var client gitproviders.Client
	var repoUrl gitproviders.RepoURL
	var fakeLogger *loggerfakes.FakeLogger

	Context("Invalid git provider name", func() {
		It("invalid token key returns an error", func() {
			fakeLogger = &loggerfakes.FakeLogger{}
			client = NewGitProviderClient(os.Stdout, fakeEnvLookupExists, fakeAuthHandlerFuncError, fakeLogger)
			repoUrl, _ = gitproviders.NewRepoURL("ssh://git@some-bucket.com/weaveworks/weave-gitops.git")

			provider, err := client.GetProvider(repoUrl, fakeAccountGetterSuccess)
			Expect(provider).To(BeNil())

			_, expectedErr := getTokenVarName(repoUrl.Provider())
			Expect(err).To(MatchError(fmt.Errorf("could not determine git provider token name: %w", expectedErr)))
		})
	})

	Describe("token exists in env variable", func() {
		Describe("github token", func() {
			BeforeEach(func() {
				fakeLogger = &loggerfakes.FakeLogger{}
				client = NewGitProviderClient(os.Stdout, fakeEnvLookupExists, fakeAuthHandlerFuncError, fakeLogger)
				repoUrl, _ = gitproviders.NewRepoURL("ssh://git@github.com/weaveworks/weave-gitops.git")
			})

			It("gitproviders.New returns an error", func() {
				provider, err := client.GetProvider(repoUrl, fakeAccountGetterError)

				Expect(provider).To(BeNil())
				Expect(err.Error()).To(HavePrefix("error creating git provider client:"))
			})

			It("success", func() {
				provider, err := client.GetProvider(repoUrl, fakeAccountGetterSuccess)

				Expect(err).To(BeNil())
				expectedProvider, _ := gitproviders.New(gitproviders.Config{Provider: repoUrl.Provider(), Token: githubToken}, repoUrl.Owner(), fakeAccountGetterSuccess)
				Expect(provider).To(Equal(expectedProvider))
				Expect(fakeLogger.WarningfCallCount()).To(Equal(0), "we should not write out a warning message to the user if a token is set")
			})
		})

		Describe("gitlab token", func() {
			BeforeEach(func() {
				fakeLogger = &loggerfakes.FakeLogger{}
				client = NewGitProviderClient(os.Stdout, fakeEnvLookupExists, fakeAuthHandlerFuncError, fakeLogger)
				repoUrl, _ = gitproviders.NewRepoURL("ssh://git@gitlab.com/weaveworks/weave-gitops.git")
			})

			It("gitproviders.New returns an error", func() {
				provider, err := client.GetProvider(repoUrl, fakeAccountGetterError)

				Expect(provider).To(BeNil())
				Expect(err.Error()).To(HavePrefix("error creating git provider client:"))
			})

			It("success", func() {
				provider, err := client.GetProvider(repoUrl, fakeAccountGetterSuccess)

				Expect(err).To(BeNil())
				Expect(provider.GetProviderDomain()).To(Equal("gitlab.com"))
			})
		})
	})
})
