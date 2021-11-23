package applicationv2

import (
	"context"
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"

	gogit "github.com/go-git/go-git/v5"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper/wrapperfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type localGitRepo struct {
	wrapper.Git
	path string
}

func (lg *localGitRepo) PlainOpen(_ string) (*gogit.Repository, error) {
	return gogit.PlainOpen(lg.path)
}

var _ = Describe("Remover", func() {
	Describe(".Remove()", func() {
		It("removes application auto-merge true", func() {
			ctx := context.Background()

			dir := os.TempDir()
			gc := git.New(nil, &localGitRepo{Git: wrapper.NewGoGit(), path: dir})

			appPath := fmt.Sprintf("%s/.wego-system/apps/my-app.yaml", dir)

			Expect(gc.Open(dir)).To(Succeed())
			Expect(gc.Write(appPath, []byte("---\nsome-app-yaml"))).To(Succeed())

			orgProvider := &gitprovidersfakes.FakeGitProvider{}

			remover := NewRemover(gc, orgProvider)

			err := remover.Remove(ctx, "my-app", "wego-system", true)
			Expect(err).NotTo(HaveOccurred())

			_, err = os.Stat(appPath)
			Expect(err).NotTo(BeNil(), "error should not be nil")
			Expect(os.IsNotExist(err)).To(BeTrue(), "should have had an IsNotExist error")
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
