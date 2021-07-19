package app

import (
	"context"
	"fmt"

	"github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"sigs.k8s.io/yaml"
)

var addParams AddParams

var _ = Describe("Add", func() {
	var _ = BeforeEach(func() {
		addParams = AddParams{
			Url:            "https://github.com/foo/bar",
			Path:           "./kustomize",
			Branch:         "main",
			Dir:            ".",
			DeploymentType: "kustomize",
			Namespace:      "wego-system",
			AppConfigUrl:   "NONE",
			AutoMerge:      true,
		}
	})

	It("checks for cluster status", func() {
		err := appSrv.Add(addParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(kubeClient.GetClusterStatusCallCount()).To(Equal(1))

		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.Unmodified
		}
		err = appSrv.Add(addParams)
		Expect(err).To(MatchError("Wego not installed... exiting"))

		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.Unknown
		}
		err = appSrv.Add(addParams)
		Expect(err).To(MatchError("Wego can not determine cluster status... exiting"))
	})

	It("gets the cluster name", func() {
		err := appSrv.Add(addParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(kubeClient.GetClusterNameCallCount()).To(Equal(1))
	})

	It("creates and deploys a git secret", func() {
		fluxClient.CreateSecretGitStub = func(s1, s2, s3 string) ([]byte, error) {
			return []byte("deploy key"), nil
		}

		err := appSrv.Add(addParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(fluxClient.CreateSecretGitCallCount()).To(Equal(1))

		secretRef, repoUrl, namespace := fluxClient.CreateSecretGitArgsForCall(0)
		Expect(secretRef).To(Equal("weave-gitops-test-cluster-bar"))
		Expect(repoUrl).To(Equal("ssh://git@github.com/foo/bar.git"))
		Expect(namespace).To(Equal("wego-system"))

		owner, repoName, deployKey := gitProviders.UploadDeployKeyArgsForCall(0)
		Expect(owner).To(Equal("foo"))
		Expect(repoName).To(Equal("bar"))
		Expect(deployKey).To(Equal([]byte("deploy key")))
	})

	Describe("checks for existing deploy key before creating secret", func() {
		It("looks up deploy key and skips creating secret if found", func() {
			addParams.SourceType = string(SourceTypeGit)

			gitProviders.DeployKeyExistsStub = func(s1, s2 string) (bool, error) {
				return true, nil
			}

			kubeClient.SecretPresentStub = func(ctx context.Context, s1, s2 string) (bool, error) {
				return true, nil
			}

			err := appSrv.Add(addParams)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(fluxClient.CreateSecretGitCallCount()).To(Equal(0))
			Expect(gitProviders.UploadDeployKeyCallCount()).To(Equal(0))
			Expect(kubeClient.SecretPresentCallCount()).To(Equal(1))
			Expect(gitProviders.DeployKeyExistsCallCount()).To(Equal(1))
		})

		It("looks up deploy key and creates secret if not found", func() {
			addParams.SourceType = string(SourceTypeGit)

			err := appSrv.Add(addParams)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(fluxClient.CreateSecretGitCallCount()).To(Equal(1))
			Expect(gitProviders.UploadDeployKeyCallCount()).To(Equal(1))
			Expect(kubeClient.SecretPresentCallCount()).To(Equal(1))
			Expect(gitProviders.DeployKeyExistsCallCount()).To(Equal(1))
		})
	})

	Context("add app with no config repo", func() {
		Describe("avoids deploy key for helm", func() {
			It("skips secret creation and lookup when source type is helm", func() {
				addParams.Url = "https://charts.kube-ops.io"
				addParams.Chart = "loki"

				err := appSrv.Add(addParams)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(fluxClient.CreateSecretGitCallCount()).To(Equal(0))
				Expect(gitProviders.UploadDeployKeyCallCount()).To(Equal(0))
				Expect(kubeClient.SecretPresentCallCount()).To(Equal(0))
				Expect(gitProviders.DeployKeyExistsCallCount()).To(Equal(0))
			})
		})

		Describe("generates source manifest", func() {
			It("creates GitRepository when source type is git", func() {
				addParams.SourceType = string(SourceTypeGit)

				err := appSrv.Add(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceGitCallCount()).To(Equal(1))

				name, url, branch, secretRef, namespace := fluxClient.CreateSourceGitArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(url).To(Equal("ssh://git@github.com/foo/bar.git"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("weave-gitops-test-cluster-bar"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates HelmRepository when source type is helm", func() {
				addParams.Url = "https://charts.kube-ops.io"
				addParams.Chart = "loki"

				err := appSrv.Add(addParams)
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
				err := appSrv.Add(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateKustomizationCallCount()).To(Equal(1))

				name, source, path, namespace := fluxClient.CreateKustomizationArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("./kustomize"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates helm release using a helm repository if source type is helm", func() {
				addParams.Chart = "loki"

				err := appSrv.Add(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseHelmRepositoryCallCount()).To(Equal(1))

				name, chart, namespace := fluxClient.CreateHelmReleaseHelmRepositoryArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(chart).To(Equal("loki"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates a helm release using a git source if source type is git", func() {
				addParams.Path = "./charts/my-chart"
				addParams.DeploymentType = string(DeployTypeHelm)

				err := appSrv.Add(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseGitRepositoryCallCount()).To(Equal(1))

				name, source, path, namespace := fluxClient.CreateHelmReleaseGitRepositoryArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("./charts/my-chart"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("fails if deployment type is invalid", func() {
				addParams.DeploymentType = "foo"

				err := appSrv.Add(addParams)
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

			err := appSrv.Add(addParams)
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
			addParams.Url = ""
			addParams.AppConfigUrl = ""

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

		Describe("avoids deploy key for helm", func() {
			It("skips secret creation and lookup when source type is helm", func() {
				addParams.Url = "https://charts.kube-ops.io"
				addParams.Chart = "loki"

				err := appSrv.Add(addParams)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(fluxClient.CreateSecretGitCallCount()).To(Equal(0))
				Expect(gitProviders.UploadDeployKeyCallCount()).To(Equal(0))
				Expect(gitProviders.DeployKeyExistsCallCount()).To(Equal(0))
				Expect(kubeClient.SecretPresentCallCount()).To(Equal(0))
			})
		})

		Describe("generates source manifest", func() {
			It("creates GitRepository when source type is git", func() {
				addParams.SourceType = string(SourceTypeGit)

				err := appSrv.Add(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceGitCallCount()).To(Equal(1))

				name, url, branch, secretRef, namespace := fluxClient.CreateSourceGitArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(url).To(Equal("ssh://git@github.com/foo/bar.git"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("weave-gitops-test-cluster-bar"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates HelmRepository when source type is helm", func() {
				addParams.Url = "https://charts.kube-ops.io"
				addParams.Chart = "loki"

				err := appSrv.Add(addParams)
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
				err := appSrv.Add(addParams)
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
				addParams.Chart = "loki"

				err := appSrv.Add(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseHelmRepositoryCallCount()).To(Equal(1))

				name, chart, namespace := fluxClient.CreateHelmReleaseHelmRepositoryArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(chart).To(Equal("loki"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates a helm release using a git source if source type is git", func() {
				addParams.Path = "./charts/my-chart"
				addParams.DeploymentType = string(DeployTypeHelm)

				err := appSrv.Add(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseGitRepositoryCallCount()).To(Equal(1))

				name, source, path, namespace := fluxClient.CreateHelmReleaseGitRepositoryArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("./charts/my-chart"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("fails if deployment type is invalid", func() {
				addParams.DeploymentType = "foo"

				err := appSrv.Add(addParams)
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

			err := appSrv.Add(addParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(kubeClient.ApplyCallCount()).To(Equal(2))

			sourceManifest, namespace := kubeClient.ApplyArgsForCall(0)
			Expect(sourceManifest).To(Equal([]byte("git source")))
			Expect(namespace).To(Equal("wego-system"))

			appWegoManifest, namespace := kubeClient.ApplyArgsForCall(1)
			Expect(appWegoManifest).To(Equal([]byte("kustomizationkustomization")))
			Expect(namespace).To(Equal("wego-system"))
		})

		Context("when using URL", func() {
			BeforeEach(func() {
				addParams.Url = "ssh://git@github.com/foo/bar.git"
			})

			It("clones the repo to a temp dir", func() {
				err := appSrv.Add(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(gitClient.CloneCallCount()).To(Equal(1))
				_, repoDir, url, branch := gitClient.CloneArgsForCall(0)

				Expect(repoDir).To(ContainSubstring("user-repo-"))
				Expect(url).To(Equal("ssh://git@github.com/foo/bar.git"))
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

			err := appSrv.Add(addParams)
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
			err := appSrv.Add(addParams)
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
			addParams.Url = "https://github.com/user/repo"
			addParams.AppConfigUrl = "https://github.com/foo/bar"
		})

		Describe("generates source manifest", func() {
			It("creates GitRepository when source type is git", func() {
				addParams.SourceType = string(SourceTypeGit)

				err := appSrv.Add(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceGitCallCount()).To(Equal(2))

				name, url, branch, secretRef, namespace := fluxClient.CreateSourceGitArgsForCall(0)
				Expect(name).To(Equal("repo"))
				Expect(url).To(Equal("ssh://git@github.com/user/repo.git"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("weave-gitops-test-cluster-repo"))
				Expect(namespace).To(Equal("wego-system"))

				name, url, branch, secretRef, namespace = fluxClient.CreateSourceGitArgsForCall(1)
				Expect(name).To(Equal("bar"))
				Expect(url).To(Equal("ssh://git@github.com/foo/bar.git"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("weave-gitops-test-cluster-bar"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates HelmRepository when source type is helm", func() {
				addParams.Url = "https://charts.kube-ops.io"
				addParams.Chart = "loki"

				err := appSrv.Add(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceHelmCallCount()).To(Equal(1))

				name, url, namespace := fluxClient.CreateSourceHelmArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(url).To(Equal("https://charts.kube-ops.io"))
				Expect(namespace).To(Equal("wego-system"))
			})
		})

		Describe("generateAppYaml", func() {
			It("generates the app yaml", func() {
				repoURL := "ssh://git@github.com/example/my-source"
				params := AddParams{
					Name:      "my-app",
					Namespace: "wego-system",
					Url:       repoURL,
					Path:      "manifests",
					Branch:    "main",
				}

				desired2 := makeWegoApplication(params)
				hash, err := getHash(repoURL, params.Path, params.Branch)
				Expect(err).To(BeNil())

				desired2.ObjectMeta.Labels = map[string]string{WeGOAppIdentifierLabelKey: hash}

				Expect(err).NotTo(HaveOccurred())
				out, err := generateAppYaml(params, hash)
				Expect(err).To(BeNil())

				result := wego.Application{}
				// Convert back to a struct to make the comparison more forgiving.
				// A straight string/byte comparison doesn't account for un-ordered keys in yaml.
				Expect(yaml.Unmarshal(out, &result))

				diff := cmp.Diff(result, desired2)
				Expect(diff).To(Equal(""))

				// Not entirely sure how to get gomega to pretty-print the output of `diff`,
				// so we assert it should be empty above, and conditionally print the diff to make a nice assertion message.
				// `diff` is a formatted string
				if diff != "" {
					fmt.Println(diff)
				}
			})
		})

		Describe("generates application goat", func() {
			It("creates a kustomization if deployment type kustomize", func() {
				err := appSrv.Add(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateKustomizationCallCount()).To(Equal(3))

				name, source, path, namespace := fluxClient.CreateKustomizationArgsForCall(0)
				Expect(name).To(Equal("repo"))
				Expect(source).To(Equal("repo"))
				Expect(path).To(Equal("./kustomize"))
				Expect(namespace).To(Equal("wego-system"))

				name, source, path, namespace = fluxClient.CreateKustomizationArgsForCall(1)
				Expect(name).To(Equal("repo"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("apps/repo"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates helm release using a helm repository if source type is helm", func() {
				addParams.Chart = "loki"

				err := appSrv.Add(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseHelmRepositoryCallCount()).To(Equal(1))

				name, chart, namespace := fluxClient.CreateHelmReleaseHelmRepositoryArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(chart).To(Equal("loki"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates a helm release using a git source if source type is git", func() {
				addParams.Path = "./charts/my-chart"
				addParams.DeploymentType = string(DeployTypeHelm)

				err := appSrv.Add(addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseGitRepositoryCallCount()).To(Equal(1))

				name, source, path, namespace := fluxClient.CreateHelmReleaseGitRepositoryArgsForCall(0)
				Expect(name).To(Equal("repo"))
				Expect(source).To(Equal("repo"))
				Expect(path).To(Equal("./charts/my-chart"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("fails if deployment type is invalid", func() {
				addParams.DeploymentType = "foo"

				err := appSrv.Add(addParams)
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

			err := appSrv.Add(addParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(kubeClient.ApplyCallCount()).To(Equal(2))

			sourceManifest, namespace := kubeClient.ApplyArgsForCall(0)
			Expect(sourceManifest).To(Equal([]byte("git source")))
			Expect(namespace).To(Equal("wego-system"))

			kustomizationManifest, namespace := kubeClient.ApplyArgsForCall(1)
			Expect(kustomizationManifest).To(Equal([]byte("kustomizationkustomization")))
			Expect(namespace).To(Equal("wego-system"))
		})

		It("clones the repo to a temp dir", func() {
			err := appSrv.Add(addParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(gitClient.CloneCallCount()).To(Equal(1))
			_, repoDir, url, branch := gitClient.CloneArgsForCall(0)

			Expect(repoDir).To(ContainSubstring("user-repo-"))
			Expect(url).To(Equal("ssh://git@github.com/foo/bar.git"))
			Expect(branch).To(Equal("main"))
		})

		It("writes the files to the disk", func() {
			fluxClient.CreateSourceGitStub = func(s1, s2, s3, s4, s5 string) ([]byte, error) {
				return []byte("git"), nil
			}
			fluxClient.CreateKustomizationStub = func(s1, s2, s3, s4 string) ([]byte, error) {
				return []byte("kustomization"), nil
			}

			err := appSrv.Add(addParams)
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
			err := appSrv.Add(addParams)
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
			addParams.DryRun = true
			addParams.AutoMerge = true

			err := appSrv.Add(addParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(fluxClient.CreateSecretGitCallCount()).To(Equal(0))
			Expect(gitClient.CloneCallCount()).To(Equal(0))
			Expect(gitClient.WriteCallCount()).To(Equal(0))
			Expect(kubeClient.ApplyCallCount()).To(Equal(0))
		})
	})

	Describe("Test app hash", func() {
		It("should return right hash for a helm app", func() {

			addParams.Url = "https://github.com/owner/repo1"
			addParams.Chart = "nginx"
			addParams.Branch = "main"
			addParams.DeploymentType = string(DeployTypeHelm)

			appHash, err := getAppHash(addParams)
			Expect(err).NotTo(HaveOccurred())

			expectedHash, err := getHash(addParams.Url, addParams.Chart, addParams.Branch)
			Expect(err).NotTo(HaveOccurred())

			Expect(appHash).To(Equal(expectedHash))

		})

		It("should return right hash for a kustomize app", func() {

			addParams.Url = "https://github.com/owner/repo1"
			addParams.Path = "custompath"
			addParams.Branch = "main"
			addParams.DeploymentType = string(DeployTypeKustomize)

			appHash, err := getAppHash(addParams)
			Expect(err).NotTo(HaveOccurred())

			expectedHash, err := getHash(addParams.Url, addParams.Path, addParams.Branch)
			Expect(err).NotTo(HaveOccurred())

			Expect(appHash).To(Equal(expectedHash))

		})
	})
})
