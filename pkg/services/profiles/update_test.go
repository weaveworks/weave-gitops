package profiles_test

import (
	"context"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/client-go/kubernetes/fake"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/testing"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/services/profiles"
	"github.com/weaveworks/weave-gitops/pkg/vendorfakes/fakegitprovider"
)

var updateOptions profiles.UpdateOptions

var _ = Describe("Update", func() {
	var (
		gitProviders *gitprovidersfakes.FakeGitProvider
		profilesSvc  *profiles.ProfilesSvc
		clientSet    *fake.Clientset
		fakeLogger   *loggerfakes.FakeLogger
		fakePR       *fakegitprovider.PullRequest
	)

	BeforeEach(func() {
		gitProviders = &gitprovidersfakes.FakeGitProvider{}
		clientSet = fake.NewSimpleClientset()
		fakeLogger = &loggerfakes.FakeLogger{}
		fakePR = &fakegitprovider.PullRequest{}
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
			JustBeforeEach(func() {
				gitProviders.RepositoryExistsReturns(true, nil)
				gitProviders.GetDefaultBranchReturns("main", nil)
				gitProviders.GetRepoDirFilesReturns(makeTestFiles(), nil)
				clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
					return true, newFakeResponseWrapper(getProfilesResp), nil
				})
			})

			JustAfterEach(func() {
				Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
				Expect(gitProviders.GetRepoDirFilesCallCount()).To(Equal(1))
				Expect(gitProviders.CreatePullRequestCallCount()).To(Equal(1))
			})

			It("creates a PR to change the version of a Helm Release", func() {
				fakePR.GetReturns(gitprovider.PullRequestInfo{
					WebURL: "url",
				})
				gitProviders.CreatePullRequestReturns(fakePR, nil)
				Expect(profilesSvc.Update(context.TODO(), gitProviders, updateOptions)).Should(Succeed())
			})
		})

		When("it fails to get a list of available profiles from the cluster", func() {
			JustBeforeEach(func() {
				gitProviders.RepositoryExistsReturns(true, nil)
				gitProviders.GetRepoDirFilesReturns(makeTestFiles(), nil)
			})

			It("fails if it's unable to get a matching available profile from the cluster", func() {
				clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
					return true, newFakeResponseWrapperWithErr("nope"), nil
				})
				err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
				Expect(err).To(MatchError("failed to get profiles from cluster: failed to make GET request to service weave-system/wego-app path \"/v1/profiles\": nope"))
				Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
			})

			It("fails if it's unable to discover the HelmRepository's name and namespace values", func() {
				clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
					return true, newFakeResponseWrapper(getRespWithoutHelmRepo()), nil
				})
				err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
				Expect(err).To(MatchError("failed to get profiles from cluster: HelmRepository's name or namespace is empty"))
				Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
			})
		})

		When("it fails to find a matching version", func() {
			It("returns an error", func() {
				gitProviders.RepositoryExistsReturns(true, nil)
				clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
					return true, newFakeResponseWrapper(getProfilesResp), nil
				})
				updateOptions.Version = "7.0.0"
				err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
				Expect(err).To(MatchError("failed to get profiles from cluster: version '7.0.0' not found for profile 'podinfo' in prod/weave-system"))
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
