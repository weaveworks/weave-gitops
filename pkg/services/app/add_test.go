package app

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5/memfs"
	gogit "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/storage/memory"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"sigs.k8s.io/yaml"
)

var (
	addParams AddParams
	ctx       context.Context
)

var _ = Describe("Add", func() {
	var _ = BeforeEach(func() {
		addParams = AddParams{
			Url:            "https://github.com/foo/bar",
			Path:           "./kustomize",
			Branch:         "main",
			Dir:            ".",
			DeploymentType: "kustomize",
			SourceType:     wego.SourceTypeGit,
			Namespace:      wego.DefaultNamespace,
			AppConfigUrl:   "NONE",
			AutoMerge:      true,
		}

		gitProviders.GetDefaultBranchReturns("main", nil)

		ctx = context.Background()
	})

	It("checks for cluster status", func() {
		err := appSrv.Add(gitClient, gitProviders, addParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(kubeClient.GetClusterStatusCallCount()).To(Equal(1))

		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.Unmodified
		}
		err = appSrv.Add(gitClient, gitProviders, addParams)
		Expect(err).To(MatchError("gitops not installed... exiting"))

		kubeClient.GetClusterStatusStub = func(ctx context.Context) kube.ClusterStatus {
			return kube.Unknown
		}
		err = appSrv.Add(gitClient, gitProviders, addParams)
		Expect(err).To(MatchError("can not determine cluster status... exiting"))
	})

	It("gets the cluster name", func() {
		err := appSrv.Add(gitClient, gitProviders, addParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(kubeClient.GetClusterNameCallCount()).To(Equal(1))
	})

	It("validates app-config-url is set when source is helm", func() {
		addParams.Chart = "my-chart"
		addParams.Url = "https://my-chart.com"
		addParams.AppConfigUrl = ""

		err := appSrv.Add(gitClient, gitProviders, addParams)
		Expect(err.Error()).Should(HaveSuffix("--app-config-url should be provided or set to NONE"))
	})

	Context("Looking up repo default branch", func() {
		var _ = BeforeEach(func() {
			gitProviders.GetDefaultBranchStub = func(_ context.Context, repoUrl gitproviders.RepoURL) (string, error) {
				branch := "an-unusual-branch" // for app repository
				if !strings.Contains(repoUrl.String(), "bar") {
					branch = "config-branch" // for config repository
				}
				return branch, nil
			}

			addParams.Branch = ""
		})

		It("Uses the default branch from the repository if no branch is specified", func() {
			updated, err := appSrv.(*App).updateParametersIfNecessary(ctx, gitProviders, addParams)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(updated.Branch).To(Equal("an-unusual-branch"))
		})

		It("Allows a specified branch to override the repo's default branch", func() {
			addParams.Branch = "an-overriding-branch"
			updated, err := appSrv.(*App).updateParametersIfNecessary(ctx, gitProviders, addParams)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(updated.Branch).To(Equal("an-overriding-branch"))
		})
	})

	Context("add app with no config repo", func() {
		Describe("generates source manifest", func() {
			It("creates GitRepository when source type is git", func() {
				addParams.SourceType = wego.SourceTypeGit

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceGitCallCount()).To(Equal(1))

				name, url, branch, secretRef, namespace := fluxClient.CreateSourceGitArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(url).To(Equal("ssh://git@github.com/foo/bar.git"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("wego-test-cluster-bar"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
			})

			It("creates HelmRepository when source type is helm", func() {
				addParams.Url = "https://charts.kube-ops.io"
				addParams.Chart = "loki"

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(fluxClient.CreateSourceHelmCallCount()).To(Equal(1))

				name, url, namespace := fluxClient.CreateSourceHelmArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(url).To(Equal("https://charts.kube-ops.io"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
			})
		})

		Describe("generates application goat", func() {
			It("creates a kustomization if deployment type kustomize", func() {
				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateKustomizationCallCount()).To(Equal(1))

				name, source, path, namespace := fluxClient.CreateKustomizationArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("./kustomize"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
			})

			It("creates helm release using a helm repository if source type is helm", func() {
				addParams.Chart = "loki"

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseHelmRepositoryCallCount()).To(Equal(1))

				name, chart, namespace, targetNamespace := fluxClient.CreateHelmReleaseHelmRepositoryArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(chart).To(Equal("loki"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
				Expect(targetNamespace).To(Equal(""))
			})

			It("creates a helm release using a git source if source type is git", func() {
				addParams.Path = "./charts/my-chart"
				addParams.DeploymentType = string(wego.DeploymentTypeHelm)

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseGitRepositoryCallCount()).To(Equal(1))

				name, source, path, namespace, targetNamespace := fluxClient.CreateHelmReleaseGitRepositoryArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("./charts/my-chart"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
				Expect(targetNamespace).To(Equal(""))
			})

			It("creates helm release for helm repository with target namespace if source type is helm", func() {
				addParams.Chart = "loki"
				addParams.HelmReleaseTargetNamespace = "sock-shop"

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseHelmRepositoryCallCount()).To(Equal(1))

				name, chart, namespace, targetNamespace := fluxClient.CreateHelmReleaseHelmRepositoryArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(chart).To(Equal("loki"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
				Expect(targetNamespace).To(Equal("sock-shop"))
			})

			It("creates a helm release for git repository with target namespace if source type is git", func() {
				addParams.Path = "./charts/my-chart"
				addParams.DeploymentType = string(wego.DeploymentTypeHelm)
				addParams.HelmReleaseTargetNamespace = "sock-shop"

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseGitRepositoryCallCount()).To(Equal(1))

				name, source, path, namespace, targetNamespace := fluxClient.CreateHelmReleaseGitRepositoryArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("./charts/my-chart"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
				Expect(targetNamespace).To(Equal("sock-shop"))
			})

			It("fails if deployment type is invalid", func() {
				addParams.DeploymentType = "foo"

				err := appSrv.Add(gitClient, gitProviders, addParams)
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

			err := appSrv.Add(gitClient, gitProviders, addParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(kubeClient.ApplyCallCount()).To(Equal(3))

			_, sourceManifest, namespace := kubeClient.ApplyArgsForCall(0)
			Expect(sourceManifest).To(Equal([]byte("git source")))
			Expect(namespace).To(Equal(wego.DefaultNamespace))

			_, kustomizationManifest, namespace := kubeClient.ApplyArgsForCall(1)
			Expect(kustomizationManifest).To(Equal([]byte("kustomization")))
			Expect(namespace).To(Equal(wego.DefaultNamespace))

			_, appSpecManifest, namespace := kubeClient.ApplyArgsForCall(2)
			Expect(string(appSpecManifest)).To(ContainSubstring("kind: Application"))
			Expect(namespace).To(Equal(wego.DefaultNamespace))
		})
	})

	Context("add app with config in app repo", func() {
		BeforeEach(func() {
			addParams.Url = "ssh://git@github.com/foo/bar.git"
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

		Describe("generates source manifest", func() {
			It("creates GitRepository when source type is git", func() {
				addParams.SourceType = wego.SourceTypeGit
				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceGitCallCount()).To(Equal(1))

				name, url, branch, secretRef, namespace := fluxClient.CreateSourceGitArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(url).To(Equal("ssh://git@github.com/foo/bar.git"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("wego-test-cluster-bar"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
			})

			It("creates HelmRepository when source type is helm", func() {
				addParams.Url = "https://charts.kube-ops.io"
				addParams.Chart = "loki"
				addParams.AppConfigUrl = "ssh://git@github.com/owner/config-repo.git"

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceHelmCallCount()).To(Equal(1))

				name, url, namespace := fluxClient.CreateSourceHelmArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(url).To(Equal("https://charts.kube-ops.io"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
			})
		})

		Describe("generates application goat", func() {
			It("creates a kustomization if deployment type kustomize", func() {
				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateKustomizationCallCount()).To(Equal(3))

				name, source, path, namespace := fluxClient.CreateKustomizationArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("./kustomize"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))

				name, source, path, namespace = fluxClient.CreateKustomizationArgsForCall(1)
				Expect(name).To(Equal("bar-apps-dir"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal(".wego/apps/bar"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))

				name, source, path, namespace = fluxClient.CreateKustomizationArgsForCall(2)
				Expect(name).To(Equal("test-cluster-bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal(".wego/targets/test-cluster/bar"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
			})

			It("creates helm release using a helm repository if source type is helm", func() {
				addParams.Url = "https://charts.kube-ops.io"
				addParams.Chart = "loki"
				addParams.AppConfigUrl = "ssh://github.com/owner/repo"

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseHelmRepositoryCallCount()).To(Equal(1))

				name, chart, namespace, targetNamespace := fluxClient.CreateHelmReleaseHelmRepositoryArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(chart).To(Equal("loki"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
				Expect(targetNamespace).To(Equal(""))
			})

			It("creates a helm release using a git source if source type is git", func() {
				addParams.Path = "./charts/my-chart"
				addParams.DeploymentType = string(wego.DeploymentTypeHelm)

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseGitRepositoryCallCount()).To(Equal(1))

				name, source, path, namespace, targetNamespace := fluxClient.CreateHelmReleaseGitRepositoryArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("./charts/my-chart"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
				Expect(targetNamespace).To(Equal(""))
			})

			It("creates helm release for helm repository with target namespace if source type is helm", func() {
				addParams.Url = "https://charts.kube-ops.io"
				addParams.Chart = "loki"
				addParams.HelmReleaseTargetNamespace = "sock-shop"
				addParams.AppConfigUrl = "ssh://git@github.com/owner/config-repo.git"

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseHelmRepositoryCallCount()).To(Equal(1))

				name, chart, namespace, targetNamespace := fluxClient.CreateHelmReleaseHelmRepositoryArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(chart).To(Equal("loki"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
				Expect(targetNamespace).To(Equal("sock-shop"))
			})

			It("creates a helm release for git repository with target namespace if source type is git", func() {
				addParams.Path = "./charts/my-chart"
				addParams.DeploymentType = string(wego.DeploymentTypeHelm)
				addParams.HelmReleaseTargetNamespace = "sock-shop"

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseGitRepositoryCallCount()).To(Equal(1))

				name, source, path, namespace, targetNamespace := fluxClient.CreateHelmReleaseGitRepositoryArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("./charts/my-chart"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
				Expect(targetNamespace).To(Equal("sock-shop"))
			})

			It("validates namespace passed as target namespace", func() {
				addParams.Url = "https://charts.kube-ops.io"
				addParams.Chart = "loki"
				addParams.HelmReleaseTargetNamespace = "sock-shop"
				addParams.AppConfigUrl = "ssh://git@github.com/owner/config-repo.git"

				goodNamespaceErr := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(goodNamespaceErr).ShouldNot(HaveOccurred())

				addParams.HelmReleaseTargetNamespace = "sock-shop&*&*&*&"

				badNamespaceErr := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(badNamespaceErr.Error()).To(HavePrefix("could not update parameters: invalid namespace"))
			})

			It("validates target namespace exists", func() {
				addParams.Url = "https://charts.kube-ops.io"
				addParams.Chart = "loki"
				addParams.HelmReleaseTargetNamespace = "sock-shop"
				addParams.AppConfigUrl = "NONE"

				goodNamespaceErr := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(goodNamespaceErr).ShouldNot(HaveOccurred())

				kubeClient.NamespacePresentReturns(false, nil)

				badNamespaceErr := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(badNamespaceErr.Error()).To(HavePrefix("could not update parameters: Helm Release Target Namespace sock-shop does not exist"))
			})

			It("fails if deployment type is invalid", func() {
				addParams.DeploymentType = "foo"

				err := appSrv.Add(gitClient, gitProviders, addParams)
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

			err := appSrv.Add(gitClient, gitProviders, addParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(kubeClient.ApplyCallCount()).To(Equal(3))

			_, sourceManifest, namespace := kubeClient.ApplyArgsForCall(0)
			Expect(sourceManifest).To(Equal([]byte("git source")))
			Expect(namespace).To(Equal(wego.DefaultNamespace))

			_, appWegoManifest, namespace := kubeClient.ApplyArgsForCall(1)
			Expect(appWegoManifest).To(Equal([]byte("kustomization")))
			Expect(namespace).To(Equal(wego.DefaultNamespace))
		})

		Context("when using URL", func() {
			BeforeEach(func() {
				addParams.Url = "ssh://git@github.com/foo/bar.git"
			})

			It("clones the repo to a temp dir", func() {
				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(gitClient.CloneCallCount()).To(Equal(1))
				_, repoDir, url, branch := gitClient.CloneArgsForCall(0)

				Expect(repoDir).To(ContainSubstring("user-repo-"))
				Expect(url).To(Equal("ssh://git@github.com/foo/bar.git"))
				Expect(branch).To(Equal("main"))
			})

			It("writes the files to the disk", func() {
				addParams.AppConfigUrl = addParams.Url // so we know the root is ".wego"
				fluxClient.CreateSourceGitStub = func(s1, s2, s3, s4, s5 string) ([]byte, error) {
					return []byte("git"), nil
				}
				fluxClient.CreateKustomizationStub = func(s1, s2, s3, s4 string) ([]byte, error) {
					return []byte("kustomization"), nil
				}

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(gitClient.WriteCallCount()).To(Equal(3))

				path, content := gitClient.WriteArgsForCall(0)
				Expect(path).To(Equal(".wego/apps/bar/app.yaml"))
				Expect(string(content)).To(ContainSubstring("kind: Application"))

				path, content = gitClient.WriteArgsForCall(1)
				Expect(path).To(Equal(".wego/targets/test-cluster/bar/bar-gitops-source.yaml"))
				Expect(content).To(Equal([]byte("git")))

				path, content = gitClient.WriteArgsForCall(2)
				Expect(path).To(Equal(".wego/targets/test-cluster/bar/bar-gitops-deploy.yaml"))
				Expect(content).To(Equal([]byte("kustomization")))
			})
		})

		It("commit and pushes the files", func() {
			err := appSrv.Add(gitClient, gitProviders, addParams)
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
				addParams.SourceType = wego.SourceTypeGit

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceGitCallCount()).To(Equal(2))

				name, url, branch, secretRef, namespace := fluxClient.CreateSourceGitArgsForCall(0)
				Expect(name).To(Equal("repo"))
				Expect(url).To(Equal("ssh://git@github.com/user/repo.git"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("wego-test-cluster-repo"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))

				name, url, branch, secretRef, namespace = fluxClient.CreateSourceGitArgsForCall(1)
				Expect(name).To(Equal("bar"))
				Expect(url).To(Equal("ssh://git@github.com/foo/bar.git"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("wego-test-cluster-bar"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
			})

			It("creates HelmRepository when source type is helm", func() {
				addParams.Url = "https://charts.kube-ops.io"
				addParams.Chart = "loki"

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceHelmCallCount()).To(Equal(1))

				name, url, namespace := fluxClient.CreateSourceHelmArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(url).To(Equal("https://charts.kube-ops.io"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
			})
		})

		Describe("generateAppYaml", func() {
			It("generates the app yaml", func() {
				repoURL := "ssh://git@github.com/example/my-source"
				params := AddParams{
					Name:         "my-app",
					Namespace:    wego.DefaultNamespace,
					Url:          repoURL,
					AppConfigUrl: "none",
					Path:         "manifests",
					Branch:       "main",
				}

				info, err := getAppResourceInfo(makeWegoApplication(params), "")
				Expect(err).ToNot(HaveOccurred())

				desired2 := info.Application
				hash := getHash(repoURL, info.Spec.Path, info.Spec.Branch)

				desired2.ObjectMeta.Labels = map[string]string{WeGOAppIdentifierLabelKey: hash}

				out, err := generateAppYaml(info, hash)
				Expect(err).To(BeNil())

				result := wego.Application{}
				// Convert back to a struct to make the comparison more forgiving.
				// A straight string/byte comparison doesn't account for un-ordered keys in yaml.
				Expect(yaml.Unmarshal(out.ToAppYAML(), &result))

				diff := cmp.Diff(result, desired2)
				Expect(diff).To(Equal(""))

				// Not entirely sure how to get gomega to pretty-print the output of `diff`,
				// so we assert it should be empty above, and conditionally print the diff to make a nice assertion message.
				// `diff` is a formatted string
				if diff != "" {
					appSrv.(*App).Logger.Println(diff)
				}
			})
		})

		Describe("generates application goat", func() {
			It("creates a kustomization if deployment type kustomize", func() {
				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateKustomizationCallCount()).To(Equal(3))

				name, source, path, namespace := fluxClient.CreateKustomizationArgsForCall(0)
				Expect(name).To(Equal("repo"))
				Expect(source).To(Equal("repo"))
				Expect(path).To(Equal("./kustomize"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))

				name, source, path, namespace = fluxClient.CreateKustomizationArgsForCall(1)
				Expect(name).To(Equal("repo-apps-dir"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("apps/repo"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
			})

			It("creates helm release using a helm repository if source type is helm", func() {
				addParams.Chart = "loki"

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseHelmRepositoryCallCount()).To(Equal(1))

				name, chart, namespace, targetNamespace := fluxClient.CreateHelmReleaseHelmRepositoryArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(chart).To(Equal("loki"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
				Expect(targetNamespace).To(Equal(""))
			})

			It("creates a helm release using a git source if source type is git", func() {
				addParams.Path = "./charts/my-chart"
				addParams.DeploymentType = string(wego.DeploymentTypeHelm)

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseGitRepositoryCallCount()).To(Equal(1))

				name, source, path, namespace, targetNamespace := fluxClient.CreateHelmReleaseGitRepositoryArgsForCall(0)
				Expect(name).To(Equal("repo"))
				Expect(source).To(Equal("repo"))
				Expect(path).To(Equal("./charts/my-chart"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
				Expect(targetNamespace).To(Equal(""))
			})

			It("creates helm release for helm repository with target namespace if source type is helm", func() {
				addParams.Chart = "loki"
				addParams.HelmReleaseTargetNamespace = "sock-shop"

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseHelmRepositoryCallCount()).To(Equal(1))

				name, chart, namespace, targetNamespace := fluxClient.CreateHelmReleaseHelmRepositoryArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(chart).To(Equal("loki"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
				Expect(targetNamespace).To(Equal("sock-shop"))
			})

			It("creates a helm release for git repository with target namespace if source type is git", func() {
				addParams.Path = "./charts/my-chart"
				addParams.DeploymentType = string(wego.DeploymentTypeHelm)
				addParams.HelmReleaseTargetNamespace = "sock-shop"

				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseGitRepositoryCallCount()).To(Equal(1))

				name, source, path, namespace, targetNamespace := fluxClient.CreateHelmReleaseGitRepositoryArgsForCall(0)
				Expect(name).To(Equal("repo"))
				Expect(source).To(Equal("repo"))
				Expect(path).To(Equal("./charts/my-chart"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
				Expect(targetNamespace).To(Equal("sock-shop"))
			})

			It("fails if deployment type is invalid", func() {
				addParams.DeploymentType = "foo"

				err := appSrv.Add(gitClient, gitProviders, addParams)
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

			err := appSrv.Add(gitClient, gitProviders, addParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(kubeClient.ApplyCallCount()).To(Equal(3))

			_, sourceManifest, namespace := kubeClient.ApplyArgsForCall(0)
			Expect(sourceManifest).To(Equal([]byte("git source")))
			Expect(namespace).To(Equal(wego.DefaultNamespace))

			_, kustomizationManifest, namespace := kubeClient.ApplyArgsForCall(1)
			Expect(kustomizationManifest).To(Equal([]byte("kustomization")))
			Expect(namespace).To(Equal(wego.DefaultNamespace))
		})

		It("clones the repo to a temp dir", func() {
			err := appSrv.Add(gitClient, gitProviders, addParams)
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

			err := appSrv.Add(gitClient, gitProviders, addParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(gitClient.WriteCallCount()).To(Equal(3))

			path, content := gitClient.WriteArgsForCall(0)
			Expect(path).To(Equal("apps/repo/app.yaml"))
			Expect(string(content)).To(ContainSubstring("kind: Application"))

			path, content = gitClient.WriteArgsForCall(1)
			Expect(path).To(Equal("targets/test-cluster/repo/repo-gitops-source.yaml"))
			Expect(content).To(Equal([]byte("git")))

			path, content = gitClient.WriteArgsForCall(2)
			Expect(path).To(Equal("targets/test-cluster/repo/repo-gitops-deploy.yaml"))
			Expect(content).To(Equal([]byte("kustomization")))
		})

		It("commit and pushes the files", func() {
			err := appSrv.Add(gitClient, gitProviders, addParams)
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

	Context("when creating a pull request", func() {
		var info *AppResourceInfo

		BeforeEach(func() {
			gitProviders.GetDefaultBranchStub = func(_ context.Context, repoUrl gitproviders.RepoURL) (string, error) {
				addUrl, err := gitproviders.NewRepoURL(addParams.Url)
				Expect(err).NotTo(HaveOccurred())

				if repoUrl.String() == addUrl.String() {
					return "default-app-branch", nil
				}
				return "default-config-branch", nil
			}

			gitProviders.CreatePullRequestReturns(testutils.DummyPullRequest{}, nil)

			addParams.Url = "ssh://github.com/user/repo.git"
		})

		JustBeforeEach(func() {
			var err error
			info, err = getAppResourceInfo(makeWegoApplication(addParams), "cluster")
			Expect(err).ToNot(HaveOccurred())
		})

		Context("uses the default app branch for config in app repository", func() {
			BeforeEach(func() {
				addParams.AppConfigUrl = ""
			})

			It("creates the pull request against the default branch for an org app repository", func() {
				Expect(appSrv.(*App).createPullRequestToRepo(ctx, gitProviders, info, info.appRepoUrl, "hash", emptyAppManifest(), emptySource(), emptyAutomation())).To(Succeed())
				_, _, prInfo := gitProviders.CreatePullRequestArgsForCall(0)
				Expect(prInfo.TargetBranch).To(Equal("default-app-branch"))
			})

			It("creates the pull request against the default branch for a user app repository", func() {
				Expect(appSrv.(*App).createPullRequestToRepo(ctx, gitProviders, info, info.appRepoUrl, "hash", emptyAppManifest(), emptySource(), emptyAutomation())).To(Succeed())
				_, _, prInfo := gitProviders.CreatePullRequestArgsForCall(0)
				Expect(prInfo.TargetBranch).To(Equal("default-app-branch"))
			})
		})

		Context("uses the default config branch for external config", func() {
			BeforeEach(func() {
				addParams.AppConfigUrl = "https://github.com/foo/bar"
			})

			It("creates the pull request against the default branch for an org config repository", func() {
				Expect(appSrv.(*App).createPullRequestToRepo(ctx, gitProviders, info, info.configRepoUrl, "hash", emptyAppManifest(), emptySource(), emptyAutomation())).To(Succeed())
				_, _, prInfo := gitProviders.CreatePullRequestArgsForCall(0)
				Expect(prInfo.TargetBranch).To(Equal("default-config-branch"))
			})

			It("creates the pull request against the default branch for a user config repository", func() {
				Expect(appSrv.(*App).createPullRequestToRepo(ctx, gitProviders, info, info.configRepoUrl, "hash", emptyAppManifest(), emptySource(), emptyAutomation())).To(Succeed())
				_, _, prInfo := gitProviders.CreatePullRequestArgsForCall(0)
				Expect(prInfo.TargetBranch).To(Equal("default-config-branch"))
			})
		})
	})

	Context("when using dry-run", func() {
		It("doesnt execute any action", func() {
			addParams.DryRun = true
			addParams.AutoMerge = true

			err := appSrv.Add(gitClient, gitProviders, addParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(fluxClient.CreateSecretGitCallCount()).To(Equal(0))
			Expect(gitClient.CloneCallCount()).To(Equal(0))
			Expect(gitClient.WriteCallCount()).To(Equal(0))
			Expect(kubeClient.ApplyCallCount()).To(Equal(0))
		})
	})

	Context("check for default values on AddParameters", func() {
		It("default values for path and deploymentType and branch should be correct", func() {
			addParams := AddParams{}
			addParams.Url = "http://github.com/weaveworks/testrepo"

			updated, err := appSrv.(*App).updateParametersIfNecessary(ctx, gitProviders, addParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(updated.DeploymentType).To(Equal(DefaultDeploymentType))
			Expect(updated.Path).To(Equal(DefaultPath))
			Expect(updated.Branch).To(Equal(DefaultBranch))
		})

		It("should fail when giving a wrong url format", func() {
			addParams := AddParams{}
			addParams.Url = "{http:/-*wrong-url-827"

			_, err := appSrv.(*App).updateParametersIfNecessary(ctx, gitProviders, addParams)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("error normalizing url"))
			Expect(err.Error()).Should(ContainSubstring(addParams.Url))

		})
	})

	Context("ensure that generated resource names are <= 63 characters in length", func() {
		It("ensures that app names are <= 63 characters", func() {
			addParams.Name = "a23456789012345678901234567890123456789012345678901234567890123"
			Expect(appSrv.Add(gitClient, gitProviders, addParams)).To(Succeed())
			info, err := getAppResourceInfo(makeWegoApplication(addParams), "cluster")
			Expect(err).ToNot(HaveOccurred())
			Expect(info.automationAppsDirKustomizationName()).To(Equal("wego-" + getHash(fmt.Sprintf("%s-apps-dir", addParams.Name))))
			addParams.Name = "a234567890123456789012345678901234567890123456789012345678901234"
			Expect(appSrv.Add(gitClient, gitProviders, addParams)).ShouldNot(Succeed())
		})

		It("ensures that url base names are <= 63 characters when used as names", func() {
			addParams.Url = "https://github.com/foo/a23456789012345678901234567890123456789012345678901234567890123"
			localParams, err := appSrv.(*App).updateParametersIfNecessary(ctx, gitProviders, addParams)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(appSrv.Add(gitClient, gitProviders, localParams)).To(Succeed())
			addParams.Name = ""
			addParams.Url = "https://github.com/foo/a234567890123456789012345678901234567890123456789012345678901234"
			_, err = appSrv.(*App).updateParametersIfNecessary(ctx, gitProviders, addParams)
			Expect(err).Should(HaveOccurred())
		})

		It("specifies a short cluster name, base url, and app name and gets them all included in resource names", func() {
			addParams.Url = "https://github.com/foo/url-base"
			addParams.Name = "app-name"
			info, err := getAppResourceInfo(makeWegoApplication(addParams), "cluster")
			Expect(err).ToNot(HaveOccurred())
			Expect(info.automationAppsDirKustomizationName()).To(Equal("app-name-apps-dir"))
			Expect(info.automationTargetDirKustomizationName()).To(Equal("cluster-app-name"))
			Expect(info.repoSecretName(addParams.Url).String()).To(Equal("wego-cluster-url-base"))
		})

		It("specifies a cluster name, base url, and app name that generate 63 characters and gets them all included in resource names", func() {
			addParams.Url = "https://github.com/foo/u"
			addParams.Name = "a12345"
			info, err := getAppResourceInfo(makeWegoApplication(addParams), "c2345678901234567890123456789012345678901234567890123456")
			Expect(err).ToNot(HaveOccurred())

			Expect(info.automationTargetDirKustomizationName()).To(Equal("c2345678901234567890123456789012345678901234567890123456-a12345"))
			Expect(info.repoSecretName(addParams.Url).String()).To(Equal("wego-c2345678901234567890123456789012345678901234567890123456-u"))
		})

		It("specifies a long cluster name, base url, and app name that generate 64 characters and gets hashed resource names", func() {
			addParams.Name = "a123456"
			addParams.Url = "https://github.com/foo/u1"
			clusterName := "c2345678901234567890123456789012345678901234567890123456"
			info, err := getAppResourceInfo(makeWegoApplication(addParams), clusterName)
			Expect(err).ToNot(HaveOccurred())

			kustName := info.automationTargetDirKustomizationName()
			secretName := info.repoSecretName(addParams.Url).String()
			repoName := generateResourceName(addParams.Url)

			Expect(kustName).To(Equal("wego-" + getHash(fmt.Sprintf("%s-%s", clusterName, addParams.Name))))
			Expect(secretName).To(Equal("wego-" + getHash(fmt.Sprintf("wego-%s-%s", clusterName, repoName))))
		})
	})
})

var _ = Describe("Test app hash", func() {

	It("should return right hash for a helm app", func() {

		app := wego.Application{
			Spec: wego.ApplicationSpec{
				Branch:         "main",
				URL:            "https://github.com/owner/repo1",
				DeploymentType: wego.DeploymentTypeHelm,
			},
		}
		app.Name = "nginx"

		info, err := getAppResourceInfo(app, "my-cluster")
		Expect(err).ToNot(HaveOccurred())

		appHash := info.getAppHash()
		expectedHash := getHash(app.Spec.URL, app.Name, app.Spec.Branch)

		Expect(appHash).To(Equal("wego-" + expectedHash))

	})

	It("should return right hash for a kustomize app", func() {
		app := wego.Application{
			Spec: wego.ApplicationSpec{
				Branch:         "main",
				URL:            "https://github.com/owner/repo1",
				Path:           "custompath",
				DeploymentType: wego.DeploymentTypeKustomize,
			},
		}

		info, err := getAppResourceInfo(app, "my-cluster")
		Expect(err).ToNot(HaveOccurred())

		appHash := info.getAppHash()
		expectedHash := getHash(app.Spec.URL, app.Spec.Path, app.Spec.Branch)

		Expect(appHash).To(Equal("wego-" + expectedHash))

	})
})

var _ = Describe("Add Gitlab", func() {
	var _ = BeforeEach(func() {
		addParams = AddParams{
			Url:            "https://gitlab.com/foo/bar",
			Path:           "./kustomize",
			Branch:         "main",
			Dir:            ".",
			DeploymentType: "kustomize",
			Namespace:      wego.DefaultNamespace,
			AppConfigUrl:   "NONE",
			AutoMerge:      true,
		}

		gitProviders.GetDefaultBranchReturns("main", nil)

		ctx = context.Background()
	})

	Context("add app with config in app repo", func() {
		BeforeEach(func() {
			addParams.Url = "ssh://git@gitlab.com/foo/bar.git"
			addParams.AppConfigUrl = ""

			gitClient.OpenStub = func(s string) (*gogit.Repository, error) {
				r, err := gogit.Init(memory.NewStorage(), memfs.New())
				Expect(err).ShouldNot(HaveOccurred())

				_, err = r.CreateRemote(&config.RemoteConfig{
					Name: "origin",
					URLs: []string{"git@gitlab.com:foo/bar.git"},
				})
				Expect(err).ShouldNot(HaveOccurred())
				return r, nil
			}
		})

		Describe("generates source manifest", func() {
			It("creates GitRepository when source type is git", func() {
				addParams.SourceType = wego.SourceTypeGit
				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceGitCallCount()).To(Equal(1))

				name, url, branch, secretRef, namespace := fluxClient.CreateSourceGitArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(url).To(Equal("ssh://git@gitlab.com/foo/bar.git"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("wego-test-cluster-bar"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
			})
		})
	})

	Context("add app from a subgroup", func() {
		BeforeEach(func() {
			addParams.Url = "ssh://git@gitlab.com/group/subgroup/bar.git"
			addParams.AppConfigUrl = ""

			gitClient.OpenStub = func(s string) (*gogit.Repository, error) {
				r, err := gogit.Init(memory.NewStorage(), memfs.New())
				Expect(err).ShouldNot(HaveOccurred())

				_, err = r.CreateRemote(&config.RemoteConfig{
					Name: "origin",
					URLs: []string{"git@gitlab.com:group/subgroup/bar.git"},
				})
				Expect(err).ShouldNot(HaveOccurred())
				return r, nil
			}
		})

		Describe("generates source manifest", func() {
			It("creates GitRepository when source type is git", func() {
				addParams.SourceType = wego.SourceTypeGit
				err := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceGitCallCount()).To(Equal(1))

				name, url, branch, secretRef, namespace := fluxClient.CreateSourceGitArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(url).To(Equal("ssh://git@gitlab.com/group/subgroup/bar.git"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("wego-test-cluster-bar"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
			})
		})
	})
})
var _ = Describe("New directory structure", func() {
	var _ = BeforeEach(func() {
		addParams = AddParams{
			Url:                      "https://github.com/user/repo",
			Path:                     "./kustomize",
			Branch:                   "main",
			Name:                     "foo",
			Dir:                      ".",
			DeploymentType:           "kustomize",
			Namespace:                "wego-system",
			AppConfigUrl:             "https://github.com/foo/bar",
			SourceType:               wego.SourceTypeGit,
			AutoMerge:                true,
			MigrateToNewDirStructure: utils.MigrateToNewDirStructure,
		}
		manifestsByPath = map[string][]byte{}
		createdResources = map[ResourceKind]map[string]bool{}
		repoPath := ""

		gitProviders.GetDefaultBranchStub = func(ctx context.Context, url gitproviders.RepoURL) (string, error) {
			return "main", nil
		}
		gitClient.WriteStub = func(path string, manifest []byte) error {
			manifestsByPath[path] = manifest
			if (repoPath) != "" {
				Expect(os.MkdirAll(filepath.Join(repoPath, filepath.Dir(path)), 0700)).To(Succeed(), "failed creating directory")
				Expect(os.WriteFile(filepath.Join(repoPath, path), manifest, 0666)).To(Succeed(), "failed writing file", path)
			}
			return nil //storeCreatedResource(manifest)
		}
		gitClient.CloneStub = func(arg1 context.Context, dir string, arg3 string, arg4 string) (bool, error) {
			for p, m := range manifestsByPath {
				// Put the manifests files written so far into this new clone dir
				Expect(os.MkdirAll(filepath.Join(dir, filepath.Dir(p)), 0700)).To(Succeed(), "failed creating directory")
				Expect(os.WriteFile(filepath.Join(dir, p), m, 0666)).To(Succeed(), "failed writing file", p)

			}

			return false, nil
		}

		ctx = context.Background()
	})

	It("adds app to the cluster kustomization file", func() {
		addParams.SourceType = wego.SourceTypeGit

		Expect(appSrv.Add(gitClient, gitProviders, addParams)).ShouldNot(HaveOccurred())
		Expect(manifestsByPath[filepath.Join(git.WegoRoot, git.WegoAppDir, addParams.Name, "kustomization.yaml")]).ToNot(BeNil())
		cname, err := kubeClient.GetClusterName(context.Background())
		Expect(err).To(BeNil())
		clusterKustomizeFile := manifestsByPath[filepath.Join(git.WegoRoot, git.WegoClusterDir, cname, "/user/kustomization.yaml")]
		Expect(clusterKustomizeFile).ToNot(BeNil())

		manifestMap := map[string]interface{}{}

		Expect(yaml.Unmarshal(clusterKustomizeFile, &manifestMap)).ShouldNot(HaveOccurred())
		r := manifestMap["resources"].([]interface{})
		Expect(len(r)).To(Equal(1))
		Expect(r[0].(string)).To(Equal("../../../apps/" + addParams.Name))
	})
	It("adds second app to the cluster kustomization file", func() {
		addParams.SourceType = wego.SourceTypeGit
		origName := addParams.Name
		Expect(appSrv.Add(gitClient, gitProviders, addParams)).ShouldNot(HaveOccurred())
		addParams.Name = "sally"
		Expect(appSrv.Add(gitClient, gitProviders, addParams)).ShouldNot(HaveOccurred())
		Expect(manifestsByPath[filepath.Join(git.WegoRoot, git.WegoAppDir, addParams.Name, "kustomization.yaml")]).ToNot(BeNil())
		cname, err := kubeClient.GetClusterName(context.Background())
		Expect(err).To(BeNil())
		clusterKustomizeFile := manifestsByPath[filepath.Join(git.WegoRoot, git.WegoClusterDir, cname, "user", "kustomization.yaml")]
		Expect(clusterKustomizeFile).ToNot(BeNil())
		manifestMap := map[string]interface{}{}

		err = yaml.Unmarshal(clusterKustomizeFile, &manifestMap)
		Expect(err).ShouldNot(HaveOccurred())
		r := manifestMap["resources"].([]interface{})
		Expect(len(r)).To(Equal(2))
		Expect(r[0].(string)).To(Equal("../../../apps/" + origName))
		Expect(r[1].(string)).To(Equal("../../../apps/" + addParams.Name))
	})

})

func getHash(inputs ...string) string {
	final := []byte(strings.Join(inputs, ""))
	return fmt.Sprintf("%x", md5.Sum(final))
}

func emptySource() Source {
	return source{yaml: []byte{}}
}

func emptyAutomation() Automation {
	return automation{yaml: []byte{}}
}

func emptyAppManifest() AppManifest {
	return appManifest{yaml: []byte{}}
}
