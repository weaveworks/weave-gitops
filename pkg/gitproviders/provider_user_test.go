package gitproviders

import (
	"errors"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/vendorfakes/fakegitprovider"
)

var _ = Describe("User Provider", func() {
	var (
		userProvider GitProvider

		gitProviderClient *fakegitprovider.Client
		userRepoClient    *fakegitprovider.UserRepositoriesClient
		userRepo          *fakegitprovider.UserRepository

		commitClient       *fakegitprovider.CommitClient
		branchesClient     *fakegitprovider.BranchClient
		pullRequestsClient *fakegitprovider.PullRequestClient
	)

	var _ = BeforeEach(func() {
		commitClient = &fakegitprovider.CommitClient{}
		branchesClient = &fakegitprovider.BranchClient{}
		pullRequestsClient = &fakegitprovider.PullRequestClient{}

		userRepo = &fakegitprovider.UserRepository{}
		userRepo.CommitsReturns(commitClient)
		userRepo.BranchesReturns(branchesClient)
		userRepo.PullRequestsReturns(pullRequestsClient)

		userRepoClient = &fakegitprovider.UserRepositoriesClient{}
		userRepoClient.GetReturns(userRepo, nil)

		gitProviderClient = &fakegitprovider.Client{}
		gitProviderClient.UserRepositoriesReturns(userRepoClient)

		userProvider = userGitProvider{
			domain:   "github.com",
			provider: gitProviderClient,
		}
	})

	Describe("RepositoryExists", func() {
		It("returns false when repo not found", func() {
			userRepoClient.GetReturns(nil, gitprovider.ErrNotFound)

			res, err := userProvider.RepositoryExists("repo-name", "owner")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeFalse())
		})

		It("returns error when can't verify", func() {
			userRepoClient.GetReturns(nil, errors.New("random error"))

			res, err := userProvider.RepositoryExists("repo-name", "owner")
			Expect(err).To(HaveOccurred())
			Expect(res).To(BeFalse())
		})

		It("returns true when repo exists", func() {
			res, err := userProvider.RepositoryExists("repo-name", "owner")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeTrue())
		})
	})

	Describe("DeployKeyExists", func() {
		var deployKeyClient *fakegitprovider.DeployKeyClient

		BeforeEach(func() {
			deployKeyClient = &fakegitprovider.DeployKeyClient{}
			userRepo.DeployKeysReturns(deployKeyClient)
		})

		It("return error when repo doest exist", func() {
			userRepoClient.GetReturns(nil, gitprovider.ErrNotFound)

			res, err := userProvider.DeployKeyExists("owner", "repo-name")
			Expect(err.Error()).Should(ContainSubstring("error getting user repo reference for owner"))
			Expect(res).To(BeFalse())
		})

		It("returns false when key not found", func() {
			deployKeyClient.GetReturns(nil, gitprovider.ErrNotFound)

			res, err := userProvider.DeployKeyExists("repo-name", "owner")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeFalse())
		})

		It("returns error when can't verify", func() {
			deployKeyClient.GetReturns(nil, errors.New("random error"))

			res, err := userProvider.DeployKeyExists("repo-name", "owner")
			Expect(err.Error()).Should(ContainSubstring("error getting deploy key"))
			Expect(res).To(BeFalse())
		})

		It("returns true when repo exists", func() {
			res, err := userProvider.DeployKeyExists("repo-name", "owner")
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeTrue())
		})
	})

	Describe("UploadDeployKey", func() {
		var deployKeyClient *fakegitprovider.DeployKeyClient

		BeforeEach(func() {
			deployKeyClient = &fakegitprovider.DeployKeyClient{}
			userRepo.DeployKeysReturns(deployKeyClient)
		})

		It("return error when repo doest exist", func() {
			userRepoClient.GetReturns(nil, gitprovider.ErrNotFound)

			err := userProvider.UploadDeployKey("owner", "repo-name", []byte("my-key"))
			Expect(err.Error()).Should(ContainSubstring("error getting user repo reference for owner"))
		})

		It("returns error when can't create the key", func() {
			deployKeyClient.CreateReturns(nil, errors.New("random error"))

			err := userProvider.UploadDeployKey("owner", "repo-name", []byte("my-key"))
			Expect(err.Error()).Should(ContainSubstring("error uploading deploy key"))
		})

		It("creates the deploy key", func() {
			deployKeyClient.CreateReturns(nil, nil)
			deployKeyClient.GetReturns(nil, nil)

			err := userProvider.UploadDeployKey("owner", "repo-name", []byte("my-key"))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("GetDefaultBranch", func() {
		It("returns error when cant extract owner from url", func() {
			_, err := userProvider.GetDefaultBranch("bad-url")
			Expect(err.Error()).Should(ContainSubstring("could get provider name from URL"))
		})

		It("returns error when can't get branch", func() {
			userRepoClient.GetReturns(nil, gitprovider.ErrNotFound)

			_, err := userProvider.GetDefaultBranch("http://github.com/owner/repo")
			Expect(err.Error()).Should(ContainSubstring("error getting user repository"))
		})

		It("returns repo default branch", func() {
			userRepo.GetReturns(gitprovider.RepositoryInfo{DefaultBranch: gitprovider.StringVar("my-branch")})

			branch, err := userProvider.GetDefaultBranch("http://github.com/owner/repo")
			Expect(err).ToNot(HaveOccurred())
			Expect(branch).To(Equal("my-branch"))
		})
	})

	Describe("GetRepoVisibility", func() {
		It("returns error when cant extract owner from url", func() {
			_, err := userProvider.GetRepoVisibility("bad-url")
			Expect(err.Error()).Should(ContainSubstring("could get provider name from URL"))
		})

		It("returns error when can't get branch", func() {
			userRepoClient.GetReturns(nil, gitprovider.ErrNotFound)

			_, err := userProvider.GetRepoVisibility("http://github.com/owner/repo")
			Expect(err.Error()).Should(ContainSubstring("error getting user repository"))
		})

		It("returns repo default branch", func() {
			visibility := gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityPrivate)
			userRepo.GetReturns(gitprovider.RepositoryInfo{Visibility: visibility})

			vis, err := userProvider.GetRepoVisibility("http://github.com/owner/repo")
			Expect(err).ToNot(HaveOccurred())
			Expect(vis).To(Equal(visibility))
		})
	})

	Describe("CreatePullRequest", func() {
		BeforeEach(func() {
			commit := &fakegitprovider.Commit{}
			commit.GetReturns(gitprovider.CommitInfo{Sha: "commit-sha"})

			commitClient.ListPageReturns([]gitprovider.Commit{commit}, nil)

			userRepo.GetReturns(gitprovider.RepositoryInfo{DefaultBranch: gitprovider.StringVar("my-branch")})
		})

		It("returns error when can't get repo", func() {
			userRepoClient.GetReturns(nil, errors.New("random error"))

			_, err := userProvider.CreatePullRequest("owner", "repo-name", "target-branch", "new-branch", []gitprovider.CommitFile{}, "commit-msg", "pr-title", "pr-desc")
			Expect(err.Error()).To(ContainSubstring("error getting user repo for"))
		})

		It("sets default branch", func() {
			_, err := userProvider.CreatePullRequest("owner", "repo-name", "", "new-branch", []gitprovider.CommitFile{}, "commit-msg", "pr-title", "pr-desc")
			Expect(err).ToNot(HaveOccurred())

			_, _, _, targetBranch, _ := pullRequestsClient.CreateArgsForCall(0)
			Expect(targetBranch).To(Equal("my-branch"))
		})

		It("returns error when unable to list commits", func() {
			commitClient.ListPageReturns(nil, errors.New("error"))

			_, err := userProvider.CreatePullRequest("owner", "repo-name", "target-branch", "new-branch", []gitprovider.CommitFile{}, "commit-msg", "pr-title", "pr-desc")
			Expect(err.Error()).To(ContainSubstring("error getting commits"))
		})

		It("returns error if no commits listed on target repo", func() {
			commitClient.ListPageReturns([]gitprovider.Commit{}, nil)

			_, err := userProvider.CreatePullRequest("owner", "repo-name", "target-branch", "new-branch", []gitprovider.CommitFile{}, "commit-msg", "pr-title", "pr-desc")
			Expect(err.Error()).To(ContainSubstring("no commits on the target branch"))
		})

		It("creates a branch", func() {
			_, err := userProvider.CreatePullRequest("owner", "repo-name", "target-branch", "new-branch", []gitprovider.CommitFile{}, "commit-msg", "pr-title", "pr-desc")
			Expect(err).ToNot(HaveOccurred())

			_, newBranch, sha := branchesClient.CreateArgsForCall(0)
			Expect(newBranch).To(Equal("new-branch"))
			Expect(sha).To(Equal("commit-sha"))
		})

		It("creates a commit", func() {
			_, err := userProvider.CreatePullRequest("owner", "repo-name", "target-branch", "new-branch", []gitprovider.CommitFile{{}}, "commit-msg", "pr-title", "pr-desc")
			Expect(err).ToNot(HaveOccurred())

			_, newBranch, commitMsg, files := commitClient.CreateArgsForCall(0)
			Expect(newBranch).To(Equal("new-branch"))
			Expect(commitMsg).To(Equal("commit-msg"))
			Expect(files).To(HaveLen(1))
		})

		It("creates a pull requests", func() {
			_, err := userProvider.CreatePullRequest("owner", "repo-name", "target-branch", "new-branch", []gitprovider.CommitFile{{}}, "commit-msg", "pr-title", "pr-desc")
			Expect(err).ToNot(HaveOccurred())

			_, prTitle, newBranch, targetBranch, prDescription := pullRequestsClient.CreateArgsForCall(0)
			Expect(prTitle).To(Equal("pr-title"))
			Expect(newBranch).To(Equal("new-branch"))
			Expect(targetBranch).To(Equal("target-branch"))
			Expect(prDescription).To(Equal("pr-desc"))
		})
	})

	Describe("GetCommits", func() {
		It("return error when repo doest exist", func() {
			userRepoClient.GetReturns(nil, gitprovider.ErrNotFound)

			_, err := userProvider.GetCommits("owner", "repo-name", "target-branch", 1, 1)
			Expect(err.Error()).Should(ContainSubstring("error getting repo"))
		})

		It("returns empty array when empty error", func() {
			commitClient.ListPageReturns(nil, errors.New("409 Git Repository is empty"))

			commits, err := userProvider.GetCommits("owner", "repo-name", "target-branch", 1, 1)
			Expect(err).ToNot(HaveOccurred())
			Expect(commits).To(HaveLen(0))
		})

		It("returns error when random error", func() {
			commitClient.ListPageReturns(nil, errors.New("error"))

			_, err := userProvider.GetCommits("owner", "repo-name", "target-branch", 1, 1)
			Expect(err.Error()).Should(ContainSubstring("error getting commits"))
		})

		It("returns a list of commits", func() {
			commit := &fakegitprovider.Commit{}
			commit.GetReturns(gitprovider.CommitInfo{Sha: "commit-sha"})

			commitClient.ListPageReturns([]gitprovider.Commit{commit}, nil)

			commits, err := userProvider.GetCommits("owner", "repo-name", "target-branch", 1, 1)
			Expect(err).ToNot(HaveOccurred())
			Expect(commits[0].Get().Sha).To(Equal("commit-sha"))
		})
	})

	Describe("GetProviderDomain", func() {
		It("returns provider domain", func() {
			gitProviderClient.ProviderIDReturns("github")

			Expect(userProvider.GetProviderDomain()).To(Equal("github.com"))
		})
	})
})
