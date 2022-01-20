package profiles_test

import (
	"context"
	"fmt"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/services/profiles"
	"k8s.io/client-go/kubernetes/fake"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var addOptions profiles.AddOptions

var _ = Describe("Add a Profile", func() {
	var (
		gitProviders *gitprovidersfakes.FakeGitProvider
		profilesSvc  *profiles.ProfilesSvc
		clientSet    *fake.Clientset
		fakeLogger   *loggerfakes.FakeLogger
	)

	BeforeEach(func() {
		gitProviders = &gitprovidersfakes.FakeGitProvider{}
		clientSet = fake.NewSimpleClientset()
		fakeLogger = &loggerfakes.FakeLogger{}
		profilesSvc = profiles.NewService(clientSet)
	})

	Context("when AddOptions are valid", func() {
		JustBeforeEach(func() {
			addOptions = profiles.AddOptions{
				ConfigRepo: "ssh://git@github.com/owner/config-repo.git",
				Name:       "podinfo",
				Cluster:    "prod",
				Logger:     fakeLogger,
				Namespace:  "weave-system",
			}
		})

		It("adds a profile", func() {
			gitProviders.RepositoryExistsReturns(true, nil)
			gitProviders.GetRepoFilesReturns(makeTestFiles(), nil)
			clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
				return true, newFakeResponseWrapper(getProfilesResp), nil
			})
			Expect(profilesSvc.Add(context.TODO(), gitProviders, addOptions)).Should(Succeed())
			Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
		})

		It("fails if the config repo does not exist", func() {
			gitProviders.RepositoryExistsReturns(false, nil)
			err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
			Expect(err).NotTo(BeNil())
			Expect(err).To(MatchError("repository 'ssh://git@github.com/owner/config-repo.git' could not be found"))
		})

		It("fails if the --config-repo url format is wrong", func() {
			addOptions = profiles.AddOptions{
				Name:       "foo",
				ConfigRepo: "{http:/-*wrong-url-827",
				Cluster:    "prod",
			}

			err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
			Expect(err).NotTo(BeNil())
			Expect(err).To(MatchError("could not get provider name from URL {http:/-*wrong-url-827: could not parse git repo url \"{http:/-*wrong-url-827\": parse \"{http:/-*wrong-url-827\": first path segment in URL cannot contain colon"))
		})

		It("fails if the config repo's filesystem could not be fetched", func() {
			gitProviders.RepositoryExistsReturns(true, nil)
			gitProviders.GetRepoFilesReturns(nil, fmt.Errorf("err"))
			err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
			Expect(err).NotTo(BeNil())
			Expect(err).To(MatchError("failed to get files in '.weave-gitops/clusters/prod/system' for config repository 'ssh://git@github.com/owner/config-repo.git': err"))
		})

		It("fails if it's unable to get a list of available profiles from the cluster", func() {
			gitProviders.RepositoryExistsReturns(true, nil)
			gitProviders.GetRepoFilesReturns(makeTestFiles(), nil)
			clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
				return true, newFakeResponseWrapperWithErr("nope"), nil
			})
			err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
			Expect(err).NotTo(BeNil())
			Expect(err).To(MatchError("failed to make GET request to service weave-system/wego-app path \"/v1/profiles\": nope"))
		})

		It("fails if no available profiles was found that matches the name for the profile being added", func() {
			gitProviders.RepositoryExistsReturns(true, nil)
			gitProviders.GetRepoFilesReturns(makeTestFiles(), nil)
			badProfileResp := `{
				"profiles": [
				  {
					"name": "foo"
				  }
				]
			  }
			  `
			clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
				return true, newFakeResponseWrapper(badProfileResp), nil
			})
			err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
			Expect(err).NotTo(BeNil())
			Expect(err).To(MatchError("no available profile 'podinfo' found in prod/weave-system"))
		})

		It("fails if matching available profiles have no available versions", func() {
			gitProviders.RepositoryExistsReturns(true, nil)
			gitProviders.GetRepoFilesReturns(makeTestFiles(), nil)
			badProfileResp := `{
				"profiles": [
				  {
					"name": "podinfo",
					"availableVersions": [
					]
				  }
				]
			  }
			  `
			clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
				return true, newFakeResponseWrapper(badProfileResp), nil
			})
			err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
			Expect(err).NotTo(BeNil())
			Expect(err).To(MatchError("no available version found for profile 'podinfo' in prod/weave-system"))
		})
	})

	Context("when AddOptions are not valid", func() {
		It("fails if --config-repo is not provided", func() {
			err := profilesSvc.Add(context.TODO(), gitProviders, profiles.AddOptions{})
			Expect(err).NotTo(BeNil())
			Expect(err).To(MatchError("--config-repo should be provided"))
		})

		When("--name is specified", func() {
			It("fails if --name value is <= 63 characters in length", func() {
				addOptions = profiles.AddOptions{
					Name:       "a234567890123456789012345678901234567890123456789012345678901234",
					ConfigRepo: "ssh://git@github.com/owner/config-repo.git",
					Cluster:    "prod",
				}

				err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
				Expect(err).NotTo(BeNil())
				Expect(err).To(MatchError("--name value is too long: a234567890123456789012345678901234567890123456789012345678901234; must be <= 63 characters"))
			})

			It("fails if --name is prefixed by 'wego'", func() {
				addOptions = profiles.AddOptions{
					Name:       "wego-app",
					ConfigRepo: "ssh://git@github.com/owner/config-repo.git",
					Cluster:    "prod",
				}

				err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
				Expect(err).NotTo(BeNil())
				Expect(err).To(MatchError("the prefix 'wego' is used by weave gitops and is not allowed for a profile name"))
			})
		})

		When("--name is not specified", func() {
			It("fails", func() {
				addOptions = profiles.AddOptions{
					ConfigRepo: "ssh://git@github.com/owner/config-repo.git",
				}

				err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
				Expect(err).NotTo(BeNil())
				Expect(err).To(MatchError("--name should be provided"))
			})
		})

		When("--cluster is not specified", func() {
			It("fails", func() {
				addOptions = profiles.AddOptions{
					ConfigRepo: "ssh://git@github.com/owner/config-repo.git",
					Name:       "test",
				}

				err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
				Expect(err).NotTo(BeNil())
				Expect(err).To(MatchError("--cluster should be provided"))
			})
		})
	})
})

func makeTestFiles() []*gitprovider.CommitFile {
	path0 := ".weave-gitops/clusters/prod/system/wego-system.yaml"
	content0 := "machine1 yaml content"
	path1 := ".weave-gitops/clusters/prod/system/podinfo-helm-release.yaml"
	content1 := "machine2 yaml content"

	files := []gitprovider.CommitFile{
		{
			Path:    &path0,
			Content: &content0,
		},
		{
			Path:    &path1,
			Content: &content1,
		},
	}

	commitFiles := make([]*gitprovider.CommitFile, 0)
	for _, file := range files {
		path := file.Path
		content := file.Content
		commitFiles = append(commitFiles, &gitprovider.CommitFile{
			Path:    path,
			Content: content,
		})
	}
	return commitFiles
}
