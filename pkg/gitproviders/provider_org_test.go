package gitproviders

import (
	"context"
	"errors"
	"fmt"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/vendorfakes/fakegitprovider"
)

var _ = Describe("Org Provider", func() {
	var (
		ctx         context.Context
		orgProvider GitProvider

		gitProviderClient *fakegitprovider.Client
		orgRepoClient     *fakegitprovider.OrgRepositoriesClient
		orgRepo           *fakegitprovider.OrgRepository

		commitClient       *fakegitprovider.CommitClient
		branchesClient     *fakegitprovider.BranchClient
		pullRequestsClient *fakegitprovider.PullRequestClient
		fileClient         *fakegitprovider.FileClient

		repoUrl RepoURL
	)

	var _ = BeforeEach(func() {
		commitClient = &fakegitprovider.CommitClient{}
		branchesClient = &fakegitprovider.BranchClient{}
		pullRequestsClient = &fakegitprovider.PullRequestClient{}
		fileClient = &fakegitprovider.FileClient{}

		orgRepo = &fakegitprovider.OrgRepository{}
		orgRepo.CommitsReturns(commitClient)
		orgRepo.BranchesReturns(branchesClient)
		orgRepo.PullRequestsReturns(pullRequestsClient)
		orgRepo.FilesReturns(fileClient)

		orgRepoClient = &fakegitprovider.OrgRepositoriesClient{}
		orgRepoClient.GetReturns(orgRepo, nil)

		gitProviderClient = &fakegitprovider.Client{}
		gitProviderClient.OrgRepositoriesReturns(orgRepoClient)

		orgProvider = orgGitProvider{
			domain:   "github.com",
			provider: gitProviderClient,
		}

		var err error
		repoUrl, err = NewRepoURL("http://github.com/owner/repo-name")
		Expect(err).ToNot(HaveOccurred())
	})

	Describe("RepositoryExists", func() {
		It("returns false when repo not found", func() {
			orgRepoClient.GetReturns(nil, gitprovider.ErrNotFound)

			res, err := orgProvider.RepositoryExists(ctx, repoUrl)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeFalse())
		})

		It("returns error when can't verify", func() {
			orgRepoClient.GetReturns(nil, errors.New("random error"))

			res, err := orgProvider.RepositoryExists(ctx, repoUrl)
			Expect(err).To(HaveOccurred())
			Expect(res).To(BeFalse())
		})

		It("returns true when repo exists", func() {
			res, err := orgProvider.RepositoryExists(ctx, repoUrl)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeTrue())
		})
	})

	Describe("DeployKeyExists", func() {
		var deployKeyClient *fakegitprovider.DeployKeyClient

		BeforeEach(func() {
			deployKeyClient = &fakegitprovider.DeployKeyClient{}
			orgRepo.DeployKeysReturns(deployKeyClient)
		})

		It("return error when repo doest exist", func() {
			orgRepoClient.GetReturns(nil, gitprovider.ErrNotFound)

			res, err := orgProvider.DeployKeyExists(ctx, repoUrl)
			Expect(err.Error()).Should(ContainSubstring("error getting org repo reference for owner"))
			Expect(res).To(BeFalse())
		})

		It("returns false when key not found", func() {
			deployKeyClient.GetReturns(nil, gitprovider.ErrNotFound)

			res, err := orgProvider.DeployKeyExists(ctx, repoUrl)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeFalse())
		})

		It("returns error when can't verify", func() {
			deployKeyClient.GetReturns(nil, errors.New("random error"))

			res, err := orgProvider.DeployKeyExists(ctx, repoUrl)
			Expect(err.Error()).Should(ContainSubstring("error getting deploy key"))
			Expect(res).To(BeFalse())
		})

		It("returns true when repo exists", func() {
			res, err := orgProvider.DeployKeyExists(ctx, repoUrl)
			Expect(err).ToNot(HaveOccurred())
			Expect(res).To(BeTrue())
		})
	})

	Describe("UploadDeployKey", func() {
		var deployKeyClient *fakegitprovider.DeployKeyClient

		BeforeEach(func() {
			deployKeyClient = &fakegitprovider.DeployKeyClient{}
			orgRepo.DeployKeysReturns(deployKeyClient)
		})

		It("return error when repo doesn't exist", func() {
			orgRepoClient.GetReturns(nil, gitprovider.ErrNotFound)

			err := orgProvider.UploadDeployKey(ctx, repoUrl, []byte("my-key"))
			Expect(err.Error()).Should(ContainSubstring("error getting org repo reference for owner"))
		})

		It("returns error when can't create the key", func() {
			deployKeyClient.CreateReturns(nil, errors.New("random error"))

			err := orgProvider.UploadDeployKey(ctx, repoUrl, []byte("my-key"))
			Expect(err.Error()).Should(ContainSubstring("error uploading deploy key"))
		})

		It("return error when it fails to upload a deploy key", func() {
			deployKeyClient.CreateReturns(nil, gitprovider.ErrNotFound)

			err := orgProvider.UploadDeployKey(ctx, repoUrl, []byte("my-key"))
			Expect(err).Should(Equal(ErrRepositoryNoPermissionsOrDoesNotExist))
		})

		It("creates the deploy key", func() {
			deployKeyClient.CreateReturns(nil, nil)
			deployKeyClient.GetReturns(nil, nil)

			err := orgProvider.UploadDeployKey(ctx, repoUrl, []byte("my-key"))
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Describe("GetDefaultBranch", func() {
		It("returns error when can't get branch", func() {
			orgRepoClient.GetReturns(nil, gitprovider.ErrNotFound)

			_, err := orgProvider.GetDefaultBranch(ctx, repoUrl)
			Expect(err.Error()).Should(ContainSubstring("error getting org repository"))
		})

		It("returns repo default branch", func() {
			orgRepo.GetReturns(gitprovider.RepositoryInfo{DefaultBranch: gitprovider.StringVar("my-branch")})

			branch, err := orgProvider.GetDefaultBranch(ctx, repoUrl)
			Expect(err).ToNot(HaveOccurred())
			Expect(branch).To(Equal("my-branch"))
		})
	})

	Describe("GetRepoVisibility", func() {
		It("returns error when can't get branch", func() {
			orgRepoClient.GetReturns(nil, gitprovider.ErrNotFound)

			_, err := orgProvider.GetRepoVisibility(ctx, repoUrl)
			Expect(err.Error()).Should(ContainSubstring("error getting org repository"))
		})

		It("returns repo default branch", func() {
			visibility := gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityPrivate)
			orgRepo.GetReturns(gitprovider.RepositoryInfo{Visibility: visibility})

			vis, err := orgProvider.GetRepoVisibility(ctx, repoUrl)
			Expect(err).ToNot(HaveOccurred())
			Expect(vis).To(Equal(visibility))
		})
	})

	Describe("CreatePullRequest", func() {
		var prInfo PullRequestInfo

		BeforeEach(func() {
			commit := &fakegitprovider.Commit{}
			commit.GetReturns(gitprovider.CommitInfo{Sha: "commit-sha"})

			commitClient.ListPageReturns([]gitprovider.Commit{commit}, nil)

			orgRepo.GetReturns(gitprovider.RepositoryInfo{DefaultBranch: gitprovider.StringVar("my-branch")})

			prInfo = PullRequestInfo{
				Title:         "pr-title",
				Description:   "pr-desc",
				CommitMessage: "commit-msg",
				TargetBranch:  "target-branch",
				NewBranch:     "new-branch",
				Files:         []gitprovider.CommitFile{},
			}
		})

		It("returns error when can't get repo", func() {
			orgRepoClient.GetReturns(nil, errors.New("random error"))

			_, err := orgProvider.CreatePullRequest(ctx, repoUrl, prInfo)
			Expect(err.Error()).To(ContainSubstring("error getting org repo for owner"))
		})

		It("sets default branch", func() {
			prInfo.TargetBranch = ""
			_, err := orgProvider.CreatePullRequest(ctx, repoUrl, prInfo)
			Expect(err).ToNot(HaveOccurred())

			_, _, _, targetBranch, _ := pullRequestsClient.CreateArgsForCall(0)
			Expect(targetBranch).To(Equal("my-branch"))
		})

		It("returns error when unable to list commits", func() {
			commitClient.ListPageReturns(nil, errors.New("error"))

			_, err := orgProvider.CreatePullRequest(ctx, repoUrl, prInfo)
			Expect(err.Error()).To(ContainSubstring("error getting commits"))
		})

		It("returns error if no commits listed on target repo", func() {
			commitClient.ListPageReturns([]gitprovider.Commit{}, nil)

			_, err := orgProvider.CreatePullRequest(ctx, repoUrl, prInfo)
			Expect(err.Error()).To(ContainSubstring("no commits on the target branch"))
		})

		It("creates a branch", func() {
			_, err := orgProvider.CreatePullRequest(ctx, repoUrl, prInfo)
			Expect(err).ToNot(HaveOccurred())

			_, newBranch, sha := branchesClient.CreateArgsForCall(0)
			Expect(newBranch).To(Equal("new-branch"))
			Expect(sha).To(Equal("commit-sha"))
		})

		It("creates a commit", func() {
			prInfo.Files = []gitprovider.CommitFile{{}}
			_, err := orgProvider.CreatePullRequest(ctx, repoUrl, prInfo)
			Expect(err).ToNot(HaveOccurred())

			_, newBranch, commitMsg, files := commitClient.CreateArgsForCall(0)
			Expect(newBranch).To(Equal("new-branch"))
			Expect(commitMsg).To(Equal("commit-msg"))
			Expect(files).To(HaveLen(1))
		})

		It("creates a pull requests", func() {
			_, err := orgProvider.CreatePullRequest(ctx, repoUrl, prInfo)
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
			orgRepoClient.GetReturns(nil, gitprovider.ErrNotFound)

			_, err := orgProvider.GetCommits(ctx, repoUrl, "target-branch", 1, 1)
			Expect(err.Error()).Should(ContainSubstring("error getting repo"))
		})

		It("returns empty array when empty error", func() {
			commitClient.ListPageReturns(nil, errors.New("409 Git Repository is empty"))

			commits, err := orgProvider.GetCommits(ctx, repoUrl, "target-branch", 1, 1)
			Expect(err).ToNot(HaveOccurred())
			Expect(commits).To(HaveLen(0))
		})

		It("returns error when random error", func() {
			commitClient.ListPageReturns(nil, errors.New("error"))

			_, err := orgProvider.GetCommits(ctx, repoUrl, "target-branch", 1, 1)
			Expect(err.Error()).Should(ContainSubstring("error getting commits"))
		})

		It("returns a list of commits", func() {
			commit := &fakegitprovider.Commit{}
			commit.GetReturns(gitprovider.CommitInfo{Sha: "commit-sha"})

			commitClient.ListPageReturns([]gitprovider.Commit{commit}, nil)

			commits, err := orgProvider.GetCommits(ctx, repoUrl, "target-branch", 1, 1)
			Expect(err).ToNot(HaveOccurred())
			Expect(commits[0].Get().Sha).To(Equal("commit-sha"))
		})
	})

	Describe("GetProviderDomain", func() {
		It("returns provider domain", func() {
			gitProviderClient.ProviderIDReturns("github")

			Expect(orgProvider.GetProviderDomain()).To(Equal("github.com"))
		})
	})

	Describe("GetRepoDirFiles", func() {
		It("returns a list of files", func() {
			file := &gitprovider.CommitFile{}
			fileClient.GetReturns([]*gitprovider.CommitFile{file}, nil)
			c, err := orgProvider.GetRepoDirFiles(context.TODO(), repoUrl, "path", "main")
			Expect(err).NotTo(HaveOccurred())
			Expect(c).To(Equal([]*gitprovider.CommitFile{file}))
			Expect(fileClient.GetCallCount()).To(Equal(1))
			_, dirPath, targetBranch := fileClient.GetArgsForCall(0)
			Expect(dirPath).To(Equal("path"))
			Expect(targetBranch).To(Equal("main"))
		})

		When("it fails to get the requested directory from the repo", func() {
			It("returns an error", func() {
				file := &gitprovider.CommitFile{}
				fileClient.GetReturns([]*gitprovider.CommitFile{file}, fmt.Errorf("err"))
				_, err := orgProvider.GetRepoDirFiles(context.TODO(), repoUrl, "path", "main")
				Expect(err).To(MatchError("err"))
				Expect(fileClient.GetCallCount()).To(Equal(1))
			})
		})
	})

	Describe("MergePullRequest", func() {
		It("merges a given pull request", func() {
			pullRequestsClient.MergeReturns(nil)
			err := orgProvider.MergePullRequest(context.TODO(), repoUrl, 1, "message")
			Expect(err).NotTo(HaveOccurred())
			Expect(pullRequestsClient.MergeCallCount()).To(Equal(1))
			_, prNumber, mergeMethod, message := pullRequestsClient.MergeArgsForCall(0)
			Expect(prNumber).To(Equal(1))
			Expect(mergeMethod).To(Equal(gitprovider.MergeMethodMerge))
			Expect(message).To(Equal("message"))
		})

		When("merge the PR fails", func() {
			It("returns an error", func() {
				pullRequestsClient.MergeReturns(fmt.Errorf("err"))
				err := orgProvider.MergePullRequest(context.TODO(), repoUrl, 1, "message")
				Expect(err).To(MatchError("err"))
				Expect(pullRequestsClient.MergeCallCount()).To(Equal(1))
			})
		})
	})
})
