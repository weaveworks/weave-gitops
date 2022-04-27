package git_test

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/weaveworks/weave-gitops/gitops/pkg/git/wrapper/wrapperfakes"

	"github.com/weaveworks/weave-gitops/gitops/pkg/git/wrapper"

	gogit "github.com/go-git/go-git/v5"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/gitops/pkg/git"
)

var (
	gitClient     git.Git
	dir           string
	err           error
	fakeGit       *wrapperfakes.FakeGit
	fakeGitClient git.Git
)

var _ = BeforeEach(func() {
	gitClient = git.New(nil, wrapper.NewGoGit())

	fakeGit = &wrapperfakes.FakeGit{}
	fakeGitClient = git.New(nil, fakeGit)

	dir, err = ioutil.TempDir("", "wego-git-test-")
	Expect(err).ShouldNot(HaveOccurred())
})

var _ = AfterEach(func() {
	Expect(os.RemoveAll(dir)).To(Succeed())
})

var _ = Describe("Init", func() {
	It("creates an empty repository", func() {
		_, err := gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())

		_, err = gitClient.Open(dir)
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("does nothing when the repository has already been initialized", func() {
		init, err := gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(init).To(BeTrue())

		init, err = gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(init).To(BeFalse())
	})

	It("returns an error when the directory already contains a repository", func() {
		_, err := gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())

		init, err := git.New(nil, wrapper.NewGoGit()).Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).Should(MatchError("repository already exists"))
		Expect(init).To(BeFalse())
	})
})

var _ = Describe("Open", func() {
	It("fails when the directory is an empty directory", func() {
		_, err := gitClient.Open(dir)
		Expect(err).To(MatchError("repository does not exist"))
	})
})

var _ = Describe("Clone", func() {
	It("clones a given repository", func() {
		_, err := gitClient.Clone(context.Background(), dir, "https://github.com/githubtraining/hellogitworld", "master")
		Expect(err).ShouldNot(HaveOccurred())

		_, err = os.Stat(dir + "/README.txt")
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("initialize a given repository if remote branch not found", func() {
		_, err := gitClient.Clone(context.Background(), dir, "https://github.com/githubtraining/hellogitworld", "new-branch")
		Expect(err).ShouldNot(HaveOccurred())

		_, err = gitClient.Open(dir)
		Expect(err).ShouldNot(HaveOccurred())
	})
})

var _ = Describe("ValidateAccess", func() {
	It("validate access to a given repository successfully", func() {
		err := gitClient.ValidateAccess(context.Background(), "https://github.com/githubtraining/hellogitworld", "master")
		Expect(err).ShouldNot(HaveOccurred())
	})

	It("should not fail on an empty repo", func() {

		fakeGit.PlainCloneContextReturns(nil, transport.ErrEmptyRemoteRepository)

		err := fakeGitClient.ValidateAccess(context.Background(), "https://github.com/githubtraining/hellogitworld", "master")

		Expect(err).ShouldNot(HaveOccurred())
	})

	It("should fail with custom error", func() {

		customError := errors.New("my-custom-error")

		fakeGit.PlainCloneContextReturns(nil, customError)

		err := fakeGitClient.ValidateAccess(context.Background(), "https://github.com/githubtraining/hellogitworld", "master")
		Expect(err).Should(HaveOccurred())
		Expect(err).Should(Equal(fmt.Errorf("error validating git repo access %w", customError)))
	})

	It("should fail to create temporary directory", func() {

		tempDirName := "-*/3486rw7f"

		os.Setenv("TMPDIR", tempDirName)
		defer os.Unsetenv("TMPDIR")

		err := gitClient.ValidateAccess(context.Background(), "https://github.com/githubtraining/hellogitworld", "master")
		Expect(err).Should(HaveOccurred())
		Expect(err.Error()).Should(Equal(fmt.Sprintf("error creating temporary folder stat %s: no such file or directory", tempDirName)))
	})

	It("fails to validate access to a possible private repository", func() {
		err := gitClient.ValidateAccess(context.Background(), "https://github.com/notexisted/repo", "master")
		Expect(err.Error()).Should(Equal("error validating git repo access authentication required"))
	})
})

var _ = Describe("Read", func() {
	It("Reads a file from a repo", func() {
		_, err = gitClient.Init(dir, "https://github.com/github/gitignore", "master")

		filePath := "/test.txt"
		content := []byte("testing")
		Expect(gitClient.Write(filePath, content)).To(Succeed())

		_, err = gitClient.Commit(git.Commit{
			Author:  git.Author{Name: "test", Email: "test@example.com"},
			Message: "test commit",
		})
		Expect(err).ShouldNot(HaveOccurred())

		fileContent, err := gitClient.Read(filePath)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(content).To(Equal(fileContent))
	})

	It("Reads a file from a repo even if not commited", func() {
		_, err = gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		filePath := "/test.txt"
		content := []byte("testing")
		err := ioutil.WriteFile(dir+filePath, content, 0644)
		Expect(err).ShouldNot(HaveOccurred())
		fileContent, err := gitClient.Read(filePath)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(content).To(Equal(fileContent))
	})

	It("gives a nice error when file not present", func() {
		_, err = gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		filePath := "/test.txt"
		_, err := gitClient.Read(filePath)
		Expect(fmt.Sprint(err)).To(MatchRegexp("failed to open file /test.txt"))
	})

	It("fails when the repository has not been initialized", func() {
		_, err := gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())

		gc := git.New(nil, wrapper.NewGoGit())
		_, err = gc.Read("testing.txt")
		Expect(err).To(MatchError("no git repository"))
	})

	It("returns an error if the repository is bare", func() {
		_, err = gogit.PlainInit(dir, true)
		Expect(err).ShouldNot(HaveOccurred())
		_, err = gitClient.Open(dir)
		Expect(err).ShouldNot(HaveOccurred())

		_, err := gitClient.Read("testing.txt")
		Expect(err).Should(MatchError("failed to open the worktree: worktree not available in a bare repository"))
	})
})

var _ = Describe("Write", func() {
	It("writes a file into a given repository", func() {
		_, err = gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())
		filePath := "/test.txt"
		content := []byte("testing")
		Expect(gitClient.Write(filePath, content)).To(Succeed())

		fileContent, err := ioutil.ReadFile(dir + filePath)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(content).To(Equal(fileContent))
	})

	It("fails when the repository has not been initialized", func() {
		_, err := gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())

		gc := git.New(nil, wrapper.NewGoGit())
		err = gc.Write("testing.txt", []byte("testing"))
		Expect(err).To(MatchError("no git repository"))
	})

	It("returns an error if the repository is bare", func() {
		_, err = gogit.PlainInit(dir, true)
		Expect(err).ShouldNot(HaveOccurred())
		_, err = gitClient.Open(dir)
		Expect(err).ShouldNot(HaveOccurred())

		err := gitClient.Write("testing.txt", []byte("testing"))
		Expect(err).Should(MatchError("failed to open the worktree: worktree not available in a bare repository"))
	})
})

