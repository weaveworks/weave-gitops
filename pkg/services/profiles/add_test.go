package profiles_test

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/services/profiles"
	"github.com/weaveworks/weave-gitops/pkg/vendorfakes/fakegitprovider"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/testing"
	"sigs.k8s.io/yaml"
)

var addOptions profiles.AddOptions

var _ = Describe("Add", func() {
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

		addOptions = profiles.AddOptions{
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

			It("creates a helm release with the latest available version of the profile", func() {
				fakePR.GetReturns(gitprovider.PullRequestInfo{
					WebURL: "url",
				})
				gitProviders.CreatePullRequestReturns(fakePR, nil)
				Expect(profilesSvc.Add(context.TODO(), gitProviders, addOptions)).Should(Succeed())
			})

			When("auto-merge is enabled", func() {
				It("merges the PR that was created", func() {
					fakePR.GetReturns(gitprovider.PullRequestInfo{
						WebURL: "url",
						Number: 42,
					})
					gitProviders.CreatePullRequestReturns(fakePR, nil)
					addOptions.AutoMerge = true
					Expect(profilesSvc.Add(context.TODO(), gitProviders, addOptions)).Should(Succeed())
				})

				When("the PR fails to be merged", func() {
					It("returns an error", func() {
						fakePR.GetReturns(gitprovider.PullRequestInfo{
							WebURL: "url",
						})
						gitProviders.CreatePullRequestReturns(fakePR, nil)
						gitProviders.MergePullRequestReturns(fmt.Errorf("err"))
						addOptions.AutoMerge = true
						err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
						Expect(err).To(MatchError("error auto-merging PR: err"))
					})
				})
			})

			When("an existing version other than 'latest' is specified", func() {
				It("creates a helm release with that version", func() {
					addOptions.Version = "6.0.0"
					fakePR.GetReturns(gitprovider.PullRequestInfo{
						WebURL: "url",
					})
					fakePR.GetReturns(gitprovider.PullRequestInfo{
						WebURL: "url",
					})
					gitProviders.CreatePullRequestReturns(fakePR, nil)
					err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
					Expect(err).To(BeNil())
				})
			})

			When("it fails to create a pull request to write the helm release to the config repo", func() {
				It("returns an error when ", func() {
					gitProviders.CreatePullRequestReturns(nil, fmt.Errorf("err"))
					err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
					Expect(err).To(MatchError("failed to create pull request: err"))
				})
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
				err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
				Expect(err).To(MatchError("failed to get profiles from cluster: failed to make GET request to service weave-system/wego-app path \"/v1/profiles\": nope"))
				Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
			})

			It("fails if it's unable to discover the HelmRepository's name and namespace values", func() {
				clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
					return true, newFakeResponseWrapper(getRespWithoutHelmRepo()), nil
				})
				err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
				Expect(err).To(MatchError("failed to discover HelmRepository's name and namespace"))
				Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
			})
		})

		When("it fails to find a matching version", func() {
			It("returns an error", func() {
				gitProviders.RepositoryExistsReturns(true, nil)
				clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
					return true, newFakeResponseWrapper(getProfilesResp), nil
				})
				addOptions.Version = "7.0.0"
				err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
				Expect(err).To(MatchError("failed to get profiles from cluster: version '7.0.0' not found for profile 'podinfo' in prod/weave-system"))
				Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
			})
		})
	})

	When("the config repository exists", func() {
		It("fails if the --config-repo url format is wrong", func() {
			addOptions = profiles.AddOptions{
				Name:       "foo",
				ConfigRepo: "{http:/-*wrong-url-827",
				Cluster:    "prod",
			}

			err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
			Expect(err).To(MatchError("failed to parse url: could not get provider name from URL {http:/-*wrong-url-827: could not parse git repo url \"{http:/-*wrong-url-827\": parse \"{http:/-*wrong-url-827\": first path segment in URL cannot contain colon"))
		})

		It("fails if the config repo does not exist", func() {
			gitProviders.RepositoryExistsReturns(false, nil)
			err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
			Expect(err).To(MatchError("repository \"ssh://git@github.com/owner/config-repo.git\" could not be found"))
			Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
		})
	})
})

var _ = Describe("AppendProfileToFile", func() {
	var (
		newRelease   *v2beta1.HelmRelease
		existingFile *gitprovider.CommitFile
		path         string
		content      string
	)

	BeforeEach(func() {
		newRelease = helm.MakeHelmRelease(
			"podinfo", "6.0.0", "prod", "weave-system",
			types.NamespacedName{Name: "helm-repo-name", Namespace: "helm-repo-namespace"},
		)
		path = git.GetProfilesPath("prod", models.WegoProfilesPath)
	})

	When("profiles.yaml does not exist", func() {
		It("creates one with the new helm release", func() {
			file, err := profiles.AppendProfileToFile(makeTestFiles(), newRelease, path)
			Expect(err).NotTo(HaveOccurred())
			r, err := yaml.Marshal(newRelease)
			Expect(err).NotTo(HaveOccurred())
			Expect(*file.Content).To(ContainSubstring(string(r)))
		})
	})

	When("profiles.yaml exists", func() {
		When("the manifest contain a release with the same name in that namespace", func() {
			When("the version is different", func() {
				It("appends the release to the manifest", func() {
					existingRelease := helm.MakeHelmRelease(
						"podinfo", "6.0.1", "prod", "weave-system",
						types.NamespacedName{Name: "helm-repo-name", Namespace: "helm-repo-namespace"},
					)
					r, _ := yaml.Marshal(existingRelease)
					content = string(r)
					file, err := profiles.AppendProfileToFile([]*gitprovider.CommitFile{{
						Path:    &path,
						Content: &content,
					}}, newRelease, path)
					Expect(err).NotTo(HaveOccurred())
					Expect(*file.Content).To(ContainSubstring(string(r)))
				})
			})

			When("the version is the same", func() {
				It("fails to add the profile", func() {
					existingRelease, _ := yaml.Marshal(newRelease)
					content = string(existingRelease)
					existingFile = &gitprovider.CommitFile{
						Path:    &path,
						Content: &content,
					}
					_, err := profiles.AppendProfileToFile([]*gitprovider.CommitFile{existingFile}, newRelease, path)
					Expect(err).To(MatchError("version 6.0.0 of profile 'prod-podinfo' already exists in namespace weave-system"))
				})
			})
		})

		It("fails if the manifest contains a resource that is not a HelmRelease", func() {
			content = "content"
			_, err := profiles.AppendProfileToFile([]*gitprovider.CommitFile{{
				Path:    &path,
				Content: &content,
			}}, newRelease, path)
			Expect(err).To(MatchError("error unmarshaling profiles.yaml: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type v2beta1.HelmRelease"))
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
		commitFiles = append(commitFiles, &gitprovider.CommitFile{
			Path:    file.Path,
			Content: file.Content,
		})
	}

	return commitFiles
}

func getRespWithoutHelmRepo() string {
	return `{
		"profiles": [
		  {
			"name": "podinfo",
			"home": "https://github.com/stefanprodan/podinfo",
			"sources": [
			  "https://github.com/stefanprodan/podinfo"
			],
			"description": "Podinfo Helm chart for Kubernetes",
			"keywords": [],
			"maintainers": [
			  {
				"name": "stefanprodan",
				"email": "stefanprodan@users.noreply.github.com",
				"url": ""
			  }
			],
			"icon": "",
			"annotations": {},
			"kubeVersion": ">=1.19.0-0",
			"availableVersions": [
			  "6.0.0",
			  "6.0.1"
			]
		  }
		]
	  }
	  `
}
