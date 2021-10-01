package gitproviders

import (
	"errors"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/vendorfakes/fakegitprovider"
)

var _ = Describe("Org Provider", func() {
	var (
		orgProvider               GitProvider
		fakeProviderClient        *fakegitprovider.Client
		fakeOrgRepositoriesClient *fakegitprovider.OrgRepositoriesClient
	)

	var _ = BeforeEach(func() {
		fakeOrgRepositoriesClient = &fakegitprovider.OrgRepositoriesClient{}

		fakeProviderClient = &fakegitprovider.Client{}
		fakeProviderClient.OrgRepositoriesReturns(fakeOrgRepositoriesClient)

		orgProvider = orgGitProvider{
			domain:   "github.com",
			provider: fakeProviderClient,
		}
	})

	Describe("RepositoryExists", func() {
		It("returns false when repo not found", func() {
			fakeOrgRepositoriesClient.GetReturns(nil, gitprovider.ErrNotFound)

			res, err := orgProvider.RepositoryExists("repo-name", "owner")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeFalse())
		})

		It("returns error when can't verify", func() {
			fakeOrgRepositoriesClient.GetReturns(nil, errors.New("random error"))

			res, err := orgProvider.RepositoryExists("repo-name", "owner")
			Expect(err).To(HaveOccurred())
			Expect(res).To(BeFalse())
		})

		It("returns true when repo exists", func() {
			res, err := orgProvider.RepositoryExists("repo-name", "owner")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeTrue())
		})
	})

	Describe("DeployKeyExists", func() {
		It("errors out when repo doest exist", func() {
			fakeOrgRepositoriesClient.GetReturns(nil, gitprovider.ErrNotFound)

			res, err := orgProvider.DeployKeyExists("owner", "repo-name")
			Expect(err.Error()).Should(ContainSubstring("error getting org repo reference for owner"))
			Expect(res).To(BeFalse())
		})
	})
})