var _ = Describe("Commit", func() {
	It("commits into a given repository", func() {
		_, err = gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())

		filePath := "/test.txt"
		content := []byte("testing")
		Expect(gitClient.Write(filePath, content)).To(Succeed())

		_, err = gitClient.Commit(git.Commit{
			Author:  git.Author{Name: "test", Email: "test@example.com"},
			Message: "test commit",
		})
		Expect(err).ShouldNot(HaveOccurred())

		out := executeCommand(dir, "sh", "-c", `git log -1 --pretty=%B`)
		Expect(out).To(ContainSubstring("test commit"))
	})

	It("commits into a given repository skipping filtered files", func() {
		_, err = gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())

		filePath := "/test.txt"
		content := []byte("testing")
		Expect(gitClient.Write(filePath, content)).To(Succeed())

		_, err = gitClient.Commit(git.Commit{
			Author:  git.Author{Name: "test", Email: "test@example.com"},
			Message: "test commit",
		},
			func(fname string) bool {
				return false
			})
		Expect(err).Should(HaveOccurred())
		isClean, err := gitClient.Status()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(isClean).To(BeFalse())
	})

	It("commits into a given repository skipping filtered files on .wego folder", func() {
		_, err = gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())

		filePath := ".wego/test.txt"
		content := []byte("testing")
		Expect(gitClient.Write(filePath, content)).To(Succeed())

		_, err = gitClient.Commit(git.Commit{
			Author:  git.Author{Name: "test", Email: "test@example.com"},
			Message: "test commit",
		},
			func(fname string) bool {
				return strings.HasPrefix(fname, ".wego")
			})
		Expect(err).ShouldNot(HaveOccurred())
		isClean, err := gitClient.Status()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(isClean).To(BeTrue())
	})

	It("fails if there are no changes to commit", func() {
		_, err = gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())
		filePath := "/test.txt"
		content := []byte("testing")
		Expect(gitClient.Write(filePath, content)).To(Succeed())

		_, err = gitClient.Commit(git.Commit{
			Author:  git.Author{Name: "test", Email: "test@example.com"},
			Message: "test commit",
		})
		Expect(err).ShouldNot(HaveOccurred())

		out := executeCommand(dir, "sh", "-c", `git log -1 --pretty=%B`)
		Expect(out).To(ContainSubstring("test commit"))

		_, err = gitClient.Commit(git.Commit{
			Author:  git.Author{Name: "test", Email: "test@example.com"},
			Message: "test commit",
		})
		Expect(err).Should(MatchError("no staged files"))
	})
})

