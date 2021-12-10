package automation

import (
	"context"
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"sigs.k8s.io/yaml"
)

var (
	ctx context.Context

	emptyRepoURL = gitproviders.RepoURL{}
)

func extractApp(app models.Application, automation []AutomationManifest) (AutomationManifest, []AutomationManifest) {
	return extractManifest(AppYamlPath(app), automation)
}

func extractAppKustomize(app models.Application, automation []AutomationManifest) (AutomationManifest, []AutomationManifest) {
	return extractManifest(AppAutomationKustomizePath(app), automation)
}

func extractManifest(path string, automation []AutomationManifest) (AutomationManifest, []AutomationManifest) {
	var m AutomationManifest

	others := []AutomationManifest{}

	for _, am := range automation {
		if am.Path == path {
			m = am
		} else {
			others = append(others, am)
		}
	}

	return m, others
}

func createRepoURL(url string) gitproviders.RepoURL {
	repoURL, err := gitproviders.NewRepoURL(url)
	Expect(err).NotTo(HaveOccurred())

	return repoURL
}

var _ = Describe("Generate manifests", func() {
	var app models.Application

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
			app.ConfigRepo = emptyRepoURL
		})

		Describe("generates source manifest", func() {
			It("creates GitRepository when source type is git", func() {
				app.SourceType = models.SourceTypeGit
				results, err := automationGen.GenerateAutomation(ctx, app, "test-cluster")
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceGitCallCount()).To(Equal(1))

				name, url, branch, secretRef, namespace := fluxClient.CreateSourceGitArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(url.String()).To(Equal("ssh://git@github.com/foo/bar.git"))
				Expect(branch).To(Equal("main"))
				Expect(secretRef).To(Equal("wego-github-bar"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))

				appManifest, nonApps := extractApp(app, results)
				Expect(len(nonApps)).To(Equal(len(results) - 1))

				wegoApp := AppToWegoApp(app)
				wegoApp.ObjectMeta.Labels = map[string]string{
					WeGOAppIdentifierLabelKey: GetAppHash(app),
				}

				bytes, err := yaml.Marshal(&wegoApp)
				Expect(err).To(BeNil())
				Expect(string(sanitizeK8sYaml(bytes))).To(Equal(string(appManifest.Content)))

				appKustomizeManifest, otherManifests := extractAppKustomize(app, results)
				Expect(len(nonApps)).To(Equal(len(results) - 1))

				km, err := createAppKustomize(app, otherManifests)
				Expect(err).To(BeNil())
				Expect(km.Content).To(Equal(appKustomizeManifest.Content))
			})

			It("creates HelmRepository when source type is helm", func() {
				app.HelmSourceURL = "https://charts.kube-ops.io"
				app.Path = "loki"
				app.Name = "loki"
				app.SourceType = models.SourceTypeHelm
				app.ConfigRepo = createRepoURL("ssh://git@github.com/owner/config-repo.git")

				results, err := automationGen.GenerateAutomation(ctx, app, "test-cluster")
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateSourceHelmCallCount()).To(Equal(1))

				name, url, namespace := fluxClient.CreateSourceHelmArgsForCall(0)
				Expect(name).To(Equal("loki"))
				Expect(url).To(Equal("https://charts.kube-ops.io"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))

				appManifest, nonApps := extractApp(app, results)
				Expect(len(nonApps)).To(Equal(len(results) - 1))

				wegoApp := AppToWegoApp(app)
				wegoApp.ObjectMeta.Labels = map[string]string{
					WeGOAppIdentifierLabelKey: GetAppHash(app),
				}

				bytes, err := yaml.Marshal(&wegoApp)
				Expect(err).To(BeNil())
				Expect(string(sanitizeK8sYaml(bytes))).To(Equal(string(appManifest.Content)))

				appKustomizeManifest, otherManifests := extractAppKustomize(app, results)
				Expect(len(nonApps)).To(Equal(len(results) - 1))

				km, err := createAppKustomize(app, otherManifests)
				Expect(err).To(BeNil())
				Expect(km.Content).To(Equal(appKustomizeManifest.Content))
			})
		})

		Describe("generates application goat", func() {
			It("creates a kustomization if deployment type kustomize", func() {
				_, err := automationGen.GenerateAutomation(ctx, app, "test-cluster")
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateKustomizationCallCount()).To(Equal(1))

				name, source, path, namespace := fluxClient.CreateKustomizationArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("./kustomize"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
			})

			It("creates a helm release using a git source if source type is git", func() {
				app.AutomationType = models.AutomationTypeHelm
				app.Path = "loki"
				app.Name = "bar"
				app.ConfigRepo = createRepoURL("ssh://github.com/owner/repo")

				app.Path = "./charts/my-chart"
				app.AutomationType = models.AutomationTypeHelm

				_, err := automationGen.GenerateAutomation(ctx, app, "test-cluster")
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseGitRepositoryCallCount()).To(Equal(1))

				name, source, path, namespace, targetNamespace := fluxClient.CreateHelmReleaseGitRepositoryArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("./charts/my-chart"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
				Expect(targetNamespace).To(Equal(""))
			})

			It("creates a helm release for git repository with target namespace if source type is git", func() {
				app.Path = "./charts/my-chart"
				app.AutomationType = models.AutomationTypeHelm
				app.HelmTargetNamespace = "sock-shop"

				_, err := automationGen.GenerateAutomation(ctx, app, "test-cluster")
				Expect(err).ShouldNot(HaveOccurred())

				Expect(fluxClient.CreateHelmReleaseGitRepositoryCallCount()).To(Equal(1))

				name, source, path, namespace, targetNamespace := fluxClient.CreateHelmReleaseGitRepositoryArgsForCall(0)
				Expect(name).To(Equal("bar"))
				Expect(source).To(Equal("bar"))
				Expect(path).To(Equal("./charts/my-chart"))
				Expect(namespace).To(Equal(wego.DefaultNamespace))
				Expect(targetNamespace).To(Equal("sock-shop"))
			})
		})

		Context("add app with external config repo", func() {
			BeforeEach(func() {
				app.GitSourceURL = createRepoURL("https://github.com/user/repo")
				app.ConfigRepo = createRepoURL("https://github.com/foo/bar")
			})

			Describe("generates source manifest", func() {
				It("creates GitRepository when source type is git", func() {
					results, err := automationGen.GenerateAutomation(ctx, app, "test-cluster")
					Expect(err).ShouldNot(HaveOccurred())

					Expect(fluxClient.CreateSourceGitCallCount()).To(Equal(1))

					name, url, branch, secretRef, namespace := fluxClient.CreateSourceGitArgsForCall(0)
					Expect(name).To(Equal("bar"))
					Expect(url.String()).To(Equal("ssh://git@github.com/user/repo.git"))
					Expect(branch).To(Equal("main"))
					Expect(secretRef).To(Equal("wego-github-repo"))
					Expect(namespace).To(Equal(wego.DefaultNamespace))

					appManifest, nonApps := extractApp(app, results)
					Expect(len(nonApps)).To(Equal(len(results) - 1))

					wegoApp := AppToWegoApp(app)
					wegoApp.ObjectMeta.Labels = map[string]string{
						WeGOAppIdentifierLabelKey: GetAppHash(app),
					}

					bytes, err := yaml.Marshal(&wegoApp)
					Expect(err).To(BeNil())
					Expect(string(sanitizeK8sYaml(bytes))).To(Equal(string(appManifest.Content)))

					appKustomizeManifest, otherManifests := extractAppKustomize(app, results)
					Expect(len(nonApps)).To(Equal(len(results) - 1))

					km, err := createAppKustomize(app, otherManifests)
					Expect(err).To(BeNil())
					Expect(km.Content).To(Equal(appKustomizeManifest.Content))
				})

				It("creates HelmRepository when source type is helm", func() {
					app.HelmSourceURL = "https://charts.kube-ops.io"
					app.Path = "loki"
					app.Name = "loki"
					app.SourceType = models.SourceTypeHelm

					results, err := automationGen.GenerateAutomation(ctx, app, "test-cluster")
					Expect(err).ShouldNot(HaveOccurred())

					Expect(fluxClient.CreateSourceHelmCallCount()).To(Equal(1))

					name, url, namespace := fluxClient.CreateSourceHelmArgsForCall(0)
					Expect(name).To(Equal("loki"))
					Expect(url).To(Equal("https://charts.kube-ops.io"))
					Expect(namespace).To(Equal(wego.DefaultNamespace))

					appManifest, nonApps := extractApp(app, results)
					Expect(len(nonApps)).To(Equal(len(results) - 1))

					wegoApp := AppToWegoApp(app)
					wegoApp.ObjectMeta.Labels = map[string]string{
						WeGOAppIdentifierLabelKey: GetAppHash(app),
					}

					bytes, err := yaml.Marshal(&wegoApp)
					Expect(err).To(BeNil())
					Expect(string(sanitizeK8sYaml(bytes))).To(Equal(string(appManifest.Content)))

					appKustomizeManifest, otherManifests := extractAppKustomize(app, results)
					Expect(len(nonApps)).To(Equal(len(results) - 1))

					km, err := createAppKustomize(app, otherManifests)
					Expect(err).To(BeNil())
					Expect(km.Content).To(Equal(appKustomizeManifest.Content))
				})
			})

			Describe("generateAppYaml", func() {
				It("generates the app yaml", func() {
					myAppModel := models.Application{
						Name:                "my-app",
						Namespace:           wego.DefaultNamespace,
						HelmSourceURL:       "",
						GitSourceURL:        createRepoURL("ssh://git@github.com/example/my-source"),
						Branch:              "main",
						Path:                "manifests",
						AutomationType:      models.AutomationTypeKustomize,
						SourceType:          models.SourceTypeGit,
						HelmTargetNamespace: "",
					}

					hash := "wego-" + getHash(myAppModel.GitSourceURL.String(), myAppModel.Path, myAppModel.Branch)
					myApp := AppToWegoApp(myAppModel)

					myApp.ObjectMeta.Labels = map[string]string{WeGOAppIdentifierLabelKey: hash}

					out, err := generateAppYaml(myAppModel)
					Expect(err).To(BeNil())

					result := wego.Application{}
					// Convert back to a struct to make the comparison more forgiving.
					// A straight string/byte comparison doesn't account for un-ordered keys in yaml.
					Expect(yaml.Unmarshal(out.Content, &result))

					diff := cmp.Diff(result, myApp)

					// Not entirely sure how to get gomega to pretty-print the output of `diff`,
					// so we assert it should be empty above, and conditionally print the diff to make a nice assertion message.
					// `diff` is a formatted string
					if diff != "" {
						GinkgoT().Errorf("comparison failed: (-actual +expected): %s\n", diff)
					}
				})
			})

			Describe("generates application goat", func() {
				It("creates a kustomization if deployment type kustomize", func() {
					_, err := automationGen.GenerateAutomation(ctx, app, "test-cluster")
					Expect(err).ShouldNot(HaveOccurred())

					Expect(fluxClient.CreateKustomizationCallCount()).To(Equal(1))

					name, source, path, namespace := fluxClient.CreateKustomizationArgsForCall(0)
					Expect(name).To(Equal("bar"))
					Expect(source).To(Equal("bar"))
					Expect(path).To(Equal("./kustomize"))
					Expect(namespace).To(Equal(wego.DefaultNamespace))
				})

				It("creates helm release using a helm repository if source type is helm", func() {
					app.HelmSourceURL = "https://charts.kube-ops.io"
					app.SourceType = models.SourceTypeHelm
					app.AutomationType = models.AutomationTypeHelm
					app.Path = "loki"
					app.Name = "loki"
					app.ConfigRepo = createRepoURL("ssh://github.com/owner/repo")

					_, err := automationGen.GenerateAutomation(ctx, app, "test-cluster")
					Expect(err).ShouldNot(HaveOccurred())

					Expect(fluxClient.CreateHelmReleaseHelmRepositoryCallCount()).To(Equal(1))

					name, chart, namespace, targetNamespace := fluxClient.CreateHelmReleaseHelmRepositoryArgsForCall(0)
					Expect(name).To(Equal("loki"))
					Expect(chart).To(Equal("loki"))
					Expect(namespace).To(Equal(wego.DefaultNamespace))
					Expect(targetNamespace).To(Equal(""))
				})

				It("creates a helm release using a git source if source type is git", func() {
					app.AutomationType = models.AutomationTypeHelm
					app.Path = "loki"
					app.Name = "bar"
					app.ConfigRepo = createRepoURL("ssh://github.com/owner/repo")

					app.Path = "./charts/my-chart"
					app.AutomationType = models.AutomationTypeHelm

					_, err := automationGen.GenerateAutomation(ctx, app, "test-cluster")
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
					app.HelmSourceURL = "https://charts.kube-ops.io"
					app.SourceType = models.SourceTypeHelm
					app.Path = "loki"
					app.Name = "loki"
					app.HelmTargetNamespace = "sock-shop"
					app.AutomationType = models.AutomationTypeHelm
					app.ConfigRepo = createRepoURL("ssh://git@github.com/owner/config-repo.git")

					_, err := automationGen.GenerateAutomation(ctx, app, "test-cluster")
					Expect(err).ShouldNot(HaveOccurred())

					Expect(fluxClient.CreateHelmReleaseHelmRepositoryCallCount()).To(Equal(1))

					name, chart, namespace, targetNamespace := fluxClient.CreateHelmReleaseHelmRepositoryArgsForCall(0)
					Expect(name).To(Equal("loki"))
					Expect(chart).To(Equal("loki"))
					Expect(namespace).To(Equal(wego.DefaultNamespace))
					Expect(targetNamespace).To(Equal("sock-shop"))
				})
			})
		})
	})
})

