package app

import (
	"github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
)

var (
	gitClient    *gitfakes.FakeGit
	fluxClient   *fluxfakes.FakeFlux
	kubeClient   *kubefakes.FakeKube
	gitProviders *gitprovidersfakes.FakeGitProviderHandler

	appSrv        AppService
	defaultParams AddParams
)

var _ = BeforeEach(func() {
	gitClient = &gitfakes.FakeGit{}
	fluxClient = &fluxfakes.FakeFlux{}
	kubeClient = &kubefakes.FakeKube{
		GetClusterNameStub: func() (string, error) {
			return "test-cluster", nil
		},
		GetClusterStatusStub: func() kube.ClusterStatus {
			return kube.WeGOInstalled
		},
	}
	gitProviders = &gitprovidersfakes.FakeGitProviderHandler{}

	appSrv = New(gitClient, fluxClient, kubeClient, gitProviders)

	defaultParams = AddParams{
		Url:            "https://github.com/foo/bar",
		Path:           "./kustomize",
		Branch:         "main",
		Dir:            ".",
		DeploymentType: "kustomize",
		Namespace:      "wego-system",
		AppConfigUrl:   "NONE",
	}
})

var _ = Describe("Add", func() {
	It("checks for cluster status", func() {
		err := appSrv.Add(defaultParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(kubeClient.GetClusterStatusCallCount()).To(Equal(1))

		kubeClient.GetClusterStatusStub = func() kube.ClusterStatus {
			return kube.Unmodified
		}
		err = appSrv.Add(defaultParams)
		Expect(err).To(MatchError("WeGO not installed... exiting"))

		kubeClient.GetClusterStatusStub = func() kube.ClusterStatus {
			return kube.Unknown
		}
		err = appSrv.Add(defaultParams)
		Expect(err).To(MatchError("WeGO can not determine cluster status... exiting"))
	})

	It("gets the cluster name", func() {
		err := appSrv.Add(defaultParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(kubeClient.GetClusterNameCallCount()).To(Equal(1))
	})

	It("creates and deploys a git secret", func() {
		fluxClient.CreateSecretGitStub = func(s1, s2, s3 string) ([]byte, error) {
			return []byte("deploy key"), nil
		}

		err := appSrv.Add(defaultParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(fluxClient.CreateSecretGitCallCount()).To(Equal(1))

		secretRef, repoUrl, namespace := fluxClient.CreateSecretGitArgsForCall(0)
		Expect(secretRef).To(Equal("weave-gitops-test-cluster"))
		Expect(repoUrl).To(Equal("ssh://git@github.com/foo/bar"))
		Expect(namespace).To(Equal("wego-system"))

		owner, repoName, deployKey := gitProviders.UploadDeployKeyArgsForCall(0)
		Expect(owner).To(Equal("foo"))
		Expect(repoName).To(Equal("bar"))
		Expect(deployKey).To(Equal([]byte("deploy key")))
	})

	Context("add app with no config repo", func() {
		Describe("generates source manifest", func() {
			It("creates GitRepository when source type is git", func() {
				defaultParams.SourceType = string(SourceTypeGit)

				err := appSrv.Add(defaultParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceGitCallCount()).To(Equal(1))

				name, url, branch, secretRef, namespace := fluxClient.CreateSourceGitArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(url).To(Equal("ssh://git@github.com/foo/bar"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("weave-gitops-test-cluster"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates HelmRepository when source type is helm", func() {
				defaultParams.Url = "https://charts.kube-ops.io"
				defaultParams.Chart = "loki"

				err := appSrv.Add(defaultParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceHelmCallCount()).To(Equal(1))

				name, url, namespace := fluxClient.CreateSourceHelmArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(url).To(Equal("https://charts.kube-ops.io"))
				Expect(namespace).To(Equal("wego-system"))
			})
		})

		Describe("generates application goat", func() {
			It("creates a kustomization if deployment type kustomize", func() {
				err := appSrv.Add(defaultParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateKustomizationCallCount()).To(Equal(1))

				name, source, path, namespace := fluxClient.CreateKustomizationArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("./kustomize"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates helm release using a helm repository if source type is helm", func() {
				defaultParams.Chart = "loki"

				err := appSrv.Add(defaultParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseHelmRepositoryCallCount()).To(Equal(1))

				name, chart, namespace := fluxClient.CreateHelmReleaseHelmRepositoryArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(chart).To(Equal("loki"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates a helm release using a git source if source type is git", func() {
				defaultParams.Path = "./charts/my-chart"
				defaultParams.DeploymentType = string(DeployTypeHelm)

				err := appSrv.Add(defaultParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseGitRepositoryCallCount()).To(Equal(1))

				name, source, path, namespace := fluxClient.CreateHelmReleaseGitRepositoryArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("./charts/my-chart"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("fails if deployment type is invalid", func() {
				defaultParams.DeploymentType = "foo"

				err := appSrv.Add(defaultParams)
				Expect(err).Should(HaveOccurred())
			})
		})

		It("applies the manifests to the cluster", func() {
			fluxClient.CreateSourceGitStub = func(s1, s2, s3, s4, s5 string) ([]byte, error) {
				return []byte("git source"), nil
			}
			fluxClient.CreateKustomizationStub = func(s1, s2, s3, s4 string) ([]byte, error) {
				return []byte("kustomization"), nil
			}

			err := appSrv.Add(defaultParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(kubeClient.ApplyCallCount()).To(Equal(3))

			sourceManifest, namespace := kubeClient.ApplyArgsForCall(0)
			Expect(sourceManifest).To(Equal([]byte("git source")))
			Expect(namespace).To(Equal("wego-system"))

			kustomizationManifest, namespace := kubeClient.ApplyArgsForCall(1)
			Expect(kustomizationManifest).To(Equal([]byte("kustomization")))
			Expect(namespace).To(Equal("wego-system"))

			appSpecManifest, namespace := kubeClient.ApplyArgsForCall(2)
			Expect(string(appSpecManifest)).To(ContainSubstring("kind: Application"))
			Expect(namespace).To(Equal("wego-system"))
		})
	})

	Context("add app with config in app repo", func() {
		BeforeEach(func() {
			defaultParams.Url = ""
			defaultParams.AppConfigUrl = ""

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

		Describe("generates source manifest", func() {
			It("creates GitRepository when source type is git", func() {
				defaultParams.SourceType = string(SourceTypeGit)

				err := appSrv.Add(defaultParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceGitCallCount()).To(Equal(1))

				name, url, branch, secretRef, namespace := fluxClient.CreateSourceGitArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(url).To(Equal("ssh://git@github.com/foo/bar"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("weave-gitops-test-cluster"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates HelmRepository when source type is helm", func() {
				defaultParams.Url = "https://charts.kube-ops.io"
				defaultParams.Chart = "loki"

				err := appSrv.Add(defaultParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceHelmCallCount()).To(Equal(1))

				name, url, namespace := fluxClient.CreateSourceHelmArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(url).To(Equal("https://charts.kube-ops.io"))
				Expect(namespace).To(Equal("wego-system"))
			})
		})

		Describe("generates application goat", func() {
			It("creates a kustomization if deployment type kustomize", func() {
				err := appSrv.Add(defaultParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateKustomizationCallCount()).To(Equal(3))

				name, source, path, namespace := fluxClient.CreateKustomizationArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("./kustomize"))
				Expect(namespace).To(Equal("wego-system"))

				name, source, path, namespace = fluxClient.CreateKustomizationArgsForCall(1)
				Expect(name).To(Equal("bar-wego-apps-dir"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal(".wego/apps/bar"))
				Expect(namespace).To(Equal("wego-system"))

				name, source, path, namespace = fluxClient.CreateKustomizationArgsForCall(2)
				Expect(name).To(Equal("test-cluster-bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal(".wego/targets/test-cluster"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates helm release using a helm repository if source type is helm", func() {
				defaultParams.Chart = "loki"

				err := appSrv.Add(defaultParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseHelmRepositoryCallCount()).To(Equal(1))

				name, chart, namespace := fluxClient.CreateHelmReleaseHelmRepositoryArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(chart).To(Equal("loki"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates a helm release using a git source if source type is git", func() {
				defaultParams.Path = "./charts/my-chart"
				defaultParams.DeploymentType = string(DeployTypeHelm)

				err := appSrv.Add(defaultParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseGitRepositoryCallCount()).To(Equal(1))

				name, source, path, namespace := fluxClient.CreateHelmReleaseGitRepositoryArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("./charts/my-chart"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("fails if deployment type is invalid", func() {
				defaultParams.DeploymentType = "foo"

				err := appSrv.Add(defaultParams)
				Expect(err).Should(HaveOccurred())
			})
		})

		It("applies the manifests to the cluster", func() {
			fluxClient.CreateSourceGitStub = func(s1, s2, s3, s4, s5 string) ([]byte, error) {
				return []byte("git source"), nil
			}
			fluxClient.CreateKustomizationStub = func(s1, s2, s3, s4 string) ([]byte, error) {
				return []byte("kustomization"), nil
			}

			err := appSrv.Add(defaultParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(kubeClient.ApplyCallCount()).To(Equal(4))

			sourceManifest, namespace := kubeClient.ApplyArgsForCall(0)
			Expect(sourceManifest).To(Equal([]byte("git source")))
			Expect(namespace).To(Equal("wego-system"))

			kustomizationManifest, namespace := kubeClient.ApplyArgsForCall(1)
			Expect(kustomizationManifest).To(Equal([]byte("kustomization")))
			Expect(namespace).To(Equal("wego-system"))

			appSpecManifest, namespace := kubeClient.ApplyArgsForCall(2)
			Expect(string(appSpecManifest)).To(ContainSubstring("kind: Application"))
			Expect(namespace).To(Equal("wego-system"))

			appWegoManifest, namespace := kubeClient.ApplyArgsForCall(3)
			Expect(string(appWegoManifest)).To(ContainSubstring("kustomization"))
			Expect(namespace).To(Equal("wego-system"))
		})

		Context("when using URL", func() {
			BeforeEach(func() {
				defaultParams.Url = "ssh://git@github.com/foo/bar.git"
			})

			It("clones the repo to a temp dir", func() {
				err := appSrv.Add(defaultParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(gitClient.CloneCallCount()).To(Equal(1))
				_, repoDir, url, branch := gitClient.CloneArgsForCall(0)

				Expect(repoDir).To(ContainSubstring("user-repo-"))
				Expect(url).To(Equal("ssh://git@github.com/foo/bar"))
				Expect(branch).To(Equal("main"))
			})
		})

		It("writes the files to the disk", func() {
			fluxClient.CreateSourceGitStub = func(s1, s2, s3, s4, s5 string) ([]byte, error) {
				return []byte("git"), nil
			}
			fluxClient.CreateKustomizationStub = func(s1, s2, s3, s4 string) ([]byte, error) {
				return []byte("kustomization"), nil
			}

			err := appSrv.Add(defaultParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(gitClient.WriteCallCount()).To(Equal(2))

			path, content := gitClient.WriteArgsForCall(0)
			Expect(path).To(Equal(".wego/apps/bar/app.yaml"))
			Expect(string(content)).To(ContainSubstring("kind: Application"))

			path, content = gitClient.WriteArgsForCall(1)
			Expect(path).To(Equal(".wego/targets/test-cluster/bar/bar-gitops-runtime.yaml"))
			Expect(content).To(Equal([]byte("gitkustomization")))
		})

		It("commit and pushes the files", func() {
			err := appSrv.Add(defaultParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(gitClient.CommitCallCount()).To(Equal(1))

			msg, filters := gitClient.CommitArgsForCall(0)
			Expect(msg).To(Equal(git.Commit{
				Author:  git.Author{Name: "Weave Gitops", Email: "weave-gitops@weave.works"},
				Message: "Add App manifests",
			}))

			Expect(filters[0](".wego/file.txt")).To(BeTrue())
		})
	})

	Context("add app with external config repo", func() {
		BeforeEach(func() {
			defaultParams.Url = "https://github.com/user/repo"
			defaultParams.AppConfigUrl = "https://github.com/foo/bar"
		})

		Describe("generates source manifest", func() {
			It("creates GitRepository when source type is git", func() {
				defaultParams.SourceType = string(SourceTypeGit)

				err := appSrv.Add(defaultParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceGitCallCount()).To(Equal(2))

				name, url, branch, secretRef, namespace := fluxClient.CreateSourceGitArgsForCall(0)
				Expect(name).To(Equal("repo"))
				Expect(url).To(Equal("ssh://git@github.com/user/repo"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("weave-gitops-test-cluster"))
				Expect(namespace).To(Equal("wego-system"))

				name, url, branch, secretRef, namespace = fluxClient.CreateSourceGitArgsForCall(1)
				Expect(name).To(Equal("bar"))
				Expect(url).To(Equal("ssh://git@github.com/foo/bar"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("weave-gitops-test-cluster"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates HelmRepository when source type is helm", func() {
				defaultParams.Url = "https://charts.kube-ops.io"
				defaultParams.Chart = "loki"

				err := appSrv.Add(defaultParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceHelmCallCount()).To(Equal(1))

				name, url, namespace := fluxClient.CreateSourceHelmArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(url).To(Equal("https://charts.kube-ops.io"))
				Expect(namespace).To(Equal("wego-system"))
			})
		})

		Describe("generates application goat", func() {
			It("creates a kustomization if deployment type kustomize", func() {
				err := appSrv.Add(defaultParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateKustomizationCallCount()).To(Equal(2))

				name, source, path, namespace := fluxClient.CreateKustomizationArgsForCall(0)
				Expect(name).To(Equal("repo"))
				Expect(source).To(Equal("repo"))
				Expect(path).To(Equal("./kustomize"))
				Expect(namespace).To(Equal("wego-system"))

				name, source, path, namespace = fluxClient.CreateKustomizationArgsForCall(1)
				Expect(name).To(Equal("weave-gitops-test-cluster"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("targets/test-cluster"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates helm release using a helm repository if source type is helm", func() {
				defaultParams.Chart = "loki"

				err := appSrv.Add(defaultParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseHelmRepositoryCallCount()).To(Equal(1))

				name, chart, namespace := fluxClient.CreateHelmReleaseHelmRepositoryArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(chart).To(Equal("loki"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates a helm release using a git source if source type is git", func() {
				defaultParams.Path = "./charts/my-chart"
				defaultParams.DeploymentType = string(DeployTypeHelm)

				err := appSrv.Add(defaultParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseGitRepositoryCallCount()).To(Equal(1))

				name, source, path, namespace := fluxClient.CreateHelmReleaseGitRepositoryArgsForCall(0)
				Expect(name).To(Equal("repo"))
				Expect(source).To(Equal("repo"))
				Expect(path).To(Equal("./charts/my-chart"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("fails if deployment type is invalid", func() {
				defaultParams.DeploymentType = "foo"

				err := appSrv.Add(defaultParams)
				Expect(err).Should(HaveOccurred())
			})
		})

		It("applies the manifests to the cluster", func() {
			fluxClient.CreateSourceGitStub = func(s1, s2, s3, s4, s5 string) ([]byte, error) {
				return []byte("git source"), nil
			}
			fluxClient.CreateKustomizationStub = func(s1, s2, s3, s4 string) ([]byte, error) {
				return []byte("kustomization"), nil
			}

			err := appSrv.Add(defaultParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(kubeClient.ApplyCallCount()).To(Equal(5))

			sourceManifest, namespace := kubeClient.ApplyArgsForCall(0)
			Expect(sourceManifest).To(Equal([]byte("git source")))
			Expect(namespace).To(Equal("wego-system"))

			kustomizationManifest, namespace := kubeClient.ApplyArgsForCall(1)
			Expect(kustomizationManifest).To(Equal([]byte("kustomization")))
			Expect(namespace).To(Equal("wego-system"))

			appSpecManifest, namespace := kubeClient.ApplyArgsForCall(2)
			Expect(string(appSpecManifest)).To(ContainSubstring("kind: Application"))
			Expect(namespace).To(Equal("wego-system"))
		})

		It("clones the repo to a temp dir", func() {
			err := appSrv.Add(defaultParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(gitClient.CloneCallCount()).To(Equal(1))
			_, repoDir, url, branch := gitClient.CloneArgsForCall(0)

			Expect(repoDir).To(ContainSubstring("user-repo-"))
			Expect(url).To(Equal("ssh://git@github.com/foo/bar"))
			Expect(branch).To(Equal("main"))
		})

		It("writes the files to the disk", func() {
			fluxClient.CreateSourceGitStub = func(s1, s2, s3, s4, s5 string) ([]byte, error) {
				return []byte("git"), nil
			}
			fluxClient.CreateKustomizationStub = func(s1, s2, s3, s4 string) ([]byte, error) {
				return []byte("kustomization"), nil
			}

			err := appSrv.Add(defaultParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(gitClient.WriteCallCount()).To(Equal(2))

			path, content := gitClient.WriteArgsForCall(0)
			Expect(path).To(Equal("apps/repo/app.yaml"))
			Expect(string(content)).To(ContainSubstring("kind: Application"))

			path, content = gitClient.WriteArgsForCall(1)
			Expect(path).To(Equal("targets/test-cluster/repo/repo-gitops-runtime.yaml"))
			Expect(content).To(Equal([]byte("gitkustomization")))
		})

		It("commit and pushes the files", func() {
			err := appSrv.Add(defaultParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(gitClient.CommitCallCount()).To(Equal(1))

			msg, filters := gitClient.CommitArgsForCall(0)
			Expect(msg).To(Equal(git.Commit{
				Author:  git.Author{Name: "Weave Gitops", Email: "weave-gitops@weave.works"},
				Message: "Add App manifests",
			}))

			Expect(len(filters)).To(Equal(0))
		})
	})

	Context("when using dry-run", func() {
		It("doesnt execute any action", func() {
			defaultParams.DryRun = true

			err := appSrv.Add(defaultParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(fluxClient.CreateSecretGitCallCount()).To(Equal(0))
			Expect(gitClient.CloneCallCount()).To(Equal(0))
			Expect(gitClient.WriteCallCount()).To(Equal(0))
			Expect(kubeClient.ApplyCallCount()).To(Equal(0))
		})
	})
})

var _ = Describe("sanitizeRepoUrl", func() {

	goodURL := "ssh://git@github.com/user/test-repo"

	It("should return proper ssh format for git@github.com:user/test-repo.git ", func() {

		url := "git@github.com:user/test-repo.git"

		sURL := sanitizeRepoUrl(url)
		Expect(goodURL).To(Equal(sURL))

	})
	It("should return proper ssh format for git@github.com:user/test-repo ", func() {

		url := "git@github.com:user/test-repo"

		sURL := sanitizeRepoUrl(url)
		Expect(goodURL).To(Equal(sURL))

	})
	It("should return proper ssh format for https://github.com/user/test-repo.git ", func() {

		url := "https://github.com/user/test-repo.git"

		sURL := sanitizeRepoUrl(url)
		Expect(goodURL).To(Equal(sURL))

	})
	It("should return proper ssh format for https://github.com/user/test-repo ", func() {

		url := "https://github.com/user/test-repo"

		sURL := sanitizeRepoUrl(url)
		Expect(goodURL).To(Equal(sURL))

	})
	It("should return proper ssh format for ssh://git@github.com/user/test-repo.git ", func() {

		url := "ssh://git@github.com/user/test-repo.git"

		sURL := sanitizeRepoUrl(url)
		Expect(goodURL).To(Equal(sURL))

	})
	It("should return proper ssh format for ssh://git@github.com/user/test-repo ", func() {

		url := "ssh://git@github.com/user/test-repo"

		sURL := sanitizeRepoUrl(url)
		Expect(goodURL).To(Equal(sURL))

	})
})
