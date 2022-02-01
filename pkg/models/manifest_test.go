package models

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/manifests"
	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"sigs.k8s.io/yaml"
)

var _ = Describe("Installer", func() {

	var fakeFluxClient *fluxfakes.FakeFlux
	var fakeGitProvider *gitprovidersfakes.FakeGitProvider
	var fakeKubeClient *kubefakes.FakeKube
	var params = ManifestsParams{
		ClusterName:   "test-cluster",
		WegoNamespace: "test-namespace",
	}
	var err error
	var ctx context.Context
	var _ = BeforeEach(func() {
		params.ConfigRepo, err = gitproviders.NewRepoURL("ssh://git@github.com/test-user/test-repo")
		ctx = context.Background()

		fakeFluxClient = &fluxfakes.FakeFlux{}
		fakeGitProvider = &gitprovidersfakes.FakeGitProvider{}
		params.ConfigRepo, err = gitproviders.NewRepoURL("ssh://git@github.com/test-user/test-repo")

		Expect(err).ShouldNot(HaveOccurred())
		fakeKubeClient = &kubefakes.FakeKube{}
		fakeKubeClient.FetchNamespaceWithLabelReturns(&v1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: params.WegoNamespace}}, nil)
	})

	Context("BootstrapManifests", func() {
		Context("error paths", func() {
			someError := errors.New("some error")

			It("should fail getting runtime manifests", func() {
				fakeFluxClient.InstallReturns(nil, someError)

				_, err = BootstrapManifests(ctx, fakeFluxClient, fakeGitProvider, fakeKubeClient, params)
				Expect(err.Error()).Should(ContainSubstring(someError.Error()))
			})

			It("should fail creating system resource kustomization", func() {

				fakeFluxClient.InstallReturns(nil, nil)

				fakeFluxClient.CreateKustomizationReturns(nil, someError)

				_, err = BootstrapManifests(ctx, fakeFluxClient, fakeGitProvider, fakeKubeClient, params)
				Expect(err.Error()).Should(ContainSubstring(someError.Error()))
			})

			It("should fail creating user resource kustomization", func() {

				fakeFluxClient.InstallReturns(nil, nil)

				fakeFluxClient.CreateKustomizationReturnsOnCall(0, nil, nil)
				fakeFluxClient.CreateKustomizationReturnsOnCall(1, nil, someError)

				_, err = BootstrapManifests(ctx, fakeFluxClient, fakeGitProvider, fakeKubeClient, params)
				Expect(err.Error()).Should(ContainSubstring(someError.Error()))
			})
		})
		Context("success case", func() {
			It("should pass successfully", func() {

				runtimeManifests := []byte("flux runtime content")
				fakeFluxClient.InstallReturns(runtimeManifests, nil)

				systemKustomizationResourceManifest := []byte("system kustomization resource")
				fakeFluxClient.CreateKustomizationReturnsOnCall(0, systemKustomizationResourceManifest, nil)

				userKustomizationResourceManifest := []byte("user kustomization resource")
				fakeFluxClient.CreateKustomizationReturnsOnCall(1, userKustomizationResourceManifest, nil)

				wegoAppManifests, err := manifests.GenerateWegoAppManifests(manifests.Params{AppVersion: "v0.0.0", Namespace: params.WegoNamespace})
				Expect(err).ShouldNot(HaveOccurred())

				wegoAppManifest := bytes.Join(wegoAppManifests, []byte("---\n"))

				gitopsConfigMap, err := CreateGitopsConfigMap(params.WegoNamespace, params.WegoNamespace, params.ConfigRepo.String())
				Expect(err).ShouldNot(HaveOccurred())

				wegoConfigManifest, err := yaml.Marshal(gitopsConfigMap)
				Expect(err).ShouldNot(HaveOccurred())

				gitSource := []byte("git source")
				fakeFluxClient.CreateSourceGitReturns(gitSource, nil)

				fakeGitProvider.GetDefaultBranchReturns("main", nil)

				privateVisibility := gitprovider.RepositoryVisibilityPrivate
				fakeGitProvider.GetRepoVisibilityReturns(&privateVisibility, nil)

				manifestsFiles, err := BootstrapManifests(ctx, fakeFluxClient, fakeGitProvider, fakeKubeClient, params)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(len(manifestsFiles)).Should(Equal(7))

				expectedManifests := []Manifest{
					{
						Path:    git.GetSystemQualifiedPath(params.ClusterName, AppCRDPath),
						Content: manifests.AppCRD,
					},
					{
						Path:    git.GetSystemQualifiedPath(params.ClusterName, RuntimePath),
						Content: runtimeManifests,
					},
					{
						Path:    git.GetSystemQualifiedPath(params.ClusterName, SystemKustResourcePath),
						Content: systemKustomizationResourceManifest,
					},
					{
						Path:    git.GetSystemQualifiedPath(params.ClusterName, UserKustResourcePath),
						Content: userKustomizationResourceManifest,
					},
					{
						Path:    git.GetSystemQualifiedPath(params.ClusterName, WegoAppPath),
						Content: wegoAppManifest,
					},
					{
						Path:    git.GetSystemQualifiedPath(params.ClusterName, WegoConfigPath),
						Content: wegoConfigManifest,
					},
					{
						Path:    git.GetSystemQualifiedPath(params.ClusterName, SourcePath),
						Content: gitSource,
					},
				}

				for ind, manifest := range manifestsFiles {
					Expect(manifest.Path).Should(Equal(expectedManifests[ind].Path))
					Expect(string(manifest.Content)).Should(Equal(string(expectedManifests[ind].Content)))
				}

			})
		})
	})

	Context("NoClusterApplicableManifests", func() {
		BeforeEach(func() {

			configRepo, err := gitproviders.NewRepoURL("ssh://git@github.com/test-user/test-repo")
			Expect(err).ShouldNot(HaveOccurred())

			params = ManifestsParams{
				ClusterName:   "test-cluster",
				WegoNamespace: "test-namespace",
				ConfigRepo:    configRepo,
			}
		})

		Context("success case", func() {
			It("should pass successfully", func() {
				systemKustomization := CreateKustomization(params.ClusterName, params.WegoNamespace, RuntimePath, SourcePath, SystemKustResourcePath, UserKustResourcePath, WegoAppPath)

				systemKustomizationManifest, err := yaml.Marshal(systemKustomization)
				Expect(err).ShouldNot(HaveOccurred())

				noClusterApplicableManifests, err := NoClusterApplicableManifests(params)
				Expect(err).ShouldNot(HaveOccurred())

				Expect(len(noClusterApplicableManifests)).Should(Equal(2))

				expectedManifests := []Manifest{
					{
						Path:    git.GetSystemQualifiedPath(params.ClusterName, SystemKustomizationPath),
						Content: systemKustomizationManifest,
					},
					{
						Path:    filepath.Join(git.GetUserPath(params.ClusterName), ".keep"),
						Content: strconv.AppendQuote(nil, "# keep"),
					},
				}

				for ind, manifest := range noClusterApplicableManifests {
					Expect(manifest.Path).Should(Equal(expectedManifests[ind].Path))
					Expect(string(manifest.Content)).Should(Equal(string(expectedManifests[ind].Content)))
				}
			})
		})

		Context("Validate name", func() {
			It("should pass successfully", func() {
				Expect(ValidateApplicationName("foobar")).ShouldNot(HaveOccurred())
				Expect(ValidateApplicationName("foobar-1234-test-bar-0123456")).ShouldNot(HaveOccurred())
				Expect(ValidateApplicationName("f")).ShouldNot(HaveOccurred())
				Expect(ValidateApplicationName("6")).ShouldNot(HaveOccurred())
				Expect(ValidateApplicationName(strings.Repeat("1", 63))).ShouldNot(HaveOccurred())
			})
			It("should fail", func() {
				Expect(ValidateApplicationName("Special")).Should(HaveOccurred())
				Expect(ValidateApplicationName("foobar.baz")).Should(HaveOccurred())
				Expect(ValidateApplicationName(strings.Repeat("1", 64))).Should(HaveOccurred())
			})
		})
	})
})
