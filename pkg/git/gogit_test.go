package git_test

import (
	"context"
	"io/ioutil"
	"os"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/git"
)

var (
	gitClient git.Git
	dir       string
	err       error
)

// var _ = AfterSuite(func() {
// 	err := os.RemoveAll("/tmp/wego-git-test-*")
// 	Expect(err).ShouldNot(HaveOccurred())
// })

var _ = BeforeEach(func() {
	gitClient = git.New(nil)

	dir, err = ioutil.TempDir("", "wego-git-test-")
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = AfterEach(func() {
	err := os.RemoveAll(dir)
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = Describe("Init", func() {
	It("creates an empty repository", func() {
		_, err := gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())

		_, err = gitClient.Open(dir)
		Expect(err).ShouldNot(HaveOccurred())
	})
})

var _ = Describe("Clone", func() {
	It("clones a given repository", func() {
		_, err := gitClient.Clone(context.Background(), dir, "https://github.com/githubtraining/hellogitworld", "master")
		Expect(err).ShouldNot(HaveOccurred())

		_, err = os.Stat(dir + "/README.txt")
		Expect(err).ShouldNot(HaveOccurred())
	})
})

var _ = Describe("Write", func() {
	It("writes a file into a given repository", func() {
		_, err = gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())
		filePath := "/test.txt"
		content := []byte("testing")
		err = gitClient.Write(filePath, content)

		fileContent, err := ioutil.ReadFile(dir + filePath)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(content).To(Equal(fileContent))
	})
})

var _ = Describe("Commit", func() {
	It("commits into a given repository", func() {
		_, err = gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())

		filePath := "/test.txt"
		content := []byte("testing")
		err = gitClient.Write(filePath, content)

		_, err = gitClient.Commit(git.Commit{
			Author:  git.Author{Name: "test", Email: "test@example.com"},
			Message: "test commit",
		})
		Expect(err).ShouldNot(HaveOccurred())

		err = os.Chdir(dir)
		Expect(err).ShouldNot(HaveOccurred())

		out, err := exec.Command("sh", "-c", `git log -1 --pretty=%B`).Output()
		Expect(err).ShouldNot(HaveOccurred())

		Expect(string(out)).To(ContainSubstring("test commit"))
	})
})
