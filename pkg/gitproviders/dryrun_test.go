package gitproviders

import (
	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/vendorfakes/fakegitprovider"
)

var (
	dryRunProvider GitProvider
)

var _ = Describe("DryRun", func() {
	var _ = BeforeEach(func() {
		orgProvider := orgGitProvider{
			domain: "github.com",
			provider: &fakegitprovider.Client{
				ProviderIDStub: func() gitprovider.ProviderID {
					return gitprovider.ProviderID(GitProviderGitHub)
				},
			},
		}

		dryRunProvider = &dryrunProvider{
			provider: orgProvider,
		}
	})

	Describe("RepositoryExists", func() {
		It("returns true", func() {
			res, err := dryRunProvider.RepositoryExists("", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeTrue())
		})
	})

	Describe("DeployKeyExists", func() {
		It("returns true", func() {
			res, err := dryRunProvider.DeployKeyExists("", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeTrue())
		})
	})

	Describe("GetDefaultBranch", func() {
		It("returns branch placeholder", func() {
			res, err := dryRunProvider.GetDefaultBranch("")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("<default-branch>"))
		})
	})

	Describe("GetRepoVisibility", func() {
		It("returns private", func() {
			res, err := dryRunProvider.GetRepoVisibility("")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityPrivate)))
		})
	})

	Describe("UploadDeployKey", func() {
		It("returns nil", func() {
			Expect(dryRunProvider.UploadDeployKey("", "", []byte{})).To(Succeed())
		})
	})

	Describe("CreatePullRequest", func() {
		It("returns nil", func() {
			res, err := dryRunProvider.CreatePullRequest("", "", PullRequestInfo{})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeNil())
		})
	})

	Describe("GetCommits", func() {
		It("returns emtpy", func() {
			res, err := dryRunProvider.GetCommits("", "", "", 1, 1)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]gitprovider.Commit{}))
		})
	})

	Describe("GetProviderDomain", func() {
		It("returns github provider", func() {
			Expect(dryRunProvider.GetProviderDomain()).To(Equal("github.com"))
		})
	})
})
