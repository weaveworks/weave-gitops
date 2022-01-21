package gitopswriter

import (
	"context"

	"github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"github.com/weaveworks/weave-gitops/pkg/services/gitrepo"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

var (
	ctx context.Context
)

func createRepoURL(url string) gitproviders.RepoURL {
	repoURL, err := gitproviders.NewRepoURL(url, false)
	Expect(err).NotTo(HaveOccurred())

	return repoURL
}

func createDirWriter() GitOpsDirectoryWriter {
	repoWriter := gitrepo.NewRepoWriter(app.ConfigRepo, gitProviders, gitClient, log)
	automationGen := automation.NewAutomationGenerator(gitProviders, fluxClient, log)

	return NewGitOpsDirectoryWriter(automationGen, repoWriter, osysClient, log)
}

var dummyGitSource = []byte(`---
apiVersion: source.toolkit.fluxcd.io/v1beta1
kind: GitRepository
metadata:
  name: wego-fork-test
  namespace: wego-system
spec:
  interval: 30s
  ref:
    branch: main
  secretRef:
    name: wego-test-cluster-wego-fork-test
  url: ssh://git@github.com/user/wego-fork-test.git`)

var dummyAppKustomization = []byte(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
metadata:
  name: bar
  namespace: wego-system
resources:
- app.yaml
- bar-gitops-deploy.yaml
- bar-gitops-source.yaml
`)

var dummyUserKustomization = []byte(`apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
metadata:
  name: test-cluster
  namespace: wego-system
resources:
- ../../../apps/bar
`)

var _ = Describe("Add", func() {
	var _ = BeforeEach(func() {
		app = models.Application{
			Name:                "bar",
			Namespace:           wego.DefaultNamespace,
			HelmSourceURL:       "",
			GitSourceURL:        createRepoURL("ssh://git@github.com/foo/bar.git"),
			Branch:              "main",
			Path:                "./kustomize",
			AutomationType:      models.AutomationTypeKustomize,
			SourceType:          models.SourceTypeGit,
			HelmTargetNamespace: "",
		}

		gitProviders.GetDefaultBranchReturns("main", nil)

		ctx = context.Background()
	})

	Context("add app with config in app repo", func() {
		BeforeEach(func() {
			app.GitSourceURL = createRepoURL("ssh://git@github.com/foo/bar.git")
			app.ConfigRepo = app.GitSourceURL
			gitOpsDirWriter = createDirWriter()

			gitClient.OpenStub = func(s string) (*gogit.Repository, error) {
				r, err := gogit.Init(memory.NewStorage(), memfs.New())
				Expect(err).ShouldNot(HaveOccurred())

				_, err = r.CreateRemote(&config.RemoteConfig{
					Name: "origin",
					URLs: []string{"git@github.com:foo/bar.git"},
				})
				Expect(err).ShouldNot(HaveOccurred())
				return r, nil
			}
		})

		Describe("generates application goat", func() {
			It("clones the repo to a temp dir", func() {
				err := gitOpsDirWriter.AddApplication(ctx, app, "test-cluster", true)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(gitClient.CloneCallCount()).To(Equal(1))
				_, repoDir, url, branch := gitClient.CloneArgsForCall(0)

				Expect(repoDir).To(ContainSubstring("user-repo-"))
				Expect(url).To(Equal("ssh://git@github.com/foo/bar.git"))
				Expect(branch).To(Equal("main"))
			})

			It("writes the files to the disk", func() {
				fluxClient.CreateSourceGitReturns(dummyGitSource, nil)
				fluxClient.CreateKustomizationReturns([]byte("kustomization"), nil)

				err := gitOpsDirWriter.AddApplication(ctx, app, "test-cluster", true)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(gitClient.WriteCallCount()).To(Equal(5))

				path, content := gitClient.WriteArgsForCall(0)
				Expect(path).To(Equal(".weave-gitops/apps/bar/app.yaml"))
				Expect(string(content)).To(ContainSubstring("kind: Application"))

				path, content = gitClient.WriteArgsForCall(1)
				Expect(path).To(Equal(".weave-gitops/apps/bar/bar-gitops-deploy.yaml"))
				Expect(content).To(Equal([]byte("kustomization")))

				path, content = gitClient.WriteArgsForCall(2)
				Expect(path).To(Equal(".weave-gitops/apps/bar/bar-gitops-source.yaml"))
				Expect(automation.AddWegoIgnore(dummyGitSource)).To(Equal(content))

				path, content = gitClient.WriteArgsForCall(3)
				Expect(path).To(Equal(".weave-gitops/apps/bar/kustomization.yaml"))
				Expect(content).To(Equal(dummyAppKustomization))

				path, content = gitClient.WriteArgsForCall(4)
				Expect(path).To(Equal(".weave-gitops/clusters/test-cluster/user/kustomization.yaml"))
				Expect(content).To(Equal(dummyUserKustomization))
			})

			It("commits and pushes the files", func() {
				err := gitOpsDirWriter.AddApplication(ctx, app, "test-cluster", true)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(gitClient.CommitCallCount()).To(Equal(1))

				msg, filters := gitClient.CommitArgsForCall(0)
				Expect(msg).To(Equal(git.Commit{
					Author:  git.Author{Name: "Weave Gitops", Email: "weave-gitops@weave.works"},
					Message: AddCommitMessage,
				}))

				Expect(len(filters)).To(Equal(1))
				Expect(filters[0](".weave-gitops/apps/bar/app.yaml")).To(BeTrue())
			})
		})
	})

	Context("add app with external config repo", func() {
		BeforeEach(func() {
			app.GitSourceURL = createRepoURL("https://github.com/user/repo")
			app.ConfigRepo = createRepoURL("https://github.com/foo/bar")
			gitOpsDirWriter = createDirWriter()
		})

		It("clones the repo to a temp dir", func() {
			err := gitOpsDirWriter.AddApplication(ctx, app, "test-cluster", true)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(gitClient.CloneCallCount()).To(Equal(1))
			_, repoDir, url, branch := gitClient.CloneArgsForCall(0)

			Expect(repoDir).To(ContainSubstring("user-repo-"))
			Expect(url).To(Equal("ssh://git@github.com/foo/bar.git"))
			Expect(branch).To(Equal("main"))
		})

		It("writes the files to the disk", func() {
			fluxClient.CreateSourceGitReturns(dummyGitSource, nil)
			fluxClient.CreateKustomizationReturns([]byte("kustomization"), nil)

			err := gitOpsDirWriter.AddApplication(ctx, app, "test-cluster", true)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(gitClient.WriteCallCount()).To(Equal(5))

			path, content := gitClient.WriteArgsForCall(0)
			Expect(path).To(Equal(".weave-gitops/apps/bar/app.yaml"))
			Expect(string(content)).To(ContainSubstring("kind: Application"))

			path, content = gitClient.WriteArgsForCall(1)
			Expect(path).To(Equal(".weave-gitops/apps/bar/bar-gitops-deploy.yaml"))
			Expect(content).To(Equal([]byte("kustomization")))

			path, content = gitClient.WriteArgsForCall(2)
			Expect(path).To(Equal(".weave-gitops/apps/bar/bar-gitops-source.yaml"))

			augmented, err := automation.AddWegoIgnore(dummyGitSource)
			Expect(err).To(BeNil())
			Expect(content).To(Equal(augmented))

			path, content = gitClient.WriteArgsForCall(3)
			Expect(path).To(Equal(".weave-gitops/apps/bar/kustomization.yaml"))
			Expect(content).To(Equal(dummyAppKustomization))

			path, content = gitClient.WriteArgsForCall(4)
			Expect(path).To(Equal(".weave-gitops/clusters/test-cluster/user/kustomization.yaml"))
			Expect(content).To(Equal(dummyUserKustomization))
		})

		It("commits and pushes the files", func() {
			err := gitOpsDirWriter.AddApplication(ctx, app, "test-cluster", true)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(gitClient.CommitCallCount()).To(Equal(1))

			msg, filters := gitClient.CommitArgsForCall(0)
			Expect(msg).To(Equal(git.Commit{
				Author:  git.Author{Name: "Weave Gitops", Email: "weave-gitops@weave.works"},
				Message: AddCommitMessage,
			}))

			Expect(len(filters)).To(Equal(1))
		})
	})

	Context("when using auto-merge", func() {
		BeforeEach(func() {
			gitProviders.GetDefaultBranchStub = func(_ context.Context, repoUrl gitproviders.RepoURL) (string, error) {
				addUrl := app.GitSourceURL

				if repoUrl.String() == addUrl.String() {
					return "default-app-branch", nil
				}
				return "default-config-branch", nil
			}

			app.GitSourceURL = createRepoURL("ssh://github.com/user/repo.git")
		})

		Context("uses the default app branch for config in app repository", func() {
			BeforeEach(func() {
				app.ConfigRepo = app.GitSourceURL
				gitOpsDirWriter = createDirWriter()
			})

			It("merges into the app default branch", func() {
				err := gitOpsDirWriter.AddApplication(ctx, app, "test-cluster", true)
				Expect(err).ShouldNot(HaveOccurred())

				_, _, _, branch := gitClient.CloneArgsForCall(0)
				Expect(branch).To(Equal("default-app-branch"))
			})
		})

		Context("uses the default config branch for external config", func() {
			BeforeEach(func() {
				app.ConfigRepo = createRepoURL("https://github.com/foo/bar")
				gitOpsDirWriter = createDirWriter()
			})

			It("merges into the config default branch", func() {
				err := gitOpsDirWriter.AddApplication(ctx, app, "test-cluster", true)
				Expect(err).ShouldNot(HaveOccurred())

				_, _, _, branch := gitClient.CloneArgsForCall(0)
				Expect(branch).To(Equal("default-config-branch"))
			})
		})
	})

	Context("when creating a pull request", func() {
		BeforeEach(func() {
			gitProviders.GetDefaultBranchStub = func(_ context.Context, repoUrl gitproviders.RepoURL) (string, error) {
				addUrl := app.GitSourceURL

				if repoUrl.String() == addUrl.String() {
					return "default-app-branch", nil
				}
				return "default-config-branch", nil
			}

			gitProviders.CreatePullRequestReturns(testutils.DummyPullRequest{}, nil)

			app.GitSourceURL = createRepoURL("ssh://github.com/user/repo.git")
		})

		Context("uses the default app branch for config in app repository", func() {
			BeforeEach(func() {
				app.ConfigRepo = app.GitSourceURL
				gitOpsDirWriter = createDirWriter()
			})

			It("creates the pull request against the app default branch", func() {
				err := gitOpsDirWriter.AddApplication(ctx, app, "test-cluster", false)
				Expect(err).ShouldNot(HaveOccurred())

				_, _, prInfo := gitProviders.CreatePullRequestArgsForCall(0)
				Expect(prInfo.TargetBranch).To(Equal("default-app-branch"))
			})
		})

		Context("uses the default config branch for external config", func() {
			BeforeEach(func() {
				app.ConfigRepo = createRepoURL("https://github.com/foo/bar")
				gitOpsDirWriter = createDirWriter()
			})

			It("creates the pull request against the config default branch", func() {
				err := gitOpsDirWriter.AddApplication(ctx, app, "test-cluster", false)
				Expect(err).ShouldNot(HaveOccurred())

				_, _, prInfo := gitProviders.CreatePullRequestArgsForCall(0)
				Expect(prInfo.TargetBranch).To(Equal("default-config-branch"))
			})
		})
	})
})
