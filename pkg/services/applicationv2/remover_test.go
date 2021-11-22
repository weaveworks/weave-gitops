package applicationv2

import (
	"context"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"

	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper/wrapperfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Remover", func() {
	Describe(".Remove()", func() {
		It("removes application auto-merge true", func() {
			ctx := context.Background()

			fakeGit := &wrapperfakes.FakeGit{}
			fakeGitClient := git.New(nil, fakeGit)

			orgProvider := &gitprovidersfakes.FakeGitProvider{}

			remover := NewRemover(fakeGitClient, orgProvider)

			err := remover.Remove(ctx, "my-app", "wego-system", true)
			Expect(err).NotTo(HaveOccurred())
		})
		It("removes application auto-merge false", func() {
			ctx := context.Background()

			fakeGit := &wrapperfakes.FakeGit{}
			fakeGitClient := git.New(nil, fakeGit)

			orgProvider := &gitprovidersfakes.FakeGitProvider{}

			remover := NewRemover(fakeGitClient, orgProvider)

			err := remover.Remove(ctx, "my-app", "wego-system", false)
			Expect(err).NotTo(HaveOccurred())
		})
	})
})
