package gitproviders

import (
	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var dryRunProvider GitProvider

var _ = Describe("DryRun", func() {
	var _ = BeforeEach(func() {
		dryRunProvider, _ = NewDryRun()
	})

	Describe("CreateRepository", func() {
		It("returns nil", func() {
			Expect(dryRunProvider.CreateRepository("", "", false)).To(Succeed())
		})
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

	Describe("GetRepoInfo", func() {
		It("returns placeholder", func() {
			res, err := dryRunProvider.GetRepoInfo(AccountTypeUser, "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(*res.DefaultBranch).To(Equal("<default-branch>"))
		})
	})

	Describe("GetRepoInfoFromUrl", func() {
		It("returns placeholder", func() {
			res, err := dryRunProvider.GetRepoInfoFromUrl("")
			Expect(err).ToNot(HaveOccurred())
			Expect(*res.DefaultBranch).To(Equal("<default-branch>"))
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

	Describe("CreatePullRequestToUserRepo", func() {
		It("returns nil", func() {
			res, err := dryRunProvider.CreatePullRequestToUserRepo(gitprovider.UserRepositoryRef{}, "", "", nil, "", "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeNil())
		})
	})

	Describe("CreatePullRequestToOrgRepo", func() {
		It("returns nil", func() {
			res, err := dryRunProvider.CreatePullRequestToOrgRepo(gitprovider.OrgRepositoryRef{}, "", "", nil, "", "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeNil())
		})
	})

	Describe("GetCommitsFromUserRepo", func() {
		It("returns emtpy", func() {
			res, err := dryRunProvider.GetCommitsFromUserRepo(gitprovider.UserRepositoryRef{}, "", 1, 1)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]gitprovider.Commit{}))
		})
	})

	Describe("GetCommitsFromOrgRepo", func() {
		It("returns empty", func() {
			res, err := dryRunProvider.GetCommitsFromOrgRepo(gitprovider.OrgRepositoryRef{}, "", 1, 1)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal([]gitprovider.Commit{}))
		})
	})

	Describe("GetAccountType", func() {
		It("returns user type", func() {
			res, err := dryRunProvider.GetAccountType("")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(AccountTypeUser))
		})
	})

	Describe("GetProviderDomain", func() {
		It("returns github provider", func() {
			Expect(dryRunProvider.GetProviderDomain()).To(Equal("github.com"))
		})
	})
})
