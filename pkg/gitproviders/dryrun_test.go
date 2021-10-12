package gitproviders

import (
	"context"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/vendorfakes/fakegitprovider"
)

var (
	dryRunProvider GitProvider
	ctx            context.Context
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

		ctx = context.Background()

		dryRunProvider = &dryrunProvider{
			provider: orgProvider,
		}
	})

	Describe("RepositoryExists", func() {
		It("returns true", func() {
			res, err := dryRunProvider.RepositoryExists(ctx, "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeTrue())
		})
	})

	Describe("DeployKeyExists", func() {
		It("returns true", func() {
			res, err := dryRunProvider.DeployKeyExists(ctx, "", "")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeTrue())
		})
	})

	Describe("GetDefaultBranch", func() {
		It("returns branch placeholder", func() {
			res, err := dryRunProvider.GetDefaultBranch(ctx, "")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("<default-branch>"))
		})
	})

	Describe("GetRepoVisibility", func() {
		It("returns private", func() {
			res, err := dryRunProvider.GetRepoVisibility(ctx, "")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityPrivate)))
		})
	})

	Describe("UploadDeployKey", func() {
		It("returns nil", func() {
			Expect(dryRunProvider.UploadDeployKey(ctx, "", "", []byte{})).To(Succeed())
		})
	})

	Describe("CreatePullRequest", func() {
		It("returns nil", func() {
			res, err := dryRunProvider.CreatePullRequest(ctx, "", "", PullRequestInfo{})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeNil())
		})
	})

	Describe("GetCommits", func() {
		It("returns emtpy", func() {
			res, err := dryRunProvider.GetCommits(ctx, "", "", "", 1, 1)
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
