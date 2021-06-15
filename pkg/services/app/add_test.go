package app_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
)

var (
	gitClient    *gitfakes.FakeGit
	fluxClient   *fluxfakes.FakeFlux
	kubeClient   *kubefakes.FakeKube
	gitProviders *gitprovidersfakes.FakeGitProviderHandler

	appSrv        app.AppService
	defaultParams app.AddParams
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

	deps := &app.Dependencies{
		Git:          gitClient,
		Flux:         fluxClient,
		Kube:         kubeClient,
		GitProviders: gitProviders,
	}

	appSrv = app.New(deps)

	defaultParams = app.AddParams{
		Url:            "https://github.com/foo/bar",
		Path:           "./kustomize",
		Branch:         "main",
		Dir:            ".",
		DeploymentType: "kustomize",
		Namespace:      "wego-system",
	}
})

var _ = Describe("Add", func() {
	BeforeEach(func() {

	})

	It("checks for cluster status", func() {
		err := appSrv.Add(defaultParams)
		Expect(err).ShouldNot(HaveOccurred())

		Expect(kubeClient.GetClusterStatusCallCount()).To(Equal(1))
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
		Expect(repoUrl).To(Equal("ssh://git@github.com/foo/bar.git"))
		Expect(namespace).To(Equal("wego-system"))

		owner, repoName, deployKey := gitProviders.UploadDeployKeyArgsForCall(0)
		Expect(owner).To(Equal("foo"))
		Expect(repoName).To(Equal("bar"))
		Expect(deployKey).To(Equal([]byte("deploy key")))
	})

	Context("add app with no config repo", func() {
		BeforeEach(func() {
			defaultParams.AutomationRepo = ""
			defaultParams.CommitManifests = false
		})

		Describe("generates source manifest", func() {
			It("creates GitRepository when source type is git", func() {
				defaultParams.SourceType = string(app.SourceTypeGit)

				err := appSrv.Add(defaultParams)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceGitCallCount()).To(Equal(1))

				name, url, branch, secretRef, namespace := fluxClient.CreateSourceGitArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(url).To(Equal("ssh://git@github.com/foo/bar.git"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("weave-gitops-test-cluster"))
				Expect(namespace).To(Equal("wego-system"))
			})

			It("creates HelmResitory when source type is helm", func() {
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
				defaultParams.DeploymentType = string(app.DeployTypeHelm)

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
})