var _ = Describe("Status", func() {
	It("returns true if no files have been changed in the repository", func() {
		_, err = gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())

		isClean, err := gitClient.Status()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(isClean).To(BeTrue())
	})

	It("returns false if a file has been changed in the repository", func() {
		_, err = gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())

		filePath := "/test.txt"
		content := []byte("testing")
		Expect(gitClient.Write(filePath, content)).To(Succeed())

		isClean, err := gitClient.Status()
		Expect(err).ShouldNot(HaveOccurred())
		Expect(isClean).To(BeFalse())
	})

	It("fails if not initialized", func() {
		_, err := gitClient.Status()
		Expect(err).Should(MatchError("no git repository"))
	})

	It("returns an error if the repository is bare", func() {
		_, err = gogit.PlainInit(dir, true)
		Expect(err).ShouldNot(HaveOccurred())
		_, err = gitClient.Open(dir)
		Expect(err).ShouldNot(HaveOccurred())
		_, err := gitClient.Status()
		Expect(err).Should(MatchError("failed to open the worktree: worktree not available in a bare repository"))
	})
})

var _ = Describe("Head", func() {
	It("returns if the working tree is clean", func() {
		_, err = gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())

		filePath := "/test.txt"
		content := []byte("testing")
		Expect(gitClient.Write(filePath, content)).To(Succeed())

		_, err = gitClient.Commit(git.Commit{
			Author:  git.Author{Name: "test", Email: "test@example.com"},
			Message: "test commit",
		})
		Expect(err).ShouldNot(HaveOccurred())

		hash, err := gitClient.Head()
		Expect(err).ShouldNot(HaveOccurred())

		out := executeCommand(dir, "sh", "-c", `git log -1`)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(out).To(ContainSubstring(hash))
	})

	It("fails if not initialized", func() {
		_, err := gitClient.Head()
		Expect(err).Should(MatchError("no git repository"))
	})

	It("fails if no commits in the git repository", func() {
		_, err := gogit.PlainInit(dir, true)
		Expect(err).ShouldNot(HaveOccurred())
		_, err = gitClient.Open(dir)
		Expect(err).ShouldNot(HaveOccurred())

		_, err = gitClient.Head()
		Expect(err.Error()).Should(ContainSubstring("reference not found"))
	})
})

var _ = Describe("Push", func() {
	It("fails if not access", func() {
		_, err = gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())

		filePath := "/test.txt"
		content := []byte("testing")
		Expect(gitClient.Write(filePath, content)).To(Succeed())

		_, err = gitClient.Commit(git.Commit{
			Author:  git.Author{Name: "test", Email: "test@example.com"},
			Message: "test commit",
		})
		Expect(err).ShouldNot(HaveOccurred())

		err := gitClient.Push(context.Background())
		Expect(err).Should(HaveOccurred())
	})

	It("fails if not initialized", func() {
		err := gitClient.Push(context.Background())
		Expect(err).Should(MatchError("no git repository"))
	})

	It("fails if no commits in the git repository", func() {
		_, err := gogit.PlainInit(dir, true)
		Expect(err).ShouldNot(HaveOccurred())
		_, err = gitClient.Open(dir)
		Expect(err).ShouldNot(HaveOccurred())

		err = gitClient.Push(context.Background())
		Expect(err).Should(MatchError("remote not found"))
	})
})

var _ = Describe("Remove", func() {
	It("fails if no file present at path in the git repository", func() {
		_, err = gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(gitClient.Remove("foo")).ShouldNot(Succeed())
	})

	It("succeeds if file present at path in the git repository", func() {
		_, err = gitClient.Init(dir, "https://github.com/github/gitignore", "master")
		Expect(err).ShouldNot(HaveOccurred())
		Expect(gitClient.Write("foo", []byte("bar"))).To(Succeed())
		Expect(gitClient.Remove("foo")).To(Succeed())
	})
})

var _ = Describe("Checkout", func() {
	It("succeeds", func() {

		_, err = gitClient.Clone(context.Background(), dir, "https://github.com/github/gitignore", "main")
		Expect(err).ShouldNot(HaveOccurred())

		err = gitClient.Checkout("new-branch")
		Expect(err).ShouldNot(HaveOccurred())

		Expect(gitClient.Write("foo", []byte("bar"))).To(Succeed())

		_, err = os.Stat(filepath.Join(dir, "foo"))
		Expect(err).ShouldNot(HaveOccurred())

		_, err = gitClient.Commit(git.Commit{
			Message: "test",
			Author: git.Author{
				Name:  "test",
				Email: "test@test.com",
			}},
			func(s string) bool {
				return true
			})
		Expect(err).ShouldNot(HaveOccurred())

		err = gitClient.Checkout("main")
		Expect(err).ShouldNot(HaveOccurred())

		_, err := os.Stat(filepath.Join(dir, "foo"))
		Expect(errors.Is(err, os.ErrNotExist)).To(Equal(true))
	})
})

func executeCommand(workingDir, cmd string, args ...string) []byte {
	c := exec.Command(cmd, args...)
	c.Dir = workingDir
	out, err := c.Output()
	Expect(err).ShouldNot(HaveOccurred())

	return out
}