var _ = Describe("Test app hash", func() {

	It("should return right hash for a helm app", func() {

		wegoapp := wego.Application{
			Spec: wego.ApplicationSpec{
				Branch:         "main",
				URL:            "https://github.com/owner/repo1",
				DeploymentType: wego.DeploymentTypeHelm,
			},
		}
		wegoapp.Name = "nginx"

		app, err := WegoAppToApp(wegoapp)
		Expect(err).ToNot(HaveOccurred())

		appHash := GetAppHash(app)
		expectedHash := getHash(app.GitSourceURL.String(), app.Name, app.Branch)

		Expect(appHash).To(Equal("wego-" + expectedHash))

	})

	It("should return right hash for a kustomize app", func() {
		wegoapp := wego.Application{
			Spec: wego.ApplicationSpec{
				Branch:         "main",
				URL:            "https://github.com/owner/repo1",
				Path:           "custompath",
				DeploymentType: wego.DeploymentTypeKustomize,
			},
		}

		app, err := WegoAppToApp(wegoapp)
		Expect(err).ToNot(HaveOccurred())

		appHash := GetAppHash(app)
		expectedHash := getHash(app.GitSourceURL.String(), app.Path, app.Branch)

		Expect(appHash).To(Equal("wego-" + expectedHash))

	})
})

func getHash(inputs ...string) string {
	final := []byte(strings.Join(inputs, ""))
	return fmt.Sprintf("%x", md5.Sum(final))
}
