package gitproviders

import (
	"context"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/vendorfakes/fakegitprovider"
)

var (
	dryRunProvider GitProvider
	ctx            context.Context
	repoURL        RepoURL
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

		repoURL = RepoURL{}
	})

	Describe("RepositoryExists", func() {
		It("returns true", func() {
			res, err := dryRunProvider.RepositoryExists(ctx, repoURL)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeTrue())
		})
	})

	Describe("DeployKeyExists", func() {
		It("returns true", func() {
			res, err := dryRunProvider.DeployKeyExists(ctx, repoURL)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeTrue())
		})
	})

	Describe("GetDefaultBranch", func() {
		It("returns branch placeholder", func() {
			res, err := dryRunProvider.GetDefaultBranch(ctx, repoURL)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal("<default-branch>"))
		})
	})

	Describe("GetRepoVisibility", func() {
		It("returns private", func() {
			res, err := dryRunProvider.GetRepoVisibility(ctx, repoURL)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(Equal(gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityPrivate)))
		})
	})

	Describe("UploadDeployKey", func() {
		It("returns nil", func() {
			Expect(dryRunProvider.UploadDeployKey(ctx, repoURL, []byte{})).To(Succeed())
		})
	})

	Describe("CreatePullRequest", func() {
		It("returns nil", func() {
			res, err := dryRunProvider.CreatePullRequest(ctx, repoURL, PullRequestInfo{})
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeNil())
		})
	})

	Describe("GetCommits", func() {
		It("returns emtpy", func() {
			res, err := dryRunProvider.GetCommits(ctx, repoURL, "", 1, 1)
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
