package install

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/fluxcd/go-git-providers/gitprovider"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/yaml"

	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/services/gitopswriter"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
		fakeKubeClient.FetchNamespaceWithLabelReturns(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: testNamespace}}, nil)
		fakeGitClient = &gitfakes.FakeGit{}
		fakeGitProvider = &gitprovidersfakes.FakeGitProvider{}
		log = &loggerfakes.FakeLogger{}
		repoWriter = gitopswriter.NewRepoWriter(log, fakeGitClient, fakeGitProvider)
		installer = NewInstaller(fakeFluxClient, fakeKubeClient, fakeGitClient, fakeGitProvider, log, repoWriter)
	})

	Context("error paths", func() {
		someError := errors.New("some error")

		BeforeEach(func() {
			fakeKubeClient.GetClusterStatusReturns(kube.Unmodified)
			fakeKubeClient.GetWegoConfigReturns(&kube.WegoConfig{
				ConfigRepo:    configRepo.String(),
				FluxNamespace: testNamespace,
				WegoNamespace: testNamespace,
			}, nil)
			fakeKubeClient.GetClusterNameReturns(clusterName, nil)

			fakeFluxClient.InstallReturnsOnCall(0, nil, nil)
			fakeFluxClient.InstallReturnsOnCall(1, nil, nil)

			fakeGitProvider.GetDefaultBranchReturnsOnCall(0, "main", nil)
			fakeGitProvider.GetDefaultBranchReturnsOnCall(1, "main", nil)

			privateVisibility := gitprovider.RepositoryVisibilityPrivate
			fakeGitProvider.GetRepoVisibilityReturns(&privateVisibility, nil)

			fakeKubeClient.ApplyReturns(nil)

			fakeFluxClient.InstallReturnsOnCall(2, nil, nil)
		})

		It("should fail validating wego installation", func() {
			fakeKubeClient.GetClusterStatusReturns(kube.Unknown)

			err := installer.Install(testNamespace, configRepo, true)
			Expect(err).Should(MatchError("failed validating wego installation: Weave GitOps cannot talk to the cluster"))
		})

		It("should fail getting cluster name", func() {
			fakeKubeClient.GetClusterNameReturns("", someError)

			err := installer.Install(testNamespace, configRepo, true)
			Expect(err).Should(MatchError(fmt.Sprintf("failed getting cluster name: %s", someError)))
		})

		It("should fail installing flux", func() {

			fakeFluxClient.InstallReturnsOnCall(0, nil, someError)

			err := installer.Install(testNamespace, configRepo, true)
			Expect(err).Should(MatchError(fmt.Sprintf("failed installing flux: %s", someError)))
		})

		It("should fail getting bootstrap manifests", func() {
			fakeFluxClient.InstallReturnsOnCall(1, nil, someError)

			err := installer.Install(testNamespace, configRepo, true)
			Expect(err).Should(MatchError(fmt.Sprintf("failed getting bootstrap manifests: failed getting runtime manifests: %s", someError)))
		})

		It("should fail getting default branch", func() {
			fakeGitProvider.GetDefaultBranchReturnsOnCall(1, "", someError)

			err := installer.Install(testNamespace, configRepo, true)
			Expect(err).Should(MatchError(fmt.Sprintf("failed getting default branch: %s", someError)))
		})

		It("should fail applying bootstrap manifests", func() {
			fakeKubeClient.ApplyReturns(someError)

			err := installer.Install(testNamespace, configRepo, true)
			Expect(err).Should(MatchError(fmt.Sprintf("error applying manifest .weave-gitops/clusters/test-cluster/system/wego-system.yaml: %s", someError)))
		})

		It("should fail getting gitops manifests", func() {
			fakeFluxClient.InstallReturnsOnCall(1, nil, someError)

			err := installer.Install(testNamespace, configRepo, true)
			Expect(err.Error()).Should(ContainSubstring(fmt.Sprintf("failed getting runtime manifests: %s", someError)))
		})

		It("should fail writing directly to branch", func() {

			fakeGitClient.CloneReturns(false, someError)

			err := installer.Install(testNamespace, configRepo, true)
			Expect(err).Should(MatchError(fmt.Sprintf("failed writting to default branch failed to clone repo: failed cloning user repo: ssh://git@github.com/test-user/test-repo.git: %s", someError)))
		})

		It("should fail creating a pull requests", func() {
			fakeGitProvider.CreatePullRequestReturns(nil, someError)

			err := installer.Install(testNamespace, configRepo, false)
			Expect(err).Should(MatchError(fmt.Sprintf("failed creating pull request: %s", someError)))
		})
	})

	Context("success path", func() {

		var wegoAppManifest []byte
		var source models.Manifest
		var wegoConfigManifest []byte
		var systemKustomizationManifest []byte
		var runtimeManifests []byte
		var systemKustomizationResource []byte
		var userKustomizationResource []byte
		BeforeEach(func() {
			fakeKubeClient.GetClusterStatusReturns(kube.Unmodified)
			fakeKubeClient.GetWegoConfigReturns(&kube.WegoConfig{
				ConfigRepo:    configRepo.String(),
				FluxNamespace: testNamespace,
				WegoNamespace: testNamespace,
			}, nil)
			fakeKubeClient.GetClusterNameReturns(clusterName, nil)

			fakeFluxClient.InstallReturns(nil, nil)

			fakeKubeClient.ApplyReturns(nil)

			privateVisibility := gitprovider.RepositoryVisibilityPrivate
			fakeGitProvider.GetRepoVisibilityReturns(&privateVisibility, nil)

			fakeGitProvider.GetDefaultBranchReturns("main", nil)

			runtimeManifests = []byte("runtime-manifests")
			fakeFluxClient.InstallReturns(runtimeManifests, nil)

			wegoAppManifests, err := manifests.GenerateWegoAppManifests(manifests.Params{AppImage: models.WegoImage, AppVersion: "v0.0.0", Namespace: testNamespace})
			Expect(err).ShouldNot(HaveOccurred())

			wegoAppManifest = bytes.Join(wegoAppManifests, []byte("---\n"))

			systemKustomizationResource = []byte("system kustomization resource")
			fakeFluxClient.CreateKustomizationReturnsOnCall(0, systemKustomizationResource, nil)
			userKustomizationResource = []byte("user kustomization resource")
			fakeFluxClient.CreateKustomizationReturnsOnCall(1, userKustomizationResource, nil)

			fakeFluxClient.CreateKustomizationReturnsOnCall(2, systemKustomizationResource, nil)
			fakeFluxClient.CreateKustomizationReturnsOnCall(3, userKustomizationResource, nil)

			gitopsConfigMap, err := models.CreateGitopsConfigMap(testNamespace, testNamespace, configRepo.String())
			Expect(err).ShouldNot(HaveOccurred())

			wegoConfigManifest, err = yaml.Marshal(gitopsConfigMap)
			Expect(err).ShouldNot(HaveOccurred())

			systemKustomization := models.CreateKustomization(clusterName, testNamespace, models.RuntimePath, models.SourcePath, models.SystemKustResourcePath, models.UserKustResourcePath, models.WegoAppPath, models.WegoProfilesPath)

			systemKustomizationManifest, err = yaml.Marshal(systemKustomization)
			Expect(err).ShouldNot(HaveOccurred())

			source, err = models.GetSourceManifest(context.Background(), fakeFluxClient, fakeGitProvider, clusterName, testNamespace, configRepo, "main")
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should succeed with auto-merge=true", func() {

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
				source,
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.SystemKustomizationPath),
					Content: systemKustomizationManifest,
				},
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.WegoProfilesPath),
					Content: []byte(""),
				},
				{
					Path:    filepath.Join(git.GetUserPath(clusterName), ".keep"),
					Content: strconv.AppendQuote(nil, "# keep"),
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
				Expect(path).Should(Equal(expectedManifests[writeIndex].Path))
				Expect(string(content)).Should(Equal(string(expectedManifests[writeIndex].Content)))
				writeIndex++
				return nil
			})

			err = installer.Install(testNamespace, configRepo, true)
			Expect(err).ShouldNot(HaveOccurred())
		})
		It("should succeed with auto-merge=false", func() {

			fakeGitProvider.CreatePullRequestReturns(nil, nil)

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
				source,
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.SystemKustomizationPath),
					Content: systemKustomizationManifest,
				},
				{
					Path:    git.GetSystemQualifiedPath(clusterName, models.WegoProfilesPath),
					Content: []byte(""),
				},
				{
					Path:    filepath.Join(git.GetUserPath(clusterName), ".keep"),
					Content: strconv.AppendQuote(nil, "# keep"),
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
