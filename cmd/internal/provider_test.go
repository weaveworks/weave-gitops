package internal

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
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

func fakeBlockingCLIHandlerSuccess(_ context.Context, _ io.Writer) (string, error) {
	return githubToken, nil
}

func fakeBlockingCLIHandlerError(_ context.Context, _ io.Writer) (string, error) {
	return "", errors.New("blocking cli handler goes ka-boom")
}

func fakeAuthHandlerFuncGoodCLI(_ gitproviders.GitProviderName) (auth.BlockingCLIAuthHandler, error) {
	return fakeBlockingCLIHandlerSuccess, nil
}

func fakeAuthHandlerFuncBadCLI(_ gitproviders.GitProviderName) (auth.BlockingCLIAuthHandler, error) {
	return fakeBlockingCLIHandlerError, nil
}

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

func fakeEnvLookupDoesNotExist(key string) (string, bool) {
	return "", false
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

	Describe("auth flow since token is not in an env variable", func() {
		BeforeEach(func() {
			fakeLogger = &loggerfakes.FakeLogger{
				WarningfStub: func(fmtArg string, restArgs ...interface{}) {},
			}
			repoUrl, _ = gitproviders.NewRepoURL("ssh://git@github.com/weaveworks/weave-gitops.git")
		})

		AfterEach(func() {
			fmtArg, restArgs := fakeLogger.WarningfArgsForCall(0)
			Expect(fmtArg).Should(Equal(envVariableWarning))
			Expect(restArgs).To(HaveLen(1))
			Expect(restArgs[0]).Should(Equal("GITHUB_TOKEN"))
			Expect(fakeLogger.WarningfCallCount()).To(Equal(1))
		})

		It("cannot generate auth flow handler", func() {
			client = NewGitProviderClient(os.Stdout, fakeEnvLookupDoesNotExist, fakeAuthHandlerFuncError, fakeLogger)
			provider, err := client.GetProvider(repoUrl, fakeAccountGetterError)

			Expect(provider).To(BeNil())
			_, expectedErr := fakeAuthHandlerFuncError(repoUrl.Provider())
			Expect(err).To(MatchError(fmt.Errorf("error initializing cli auth handler: %w", expectedErr)))
		})

		It("blocking cli handler returns an error during auth flow", func() {
			client = NewGitProviderClient(os.Stdout, fakeEnvLookupDoesNotExist, fakeAuthHandlerFuncBadCLI, fakeLogger)
			provider, err := client.GetProvider(repoUrl, fakeAccountGetterError)

			Expect(provider).To(BeNil())
			_, expectedErr := fakeBlockingCLIHandlerError(context.Background(), bytes.NewBufferString(""))
			Expect(err).To(MatchError(fmt.Errorf("could not complete auth flow: %w", expectedErr)))
		})

		It("success", func() {
			client = NewGitProviderClient(os.Stdout, fakeEnvLookupDoesNotExist, fakeAuthHandlerFuncGoodCLI, fakeLogger)
			provider, err := client.GetProvider(repoUrl, fakeAccountGetterSuccess)

			Expect(err).To(BeNil())
			expectedProvider, _ := gitproviders.New(gitproviders.Config{Provider: repoUrl.Provider(), Token: githubToken}, repoUrl.Owner(), fakeAccountGetterSuccess)
			Expect(provider).To(Equal(expectedProvider))
		})
	})

})
