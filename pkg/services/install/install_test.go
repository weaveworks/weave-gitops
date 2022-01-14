package install

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"sigs.k8s.io/yaml"

	"github.com/weaveworks/weave-gitops/pkg/kube"

	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"

	"github.com/fluxcd/go-git-providers/gitprovider"

	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"

	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"

	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"

	"github.com/weaveworks/weave-gitops/pkg/services/gitopswriter"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Installer", func() {

	var installer Installer
	var fakeFluxClient *fluxfakes.FakeFlux
	var fakeKubeClient *kubefakes.FakeKube
	var fakeGitClient *gitfakes.FakeGit
	var fakeGitProvider *gitprovidersfakes.FakeGitProvider
	var repoWriter gitopswriter.RepoWriter
	var log logger.Logger
	var testNamespace string
	var configRepo gitproviders.RepoURL
	var err error
	const clusterName = "test-cluster"
	var _ = BeforeEach(func() {
		testNamespace = "test-namespace"
		configRepo, err = gitproviders.NewRepoURL("ssh://git@github.com/test-user/test-repo")
		Expect(err).ShouldNot(HaveOccurred())
		fakeFluxClient = &fluxfakes.FakeFlux{}
		fakeKubeClient = &kubefakes.FakeKube{}
		fakeGitClient = &gitfakes.FakeGit{}
		fakeGitProvider = &gitprovidersfakes.FakeGitProvider{}
		log = &loggerfakes.FakeLogger{}
		repoWriter = gitopswriter.NewRepoWriter(log, fakeGitClient, fakeGitProvider)
		installer = NewInstaller(fakeFluxClient, fakeKubeClient, fakeGitClient, fakeGitProvider, log, repoWriter)
	})

	// Should I include more specific error messages matches
	// Or maybe create a template of the errors and reuse it here
	Context("error paths", func() {
		someError := errors.New("some error")

		It("should fail validating wego installation", func() {
			fakeKubeClient.GetClusterStatusReturns(kube.Unknown)

			err := installer.Install(testNamespace, configRepo, true)
			Expect(err).Should(HaveOccurred())
		})

		It("should fail getting cluster name", func() {
			fakeKubeClient.GetClusterStatusReturns(kube.Unmodified)
			fakeKubeClient.GetWegoConfigReturns(&kube.WegoConfig{
				FluxNamespace: testNamespace,
				WegoNamespace: testNamespace,
			}, nil)
			fakeKubeClient.GetClusterNameReturns("", someError)

			err := installer.Install(testNamespace, configRepo, true)
			Expect(err.Error()).Should(ContainSubstring(someError.Error()))
		})

		It("should fail installing flux", func() {
			fakeKubeClient.GetClusterStatusReturns(kube.Unmodified)
			fakeKubeClient.GetWegoConfigReturns(&kube.WegoConfig{
				FluxNamespace: testNamespace,
				WegoNamespace: testNamespace,
			}, nil)
			fakeKubeClient.GetClusterNameReturns(testNamespace, nil)

			fakeFluxClient.InstallReturns(nil, someError)

			err := installer.Install(testNamespace, configRepo, true)
			Expect(err.Error()).Should(ContainSubstring(someError.Error()))
		})

		It("should fail getting bootstrap manifests", func() {
			fakeKubeClient.GetClusterStatusReturns(kube.Unmodified)
			fakeKubeClient.GetWegoConfigReturns(&kube.WegoConfig{
				FluxNamespace: testNamespace,
				WegoNamespace: testNamespace,
			}, nil)
			fakeKubeClient.GetClusterNameReturns(testNamespace, nil)

			fakeFluxClient.InstallReturnsOnCall(0, nil, nil)
			fakeFluxClient.InstallReturnsOnCall(1, nil, someError)

			err := installer.Install(testNamespace, configRepo, true)
			Expect(err.Error()).Should(ContainSubstring(someError.Error()))
		})

		It("should fail applying bootstrap manifests", func() {
			fakeKubeClient.GetClusterStatusReturns(kube.Unmodified)
			fakeKubeClient.GetWegoConfigReturns(&kube.WegoConfig{
				FluxNamespace: testNamespace,
				WegoNamespace: testNamespace,
			}, nil)
			fakeKubeClient.GetClusterNameReturns(testNamespace, nil)

			fakeFluxClient.InstallReturns(nil, nil)

			fakeKubeClient.ApplyReturns(someError)

			err := installer.Install(testNamespace, configRepo, true)
			Expect(err.Error()).Should(ContainSubstring(someError.Error()))
		})

		It("should fail getting gitops manifests", func() {
			fakeKubeClient.GetClusterStatusReturns(kube.Unmodified)
			fakeKubeClient.GetWegoConfigReturns(&kube.WegoConfig{
				FluxNamespace: testNamespace,
				WegoNamespace: testNamespace,
			}, nil)
			fakeKubeClient.GetClusterNameReturns(testNamespace, nil)

			fakeFluxClient.InstallReturnsOnCall(0, nil, nil)
			fakeFluxClient.InstallReturnsOnCall(1, nil, nil)

			fakeKubeClient.ApplyReturns(nil)

			fakeFluxClient.InstallReturnsOnCall(2, nil, someError)

			err := installer.Install(testNamespace, configRepo, true)
			Expect(err.Error()).Should(ContainSubstring(someError.Error()))
		})

		It("should fail getting default branch", func() {
			fakeKubeClient.GetClusterStatusReturns(kube.Unmodified)
			fakeKubeClient.GetWegoConfigReturns(&kube.WegoConfig{
				FluxNamespace: testNamespace,
				WegoNamespace: testNamespace,
			}, nil)
			fakeKubeClient.GetClusterNameReturns(testNamespace, nil)

			fakeFluxClient.InstallReturns(nil, nil)

			fakeKubeClient.ApplyReturns(nil)

			privateVisibility := gitprovider.RepositoryVisibilityPrivate
			fakeGitProvider.GetRepoVisibilityReturns(&privateVisibility, nil)

			fakeGitProvider.GetDefaultBranchReturnsOnCall(0, "main", nil)

			fakeGitProvider.GetDefaultBranchReturnsOnCall(1, "", someError)

			err := installer.Install(testNamespace, configRepo, true)
			Expect(err.Error()).Should(ContainSubstring(someError.Error()))
		})

		It("should fail writing directly to branch", func() {
			fakeKubeClient.GetClusterStatusReturns(kube.Unmodified)
			fakeKubeClient.GetWegoConfigReturns(&kube.WegoConfig{
				FluxNamespace: testNamespace,
				WegoNamespace: testNamespace,
			}, nil)
			fakeKubeClient.GetClusterNameReturns(testNamespace, nil)

			fakeFluxClient.InstallReturns(nil, nil)

			fakeKubeClient.ApplyReturns(nil)

			privateVisibility := gitprovider.RepositoryVisibilityPrivate
			fakeGitProvider.GetRepoVisibilityReturns(&privateVisibility, nil)

			fakeGitProvider.GetDefaultBranchReturns("main", nil)

			fakeGitClient.CloneReturns(false, someError)

			err := installer.Install(testNamespace, configRepo, true)
			Expect(err.Error()).Should(ContainSubstring(someError.Error()))
		})

		It("should fail creating a pull requests", func() {
			fakeKubeClient.GetClusterStatusReturns(kube.Unmodified)
			fakeKubeClient.GetWegoConfigReturns(&kube.WegoConfig{
				FluxNamespace: testNamespace,
				WegoNamespace: testNamespace,
			}, nil)
			fakeKubeClient.GetClusterNameReturns(testNamespace, nil)

			fakeFluxClient.InstallReturns(nil, nil)

			fakeKubeClient.ApplyReturns(nil)

			privateVisibility := gitprovider.RepositoryVisibilityPrivate
			fakeGitProvider.GetRepoVisibilityReturns(&privateVisibility, nil)

			fakeGitProvider.GetDefaultBranchReturns("main", nil)

			fakeGitProvider.CreatePullRequestReturns(nil, someError)

			err := installer.Install(testNamespace, configRepo, false)
			Expect(err.Error()).Should(ContainSubstring(someError.Error()))
		})
	})
	Context("success path", func() {
		It("should succeed with auto-merge=true", func() {
			fakeKubeClient.GetClusterStatusReturns(kube.Unmodified)
			fakeKubeClient.GetWegoConfigReturns(&kube.WegoConfig{
				FluxNamespace: testNamespace,
				WegoNamespace: testNamespace,
			}, nil)

			fakeKubeClient.GetClusterNameReturns(clusterName, nil)

			fakeFluxClient.InstallReturns(nil, nil)

			fakeKubeClient.ApplyReturns(nil)

			privateVisibility := gitprovider.RepositoryVisibilityPrivate
			fakeGitProvider.GetRepoVisibilityReturns(&privateVisibility, nil)

			fakeGitProvider.GetDefaultBranchReturns("main", nil)

			fakeGitProvider.CreatePullRequestReturns(nil, nil)

			runtimeManifests := []byte("runtime-manifests")
			fakeFluxClient.InstallReturns(runtimeManifests, nil)

			wegoAppManifests, err := manifests.GenerateWegoAppManifests(manifests.Params{AppVersion: "v0.0.0", Namespace: testNamespace})
			Expect(err).ShouldNot(HaveOccurred())

			wegoAppManifest := bytes.Join(wegoAppManifests, []byte("---\n"))

			systemKustomizationResource := []byte("system kustomization resource")
			fakeFluxClient.CreateKustomizationReturnsOnCall(0, systemKustomizationResource, nil)
			userKustomizationResource := []byte("user kustomization resource")
			fakeFluxClient.CreateKustomizationReturnsOnCall(1, userKustomizationResource, nil)

			fakeFluxClient.CreateKustomizationReturnsOnCall(2, systemKustomizationResource, nil)
			fakeFluxClient.CreateKustomizationReturnsOnCall(3, userKustomizationResource, nil)

			gitopsConfigMap, err := models.GitopsConfigMap(testNamespace, testNamespace)
			Expect(err).ShouldNot(HaveOccurred())

			wegoConfigManifest, err := yaml.Marshal(gitopsConfigMap)
			Expect(err).ShouldNot(HaveOccurred())

			systemKustomization := models.CreateKustomize(clusterName, testNamespace, models.RuntimePath, models.SourcePath, models.SystemKustResourcePath, models.UserKustResourcePath)

			systemKustomizationManifest, err := yaml.Marshal(systemKustomization)
			Expect(err).ShouldNot(HaveOccurred())

			gitSource := []byte("git source")
			fakeFluxClient.CreateSourceGitReturns(gitSource, nil)

			expectedManifests := []models.Manifest{
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.AppCRDPath),
					Content: manifests.AppCRD,
				},
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.RuntimePath),
					Content: runtimeManifests,
				},
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.SystemKustResourcePath),
					Content: systemKustomizationResource,
				},
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.UserKustResourcePath),
					Content: userKustomizationResource,
				},
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.WegoAppPath),
					Content: wegoAppManifest,
				},
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.WegoConfigPath),
					Content: wegoConfigManifest,
				},
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.SystemKustomizationPath),
					Content: systemKustomizationManifest,
				},
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.SourcePath),
					Content: gitSource,
				},
			}

			applyIndex := 0
			fakeKubeClient.ApplyCalls(func(ctx context.Context, manifest []byte, namespace string) error {

				if applyIndex <= 6 {

					Expect(namespace).Should(Equal(namespace))

					if applyIndex != 0 {
						partOfPreviousManifest := bytes.Contains(expectedManifests[applyIndex-1].Content, manifest)
						if partOfPreviousManifest {
							Expect(string(expectedManifests[applyIndex-1].Content)).Should(ContainSubstring(string(manifest)))
							return nil
						}
					}

					Expect(string(expectedManifests[applyIndex].Content)).Should(ContainSubstring(string(manifest)))

					applyIndex++
				}
				return nil
			})

			writeIndex := 0
			fakeGitClient.CloneReturns(true, nil)
			fakeGitClient.WriteCalls(func(path string, content []byte) error {
				if writeIndex < 7 {
					Expect(path).Should(Equal(expectedManifests[writeIndex].Path))
					Expect(string(content)).Should(Equal(string(expectedManifests[writeIndex].Content)))
				}
				writeIndex++
				return nil
			})

			err = installer.Install(testNamespace, configRepo, true)
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should succeed with auto-merge=false", func() {
			fakeKubeClient.GetClusterStatusReturns(kube.Unmodified)
			fakeKubeClient.GetWegoConfigReturns(&kube.WegoConfig{
				FluxNamespace: testNamespace,
				WegoNamespace: testNamespace,
			}, nil)
			fakeKubeClient.GetClusterNameReturns(clusterName, nil)

			fakeFluxClient.InstallReturns(nil, nil)

			fakeKubeClient.ApplyReturns(nil)

			privateVisibility := gitprovider.RepositoryVisibilityPrivate
			fakeGitProvider.GetRepoVisibilityReturns(&privateVisibility, nil)

			fakeGitProvider.GetDefaultBranchReturns("main", nil)

			fakeGitProvider.CreatePullRequestReturns(nil, nil)

			runtimeManifests := []byte("runtime-manifests")
			fakeFluxClient.InstallReturns(runtimeManifests, nil)

			wegoAppManifests, err := manifests.GenerateWegoAppManifests(manifests.Params{AppVersion: "v0.0.0", Namespace: testNamespace})
			Expect(err).ShouldNot(HaveOccurred())

			wegoAppManifest := bytes.Join(wegoAppManifests, []byte("---\n"))

			systemKustomizationResource := []byte("system kustomization resource")
			fakeFluxClient.CreateKustomizationReturnsOnCall(0, systemKustomizationResource, nil)
			userKustomizationResource := []byte("user kustomization resource")
			fakeFluxClient.CreateKustomizationReturnsOnCall(1, userKustomizationResource, nil)

			fakeFluxClient.CreateKustomizationReturnsOnCall(2, systemKustomizationResource, nil)
			fakeFluxClient.CreateKustomizationReturnsOnCall(3, userKustomizationResource, nil)

			gitopsConfigMap, err := models.GitopsConfigMap(testNamespace, testNamespace)
			Expect(err).ShouldNot(HaveOccurred())

			wegoConfigManifest, err := yaml.Marshal(gitopsConfigMap)
			Expect(err).ShouldNot(HaveOccurred())

			systemKustomization := models.CreateKustomize(clusterName, testNamespace, models.RuntimePath, models.SourcePath, models.SystemKustResourcePath, models.UserKustResourcePath)

			systemKustomizationManifest, err := yaml.Marshal(systemKustomization)
			Expect(err).ShouldNot(HaveOccurred())

			gitSource := []byte("git source")
			fakeFluxClient.CreateSourceGitReturns(gitSource, nil)

			expectedManifests := []models.Manifest{
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.AppCRDPath),
					Content: manifests.AppCRD,
				},
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.RuntimePath),
					Content: runtimeManifests,
				},
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.SystemKustResourcePath),
					Content: systemKustomizationResource,
				},
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.UserKustResourcePath),
					Content: userKustomizationResource,
				},
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.WegoAppPath),
					Content: wegoAppManifest,
				},
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.WegoConfigPath),
					Content: wegoConfigManifest,
				},
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.SystemKustomizationPath),
					Content: systemKustomizationManifest,
				},
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.SourcePath),
					Content: gitSource,
				},
			}

			applyIndex := 0
			fakeKubeClient.ApplyCalls(func(ctx context.Context, manifest []byte, namespace string) error {

				if applyIndex <= 6 {

					Expect(namespace).Should(Equal(namespace))

					if applyIndex != 0 {
						partOfPreviousManifest := bytes.Contains(expectedManifests[applyIndex-1].Content, manifest)
						if partOfPreviousManifest {
							Expect(string(expectedManifests[applyIndex-1].Content)).Should(ContainSubstring(string(manifest)))
							return nil
						}
					}

					Expect(string(expectedManifests[applyIndex].Content)).Should(ContainSubstring(string(manifest)))

					applyIndex++
				}
				return nil
			})

			expectedPullRequestInfo := gitproviders.PullRequestInfo{
				Title:         fmt.Sprintf("GitOps associate %s", clusterName),
				Description:   fmt.Sprintf("Add gitops automation manifests for cluster %s", clusterName),
				CommitMessage: gitopswriter.ClusterCommitMessage,
				NewBranch:     models.GetClusterHash(clusterName),
				TargetBranch:  "main",
				Files:         models.ConvertManifestsToCommitFiles(expectedManifests),
			}

			fakeGitProvider.CreatePullRequestCalls(func(ctx context.Context, url gitproviders.RepoURL, info gitproviders.PullRequestInfo) (gitprovider.PullRequest, error) {
				Expect(url).To(Equal(configRepo))
				Expect(info.Title).To(Equal(expectedPullRequestInfo.Title))
				Expect(info.Description).To(Equal(expectedPullRequestInfo.Description))
				Expect(info.CommitMessage).To(Equal(expectedPullRequestInfo.CommitMessage))
				Expect(info.NewBranch).To(Equal(expectedPullRequestInfo.NewBranch))
				Expect(info.TargetBranch).To(Equal(expectedPullRequestInfo.TargetBranch))
				for ind, manifest := range info.Files {
					Expect(*manifest.Path).Should(Equal(expectedManifests[ind].Path))
					Expect(*manifest.Content).Should(Equal(string(expectedManifests[ind].Content)))
				}
				return NewFakePullRequest("test", "test", 1), nil
			})

			err = installer.Install(testNamespace, configRepo, false)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})
})

type fakePullRequest struct {
	pullRequestInfo gitprovider.PullRequestInfo
}

func (fc *fakePullRequest) APIObject() interface{} {
	return &fc.pullRequestInfo
}

func (fc *fakePullRequest) Get() gitprovider.PullRequestInfo {
	return fc.pullRequestInfo
}

func NewFakePullRequest(owner string, repoName string, prNumber int) gitprovider.PullRequest {
	return &fakePullRequest{
		pullRequestInfo: gitprovider.PullRequestInfo{
			WebURL: fmt.Sprintf("https://github.com/%s/%s/pull/%d", owner, repoName, prNumber),
			Merged: false,
			Number: 1,
		},
	}
}
