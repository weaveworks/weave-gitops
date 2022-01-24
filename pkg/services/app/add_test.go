package app

import (
	"context"
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
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"sigs.k8s.io/yaml"
)

var (
	addParams       AddParams
	ctx             context.Context
	manifestsByPath map[string][]byte = map[string][]byte{}
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
			ConfigRepo:     "",
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
		err := appSrv.Add(gitClient, gitProviders, addParams)
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

	It("validates config-repo is set when source is helm", func() {
		addParams.Chart = "my-chart"
		addParams.Url = "https://my-chart.com"
		addParams.ConfigRepo = ""

		err := appSrv.Add(gitClient, gitProviders, addParams)
		Expect(err.Error()).Should(HaveSuffix("--config-repo should be provided"))
	})

	It("validates invalid chartname is handled", func() {
		addParams.Chart = "invalid_Chartname.bar"
		addParams.Url = "https://my-chart.com"
		addParams.ConfigRepo = ""

		err := appSrv.Add(gitClient, gitProviders, addParams)
		Expect(err).Should(HaveOccurred())
	})

	It("validates invalid name is handled", func() {
		addParams.Name = "myapp.isInvalid"
		addParams.Url = "https://github.com/weaveworks/weave-gitops-interlock.git"
		addParams.ConfigRepo = ""

		err := appSrv.Add(gitClient, gitProviders, addParams)
		Expect(err).Should(HaveOccurred())
	})

	It("validates ssh git reponame with invalid characters is handled", func() {
		addParams.Url = "git@github.com:weaveworks/weave-gitops-interlockUPPER_LOWER.git"
		addParams.ConfigRepo = ""

		err := appSrv.Add(gitClient, gitProviders, addParams)
		Expect(err).Should(HaveOccurred())
	})

	It("validates http reponame with invalid characters is handled", func() {
		addParams.Url = "https://github.com/weaveworks/weave-gitops-interlockUPPER_LOWER.git"
		addParams.ConfigRepo = ""

		err := appSrv.Add(gitClient, gitProviders, addParams)
		Expect(err).Should(HaveOccurred())
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
			addParams.ConfigRepo = ""

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
			It("validates namespace passed as target namespace", func() {
				addParams.Url = "https://charts.kube-ops.io"
				addParams.Chart = "loki"
				addParams.HelmReleaseTargetNamespace = "sock-shop"
				addParams.ConfigRepo = "ssh://git@github.com/owner/config-repo.git"

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
				addParams.ConfigRepo = "ssh://git@github.com/owner/config-repo.git"

				goodNamespaceErr := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(goodNamespaceErr).ShouldNot(HaveOccurred())

				kubeClient.NamespacePresentReturns(false, nil)

				badNamespaceErr := appSrv.Add(gitClient, gitProviders, addParams)
				Expect(badNamespaceErr.Error()).To(HavePrefix("could not update parameters: Helm Release Target Namespace sock-shop does not exist"))
			})

			It("fails if deployment type is invalid", func() {
				addParams.DeploymentType = "foo"

				Expect(appSrv.Add(gitClient, gitProviders, addParams)).ShouldNot(Succeed())
			})
		})
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

	Context("Fails to add app that has the prefix wego as an app name", func() {
		It("Adds an app with ", func() {
			addParams.Name = "wego-appname"
			err := appSrv.Add(gitClient, gitProviders, addParams)
			Expect(err.Error()).Should(ContainSubstring("the prefix 'wego' is used by weave gitops and is not allowed for an app name"))
		})
	})

	Context("ensure that app names are <= 63 characters in length", func() {
		It("ensures that app names are <= 63 characters", func() {
			addParams.Name = "a23456789012345678901234567890123456789012345678901234567890123"
			Expect(appSrv.Add(gitClient, gitProviders, addParams)).To(Succeed())

			addParams.Name = "a234567890123456789012345678901234567890123456789012345678901234"
			Expect(appSrv.Add(gitClient, gitProviders, addParams)).ShouldNot(Succeed())
		})

		It("ensures that url base names are <= 63 characters when used as names", func() {
			addParams.Url = "https://github.com/foo/a23456789012345678901234567890123456789012345678901234567890123"
			localParams, err := appSrv.(*AppSvc).updateParametersIfNecessary(ctx, gitProviders, addParams)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(appSrv.Add(gitClient, gitProviders, localParams)).To(Succeed())
			addParams.Name = ""
			addParams.Url = "https://github.com/foo/a234567890123456789012345678901234567890123456789012345678901234"
			_, err = appSrv.(*AppSvc).updateParametersIfNecessary(ctx, gitProviders, addParams)
			Expect(err).Should(HaveOccurred())
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
			ConfigRepo:     "",
			AutoMerge:      true,
		}

		gitProviders.GetDefaultBranchReturns("main", nil)

		ctx = context.Background()
	})

	Context("add app with config in app repo", func() {
		BeforeEach(func() {
			addParams.Url = "ssh://git@gitlab.com/foo/bar.git"
			addParams.ConfigRepo = ""

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
				Expect(secretRef).To(Equal("wego-gitlab-bar"))
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
			ConfigRepo:               "https://github.com/foo/bar",
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
		clusterKustomizeFile := manifestsByPath[filepath.Join(git.WegoRoot, git.WegoClusterDir, cname, git.WegoClusterUserWorkloadDir, "kustomization.yaml")]
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
		clusterKustomizeFile := manifestsByPath[filepath.Join(git.WegoRoot, git.WegoClusterDir, cname, git.WegoClusterUserWorkloadDir, "kustomization.yaml")]
		Expect(clusterKustomizeFile).ToNot(BeNil())
		manifestMap := map[string]interface{}{}

		Expect(yaml.Unmarshal(clusterKustomizeFile, &manifestMap)).Should(Succeed())

		r := manifestMap["resources"].([]interface{})
		Expect(len(r)).To(Equal(2))
		Expect(r[0].(string)).To(Equal("../../../apps/" + origName))
		Expect(r[1].(string)).To(Equal("../../../apps/" + addParams.Name))
	})

	It("deals with a cluster dir with additional subdirectories", func() {
		kubeClient.GetClusterNameReturns("arn:aws:eks:us-west-2:01234567890:cluster/default-my-wego-control-plan", nil)
		appNames := []string{
			"oracle",
			"sqlserver",
		}
		for _, a := range appNames {
			addParams.Name = a
			Expect(appSrv.Add(gitClient, gitProviders, addParams)).Should(Succeed())
		}

		Expect(manifestsByPath[filepath.Join(git.WegoRoot, git.WegoAppDir, addParams.Name, "kustomization.yaml")]).ToNot(BeNil())
		cname, err := kubeClient.GetClusterName(context.Background())
		Expect(err).To(BeNil())
		clusterKustomizeFile := manifestsByPath[filepath.Join(git.WegoRoot, git.WegoClusterDir, cname, git.WegoClusterUserWorkloadDir, "kustomization.yaml")]
		Expect(clusterKustomizeFile).ToNot(BeNil())
		manifestMap := map[string]interface{}{}

		Expect(yaml.Unmarshal(clusterKustomizeFile, &manifestMap)).Should(Succeed())

		resources := manifestMap["resources"].([]interface{})
		Expect(len(resources)).To(Equal(len(appNames)))
		for i, name := range appNames {
			Expect(resources[i].(string)).To(Equal("../../../../apps/" + name))
		}

	})

})
