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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/services/gitopswriter"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"sigs.k8s.io/yaml"
)

var (
	addParams       AddParams
	ctx             context.Context
	manifestsByPath map[string][]byte = map[string][]byte{}
)

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
		Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
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
		Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
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
			updated, err := appSrv.(*AppSvc).updateParametersIfNecessary(ctx, gitProviders, addParams)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(updated.Branch).To(Equal("an-unusual-branch"))
		})

		It("Allows a specified branch to override the repo's default branch", func() {
			addParams.Branch = "an-overriding-branch"
			updated, err := appSrv.(*AppSvc).updateParametersIfNecessary(ctx, gitProviders, addParams)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(updated.Branch).To(Equal("an-overriding-branch"))
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
				Expect(url.String()).To(Equal("ssh://git@github.com/foo/bar.git"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("wego-test-cluster-bar"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
			})

			It("creates HelmRepository when source type is helm", func() {
				addParams.Url = "https://charts.kube-ops.io"
				addParams.Chart = "loki"
				addParams.AppConfigUrl = "ssh://git@github.com/owner/config-repo.git"

				Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
				Expect(fluxClient.CreateSourceHelmCallCount()).To(Equal(1))

				name, url, namespace := fluxClient.CreateSourceHelmArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(url).To(Equal("https://charts.kube-ops.io"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
			})
		})

		Describe("generates application goat", func() {
			It("creates a kustomization if deployment type kustomize", func() {
				Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
				Expect(fluxClient.CreateKustomizationCallCount()).To(Equal(1))

				name, source, path, namespace := fluxClient.CreateKustomizationArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("./kustomize"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
			})

			It("creates helm release using a helm repository if source type is helm", func() {
				addParams.Url = "https://charts.kube-ops.io"
				addParams.Chart = "loki"
				addParams.AppConfigUrl = "ssh://github.com/owner/repo"

				Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
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

				Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
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

				Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
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

				Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
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

				err := appSrv.Add(gitClient, gitProviders, gitClient, gitProviders, addParams)
				Expect(err).Should(HaveOccurred())
			})
		})

		Context("when using URL", func() {
			BeforeEach(func() {
				addParams.Url = "ssh://git@github.com/foo/bar.git"
			})

			It("clones the repo to a temp dir", func() {
				Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
				Expect(gitClient.CloneCallCount()).To(Equal(1))
				_, repoDir, url, branch := gitClient.CloneArgsForCall(0)

				Expect(repoDir).To(ContainSubstring("user-repo-"))
				Expect(url).To(Equal("ssh://git@github.com/foo/bar.git"))
				Expect(branch).To(Equal("main"))
			})

			It("writes the files to the disk", func() {
				addParams.AppConfigUrl = addParams.Url // so we know the root is ".wego"
				fluxClient.CreateSourceGitReturns(dummyGitSource, nil)
				fluxClient.CreateKustomizationReturns([]byte("kustomization"), nil)

				Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
				Expect(gitClient.WriteCallCount()).To(Equal(5))
			})
		})

		It("commit and pushes the files", func() {
			Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
			Expect(gitClient.CommitCallCount()).To(Equal(1))

			msg, filters := gitClient.CommitArgsForCall(0)
			Expect(msg).To(Equal(git.Commit{
				Author:  git.Author{Name: "Weave Gitops", Email: "weave-gitops@weave.works"},
				Message: gitopswriter.AddCommitMessage,
			}))

			Expect(filters[0](".weave-gitops/apps/bar/app.yaml")).To(BeTrue())
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

				Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
				Expect(fluxClient.CreateSourceGitCallCount()).To(Equal(1))

				name, url, branch, secretRef, namespace := fluxClient.CreateSourceGitArgsForCall(0)
				Expect(name).To(Equal("repo"))
				Expect(url.String()).To(Equal("ssh://git@github.com/user/repo.git"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("wego-test-cluster-repo"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
			})

			It("creates HelmRepository when source type is helm", func() {
				addParams.Url = "https://charts.kube-ops.io"
				addParams.Chart = "loki"

				Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
				Expect(fluxClient.CreateSourceHelmCallCount()).To(Equal(1))

				name, url, namespace := fluxClient.CreateSourceHelmArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(url).To(Equal("https://charts.kube-ops.io"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
			})
		})

		It("clones the repo to a temp dir", func() {
			Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
			Expect(gitClient.CloneCallCount()).To(Equal(1))
			_, repoDir, url, branch := gitClient.CloneArgsForCall(0)

			Expect(repoDir).To(ContainSubstring("user-repo-"))
			Expect(url).To(Equal("ssh://git@github.com/foo/bar.git"))
			Expect(branch).To(Equal("main"))
		})

		It("writes the files to the disk", func() {
			fluxClient.CreateSourceGitReturns(dummyGitSource, nil)
			fluxClient.CreateKustomizationReturns([]byte("kustomization"), nil)

			Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
			Expect(gitClient.WriteCallCount()).To(Equal(5))

			found := 0
			for idx := 0; idx < 3; idx++ {
				path, _ := gitClient.WriteArgsForCall(idx)
				if path == ".weave-gitops/apps/repo/app.yaml" || path == ".weave-gitops/apps/repo/repo-gitops-source.yaml" || path == ".weave-gitops/apps/repo/repo-gitops-deploy.yaml" {
					found++
				}
			}

			Expect(found).To(Equal(3))
		})

		It("commit and pushes the files", func() {
			Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
			Expect(gitClient.CommitCallCount()).To(Equal(1))

			msg, filters := gitClient.CommitArgsForCall(0)
			Expect(msg).To(Equal(git.Commit{
				Author:  git.Author{Name: "Weave Gitops", Email: "weave-gitops@weave.works"},
				Message: gitopswriter.AddCommitMessage,
			}))

			Expect(len(filters)).To(Equal(1))
		})
	})

	Context("when creating a pull request", func() {
		//		var app models.Application
		//		var emptyManifests []automation.AutomationManifest

		//		clusterName := "cluster"

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
			app, err = makeApplication(addParams)
			Expect(err).ToNot(HaveOccurred())

			// emptyManifests = []automation.AutomationManifest{
			//  automation.AutomationManifest{
			//      Path:    automation.AppYamlPath(app),
			//      Content: []byte{}},
			//  automation.AutomationManifest{
			//      Path:    automation.AppAutomationSourcePath(app, clusterName),
			//      Content: []byte{}},
			//  automation.AutomationManifest{
			//      Path:    automation.AppAutomationDeployPath(app, clusterName),
			//      Content: []byte{}}}
		})

		// Context("uses the default app branch for config in app repository", func() {
		//  BeforeEach(func() {
		//      addParams.AppConfigUrl = ""
		//  })

		//  It("creates the pull request against the default branch for an org app repository", func() {
		//      Expect(appSrv.(*AppSvc).createPullRequestToRepo(ctx, app, app.GitSourceURL, emptyManifests)).To(Succeed())
		//      _, _, prInfo := gitProviders.CreatePullRequestArgsForCall(0)
		//      Expect(prInfo.TargetBranch).To(Equal("default-app-branch"))
		//  })

		//  It("creates the pull request against the default branch for a user app repository", func() {
		//      Expect(appSrv.(*AppSvc).createPullRequestToRepo(ctx, app, app.GitSourceURL, emptyManifests)).To(Succeed())
		//      _, _, prInfo := gitProviders.CreatePullRequestArgsForCall(0)
		//      Expect(prInfo.TargetBranch).To(Equal("default-app-branch"))
		//  })
		// })

		// Context("uses the default config branch for external config", func() {
		//  BeforeEach(func() {
		//      addParams.AppConfigUrl = "https://github.com/foo/bar"
		//  })

		//  It("creates the pull request against the default branch for an org config repository", func() {
		//      Expect(appSrv.(*AppSvc).createPullRequestToRepo(ctx, app, app.ConfigURL, emptyManifests)).To(Succeed())
		//      _, _, prInfo := gitProviders.CreatePullRequestArgsForCall(0)
		//      Expect(prInfo.TargetBranch).To(Equal("default-config-branch"))
		//  })

		//  It("creates the pull request against the default branch for a user config repository", func() {
		//      Expect(appSrv.(*AppSvc).createPullRequestToRepo(ctx, app, app.ConfigURL, emptyManifests)).To(Succeed())
		//      _, _, prInfo := gitProviders.CreatePullRequestArgsForCall(0)
		//      Expect(prInfo.TargetBranch).To(Equal("default-config-branch"))
		//  })
		// })
	})

	Context("when using dry-run", func() {
		It("doesnt execute any action", func() {
			addParams.DryRun = true
			addParams.AutoMerge = true

			Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
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

			updated, err := appSrv.(*AppSvc).updateParametersIfNecessary(ctx, gitProviders, addParams)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(updated.DeploymentType).To(Equal(DefaultDeploymentType))
			Expect(updated.Path).To(Equal(DefaultPath))
			Expect(updated.Branch).To(Equal(DefaultBranch))
		})

		It("should fail when giving a wrong url format", func() {
			addParams := AddParams{}
			addParams.Url = "{http:/-*wrong-url-827"

			_, err := appSrv.(*AppSvc).updateParametersIfNecessary(ctx, gitProviders, addParams)
			Expect(err).Should(HaveOccurred())
			Expect(err.Error()).Should(ContainSubstring("error normalizing url"))
			Expect(err.Error()).Should(ContainSubstring(addParams.Url))

		})
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
				Expect(url.String()).To(Equal("ssh://git@gitlab.com/foo/bar.git"))
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

		gitProviders.GetDefaultBranchStub = func(ctx context.Context, url gitproviders.RepoURL) (string, error) {
			return "main", nil
		}
		gitClient.WriteStub = func(path string, manifest []byte) error {
			manifestsByPath[path] = manifest

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

		Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
		Expect(manifestsByPath[filepath.Join(git.WegoRoot, git.WegoAppDir, addParams.Name, "kustomization.yaml")]).ToNot(BeNil())
		cname, err := kubeClient.GetClusterName(context.Background())
		Expect(err).To(BeNil())
		clusterKustomizeFile := manifestsByPath[filepath.Join(git.WegoRoot, git.WegoClusterDir, cname, "/user/kustomization.yaml")]
		Expect(clusterKustomizeFile).ToNot(BeNil())

		manifestMap := map[string]interface{}{}

		Expect(yaml.Unmarshal(clusterKustomizeFile, &manifestMap)).Should(Succeed())
		r := manifestMap["resources"].([]interface{})
		Expect(len(r)).To(Equal(1))
		Expect(r[0].(string)).To(Equal("../../../apps/" + addParams.Name))
	})
	It("adds second app to the cluster kustomization file", func() {
		addParams.SourceType = wego.SourceTypeGit
		origName := addParams.Name
		Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
		addParams.Name = "sally"
		Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
		Expect(manifestsByPath[filepath.Join(git.WegoRoot, git.WegoAppDir, addParams.Name, "kustomization.yaml")]).ToNot(BeNil())
		cname, err := kubeClient.GetClusterName(context.Background())
		Expect(err).To(BeNil())
		clusterKustomizeFile := manifestsByPath[filepath.Join(git.WegoRoot, git.WegoClusterDir, cname, "user", "kustomization.yaml")]
		Expect(clusterKustomizeFile).ToNot(BeNil())
		manifestMap := map[string]interface{}{}

		Expect(yaml.Unmarshal(clusterKustomizeFile, &manifestMap)).Should(Succeed())

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
