package app

import (
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
)

var _ = Describe("Get Commits", func() {
	It("gets commits for a user app", func() {
		commitParams := CommitParams{
			Name:      "test",
			Namespace: "wego-system",
		}

		application := &wego.Application{
			Spec: wego.ApplicationSpec{URL: "https://github.com/foo/bar"},
		}

		gitProviders.GetAccountTypeStub = func(s string) (gitproviders.ProviderAccountType, error) {
			return gitproviders.AccountTypeUser, nil
		}

		commits := []gitprovider.Commit{&fakeCommit{}}
		gitProviders.GetCommitsFromUserRepoStub = func(gitprovider.UserRepositoryRef, string, int, int) ([]gitprovider.Commit, error) {
			return commits, nil
		}

		commit, err := appSrv.GetCommits(commitParams, application)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(len(commit)).To(Equal(1))
	})

	It("gets commits for a user app", func() {
		commitParams := CommitParams{
			Name:      "test",
			Namespace: "wego-system",
		}

		application := &wego.Application{
			Spec: wego.ApplicationSpec{URL: "https://github.com/foo/bar"},
		}

		gitProviders.GetAccountTypeStub = func(s string) (gitproviders.ProviderAccountType, error) {
			return gitproviders.AccountTypeOrg, nil
		}

		commits := []gitprovider.Commit{&fakeCommit{}}
		gitProviders.GetCommitsFromOrgRepoStub = func(gitprovider.OrgRepositoryRef, string, int, int) ([]gitprovider.Commit, error) {
			return commits, nil
		}

		commit, err := appSrv.GetCommits(commitParams, application)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(len(commit)).To(Equal(1))
	})

	It("fails to get commits for helm", func() {
		commitParams := CommitParams{
			Name:      "test",
			Namespace: "wego-system",
		}

		application := &wego.Application{
			Spec: wego.ApplicationSpec{SourceType: wego.SourceTypeHelm},
		}

		_, err := appSrv.GetCommits(commitParams, application)
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("unable to get commits for a helm chart"))
	})

	It("fails to get commits when config_url set to NONE", func() {
		commitParams := CommitParams{
			Name:      "test",
			Namespace: "wego-system",
		}

		application := &wego.Application{
			Spec: wego.ApplicationSpec{ConfigURL: "NONE"},
		}

		_, err := appSrv.GetCommits(commitParams, application)
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal("unable to get commits when config_url is empty"))
	})
})

type fakeCommit struct {
	commitInfo gitprovider.CommitInfo
}

func (fc *fakeCommit) APIObject() interface{} {
	return &fc.commitInfo
}

func (fc *fakeCommit) Get() gitprovider.CommitInfo {
	return testCommit()
}

func testCommit() gitprovider.CommitInfo {
	return gitprovider.CommitInfo{
		Sha:       "23498987239879892348768",
		Author:    "testauthor",
		Message:   "if a message is above fifty characters then it will be truncated",
		CreatedAt: time.Now(),
		URL:       "http://github.com/testrepo/commit/2349898723987989234",
	}
}
