package app

import (
	"time"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
)

var _ = Describe("Get Commits", func() {
	It("gets commits for a user repo", func() {
		commitParams := CommitParams{
			Name:      "test",
			Namespace: wego.DefaultNamespace,
		}

		application := &wego.Application{
			Spec: wego.ApplicationSpec{URL: "https://github.com/foo/bar"},
		}

		commits := []gitprovider.Commit{&fakeCommit{}}
		gitProviders.GetCommitsReturns(commits, nil)

		commit, err := appSrv.GetCommits(gitProviders, commitParams, application)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(len(commit)).To(Equal(1))
	})

	It("gets commits for an org", func() {
		commitParams := CommitParams{
			Name:      "test",
			Namespace: wego.DefaultNamespace,
		}

		application := &wego.Application{
			Spec: wego.ApplicationSpec{URL: "https://github.com/foo/bar"},
		}

		commits := []gitprovider.Commit{&fakeCommit{}}
		gitProviders.GetCommitsReturns(commits, nil)

		commit, err := appSrv.GetCommits(gitProviders, commitParams, application)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(len(commit)).To(Equal(1))
	})

	It("fails to get commits for helm", func() {
		commitParams := CommitParams{
			Name:      "test",
			Namespace: wego.DefaultNamespace,
		}

		application := &wego.Application{
			Spec: wego.ApplicationSpec{SourceType: wego.SourceTypeHelm},
		}

		_, err := appSrv.GetCommits(gitProviders, commitParams, application)
		Expect(err).Should(HaveOccurred())
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
