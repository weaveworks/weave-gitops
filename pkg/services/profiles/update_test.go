package profiles_test

import (
	"context"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/services/profiles"
	"sigs.k8s.io/yaml"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/testing"
)

var updateOptions profiles.UpdateOptions

var _ = Describe("Update Profile(s)", func() {
	var (
		gitProviders *gitprovidersfakes.FakeGitProvider
		profilesSvc  *profiles.ProfilesSvc
		clientSet    *fake.Clientset
		fakeLogger   *loggerfakes.FakeLogger
		// fakePR       *fakegitprovider.PullRequest
	)

	BeforeEach(func() {
		gitProviders = &gitprovidersfakes.FakeGitProvider{}
		clientSet = fake.NewSimpleClientset()
		fakeLogger = &loggerfakes.FakeLogger{}
		// fakePR = &fakegitprovider.PullRequest{}
		profilesSvc = profiles.NewService(clientSet, fakeLogger)

		updateOptions = profiles.UpdateOptions{
			ConfigRepo: "ssh://git@github.com/owner/config-repo.git",
			Name:       "podinfo",
			Cluster:    "prod",
			Namespace:  "weave-system",
			Version:    "latest",
		}
	})

	When("the config repository exists", func() {
		When("the version and HelmRepository name and namespace were discovered", func() {
			When("the file containing the HelmReleases is empty", func() {
				It("returns an error", func() {
					gitProviders.RepositoryExistsReturns(true, nil)
					gitProviders.GetDefaultBranchReturns("main", nil)
					gitProviders.GetRepoDirFilesReturns(makeTestFiles(), nil)
					clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
						return true, newFakeResponseWrapper(getProfilesResp), nil
					})
					err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
					Expect(err).To(MatchError("failed to find installed profiles in '.weave-gitops/clusters/prod/system/profiles.yaml' of config repo \"ssh://git@github.com/owner/config-repo.git\""))
					Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
					Expect(gitProviders.GetRepoDirFilesCallCount()).To(Equal(1))
				})
			})

			When("the file containing the HelmReleases does not contain a HelmRelease with the given name and namespace", func() {
				It("returns an error", func() {
					gitProviders.RepositoryExistsReturns(true, nil)
					gitProviders.GetDefaultBranchReturns("main", nil)
					existingRelease := helm.MakeHelmRelease(
						"random-profile", "6.0.1", "prod", "weave-system",
						types.NamespacedName{Name: "helm-repo-name", Namespace: "helm-repo-namespace"},
					)
					r, _ := yaml.Marshal(existingRelease)
					content := string(r)
					path := git.GetProfilesPath("prod", models.WegoProfilesPath)
					gitProviders.GetRepoDirFilesReturns([]*gitprovider.CommitFile{{
						Path:    &path,
						Content: &content,
					}}, nil)
					clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
						return true, newFakeResponseWrapper(getProfilesResp), nil
					})
					err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
					Expect(err).To(MatchError("failed to update HelmRelease for profile 'podinfo' in profiles.yaml: profile 'podinfo' could not be found in weave-system/prod"))
					Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
					Expect(gitProviders.GetRepoDirFilesCallCount()).To(Equal(1))
				})
			})
		})

		Context("it fails to discover the HelmRepository name and namespace", func() {
			It("fails if it's unable to list available profiles from the cluster", func() {
				gitProviders.RepositoryExistsReturns(true, nil)
				clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
					return true, newFakeResponseWrapperWithErr("nope"), nil
				})
				err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
				Expect(err).To(MatchError("failed to discover HelmRepository: failed to get profiles from cluster: failed to make GET request to service weave-system/wego-app path \"/v1/profiles\": nope"))
				Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
			})

			It("fails to find an available profile with the given version", func() {
				gitProviders.RepositoryExistsReturns(true, nil)
				clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
					return true, newFakeResponseWrapper(getProfilesResp), nil
				})
				updateOptions.Version = "7.0.0"
				err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
				Expect(err).To(MatchError("failed to discover HelmRepository: failed to get profiles from cluster: version '7.0.0' not found for profile 'podinfo' in prod/weave-system"))
				Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
			})
		})
	})

	When("the config repository does not exist", func() {
		It("fails if the --config-repo url format is wrong", func() {
			updateOptions = profiles.UpdateOptions{
				Name:       "foo",
				ConfigRepo: "{http:/-*wrong-url-827",
				Cluster:    "prod",
			}

			err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
			Expect(err).To(MatchError("failed to parse url: could not get provider name from URL {http:/-*wrong-url-827: could not parse git repo url \"{http:/-*wrong-url-827\": parse \"{http:/-*wrong-url-827\": first path segment in URL cannot contain colon"))
		})

		It("fails if the config repo does not exist", func() {
			gitProviders.RepositoryExistsReturns(false, nil)
			err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
			Expect(err).To(MatchError("repository \"ssh://git@github.com/owner/config-repo.git\" could not be found"))
			Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
		})
	})
})
