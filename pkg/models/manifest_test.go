package models

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"strconv"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"

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
	var params = BootstrapManifestsParams{
		ClusterName:   "test-cluster",
		WegoNamespace: "test-namespace",
		FluxNamespace: "test-namespace",
	}
	var err error
	var _ = BeforeEach(func() {
		fakeFluxClient = &fluxfakes.FakeFlux{}
		params.ConfigRepo, err = gitproviders.NewRepoURL("ssh://git@github.com/test-user/test-repo", true)
		const clusterName = "test-cluster"
		const testNamespace = "test-namespace"

		Context("BootstrapManifests", func() {

			Context("error paths", func() {
				someError := errors.New("some error")
				It("should fail getting runtime manifests", func() {

					fakeFluxClient.InstallReturns(nil, someError)

					_, err = BootstrapManifests(fakeFluxClient, params)
					Expect(err.Error()).Should(ContainSubstring(someError.Error()))
				})

				It("should fail creating system resource kustomization", func() {

					fakeFluxClient.InstallReturns(nil, nil)

					fakeFluxClient.CreateKustomizationReturns(nil, someError)

					_, err = BootstrapManifests(fakeFluxClient, params)
					Expect(err.Error()).Should(ContainSubstring(someError.Error()))
				})

				It("should fail creating user resource kustomization", func() {

					fakeFluxClient.InstallReturns(nil, nil)

					fakeFluxClient.CreateKustomizationReturnsOnCall(0, nil, nil)
					fakeFluxClient.CreateKustomizationReturnsOnCall(1, nil, someError)

					_, err = BootstrapManifests(fakeFluxClient, params)
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

					gitopsConfigMap, err := CreateGitopsConfigMap(params.FluxNamespace, params.WegoNamespace, params.ConfigRepo.String())
					Expect(err).ShouldNot(HaveOccurred())

					wegoConfigManifest, err := yaml.Marshal(gitopsConfigMap)
					Expect(err).ShouldNot(HaveOccurred())

					manifestsFiles, err := BootstrapManifests(fakeFluxClient, params)
					Expect(err).ShouldNot(HaveOccurred())
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
					}

					for ind, manifest := range manifestsFiles {
						Expect(manifest.Path).Should(Equal(expectedManifests[ind].Path))
						Expect(string(manifest.Content)).Should(Equal(string(expectedManifests[ind].Content)))
					}

				})
			})
		})

		Context("GitopsManifests", func() {
			var ctx context.Context
			var boostrapManifests []Manifest
			var params GitopsManifestsParams
			BeforeEach(func() {
				ctx = context.Background()
				fakeFluxClient = &fluxfakes.FakeFlux{}
				fakeGitProvider = &gitprovidersfakes.FakeGitProvider{}
				fakeGitProvider.GetRepoVisibilityReturns(gitprovider.RepositoryVisibilityVar(gitprovider.RepositoryVisibilityPublic), nil)

				boostrapManifests = []Manifest{}

				configRepo, err := gitproviders.NewRepoURL("ssh://git@github.com/test-user/test-repo", true)
				Expect(err).ShouldNot(HaveOccurred())

				params = GitopsManifestsParams{
					FluxClient:    fakeFluxClient,
					GitProvider:   fakeGitProvider,
					ClusterName:   "test-cluster",
					WegoNamespace: "test-namespace",
					ConfigRepo:    configRepo,
				}
			})
			Context("error paths", func() {
				someError := errors.New("some error")
				It("should fail getting runtime manifests", func() {
					fakeGitProvider.GetDefaultBranchReturns("", someError)

					_, err = GitopsManifests(ctx, boostrapManifests, params)
					Expect(err.Error()).Should(ContainSubstring(someError.Error()))
				})

				It("should fail getting secret name for private git source", func() {
					fakeFluxClient.InstallReturns(nil, nil)

					fakeGitProvider.GetRepoVisibilityReturns(nil, someError)

					_, err = GitopsManifests(ctx, boostrapManifests, params)
					Expect(err.Error()).Should(ContainSubstring(someError.Error()))
				})

				It("should fail getting secret name for private git source", func() {
					fakeFluxClient.InstallReturns(nil, nil)

					privateVisibility := gitprovider.RepositoryVisibilityPrivate
					fakeGitProvider.GetRepoVisibilityReturns(&privateVisibility, nil)

					fakeGitProvider.GetDefaultBranchReturns("", someError)

					_, err = GitopsManifests(ctx, boostrapManifests, params)
					Expect(err.Error()).Should(ContainSubstring(someError.Error()))
				})

				It("should fail creating flux source", func() {
					fakeFluxClient.InstallReturns(nil, nil)

					privateVisibility := gitprovider.RepositoryVisibilityPrivate
					fakeGitProvider.GetRepoVisibilityReturns(&privateVisibility, nil)

					fakeGitProvider.GetDefaultBranchReturns("main", nil)

					fakeFluxClient.CreateSourceGitReturns(nil, someError)

					_, err = GitopsManifests(ctx, boostrapManifests, params)
					Expect(err.Error()).Should(ContainSubstring(someError.Error()))
				})
			})
			Context("success case", func() {
				It("should pass successfully", func() {
					runtimeManifests := []byte("flux runtime content")
					fakeFluxClient.InstallReturns(runtimeManifests, nil)

					privateVisibility := gitprovider.RepositoryVisibilityPrivate
					fakeGitProvider.GetRepoVisibilityReturns(&privateVisibility, nil)

					fakeGitProvider.GetDefaultBranchReturns("main", nil)

					systemKustomization := CreateKustomization(params.ClusterName, params.WegoNamespace, RuntimePath, SourcePath, SystemKustResourcePath, UserKustResourcePath, WegoAppPath)

					systemKustomizationManifest, err := yaml.Marshal(systemKustomization)
					Expect(err).ShouldNot(HaveOccurred())

					gitSource := []byte("git source")
					fakeFluxClient.CreateSourceGitReturns(gitSource, nil)

					manifestsFiles, err := GitopsManifests(ctx, boostrapManifests, params)
					Expect(err).ShouldNot(HaveOccurred())

					Expect(len(manifestsFiles)).Should(Equal(len(boostrapManifests) + 3))

					expectedManifests := []Manifest{
						{
							Path:    git.GetSystemQualifiedPath(params.ClusterName, SourcePath),
							Content: gitSource,
						},
						{
							Path:    git.GetSystemQualifiedPath(params.ClusterName, SystemKustomizationPath),
							Content: systemKustomizationManifest,
						},
						{
							Path:    filepath.Join(git.GetUserPath(params.ClusterName), ".keep"),
							Content: strconv.AppendQuote(nil, "# keep"),
						},
					}

					for ind, manifest := range manifestsFiles {
						Expect(manifest.Path).Should(Equal(expectedManifests[ind].Path))
						Expect(string(manifest.Content)).Should(Equal(string(expectedManifests[ind].Content)))
					}
				})
			})
		})
	})
})
