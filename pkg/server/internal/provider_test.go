package internal

import (
	"errors"
	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const fakeToken = "keep-it-secret:keep-it-safe"

func fakeAccountGetterError(_ gitprovider.Client, _ string, _ string) (gitproviders.ProviderAccountType, error) {
	return gitproviders.AccountTypeUser, errors.New("ka-boom")
}

func fakeAccountGetterSuccess(_ gitprovider.Client, _ string, _ string) (gitproviders.ProviderAccountType, error) {
	return gitproviders.AccountTypeUser, nil
}

var _ = Describe("Get git provider", func() {
	var client gitproviders.Client
	var repoUrl gitproviders.RepoURL
	BeforeEach(func() {
		client = NewGitProviderClient(fakeToken)
		repoUrl, _ = gitproviders.NewRepoURL("ssh://git@github.com/weaveworks/weave-gitops.git")
	})

	It("gitproviders.New throws an error", func() {
		provider, err := client.GetProvider(repoUrl, fakeAccountGetterError)

		Expect(provider).To(BeNil())
		Expect(err.Error()).To(HavePrefix("error creating git provider client:"))
	})

	It("success", func() {
		provider, err := client.GetProvider(repoUrl, fakeAccountGetterSuccess)

		Expect(err).To(BeNil())
		expectedProvider, _ := gitproviders.New(gitproviders.Config{Provider: repoUrl.Provider(), Token: fakeToken}, repoUrl.Owner(), fakeAccountGetterSuccess)
		Expect(provider).To(Equal(expectedProvider))
	})
})
