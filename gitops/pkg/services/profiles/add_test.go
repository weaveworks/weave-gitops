package profiles_test

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/gitops/pkg/services/profiles"
	"github.com/weaveworks/weave-gitops/gitops/pkg/vendorfakes/fakegitprovider"

	"github.com/fluxcd/go-git-providers/gitprovider"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/testing"
	"sigs.k8s.io/yaml"
)

var addOptions profiles.Options

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

		addOptions = profiles.Options{
			ConfigRepo: "ssh://git@github.com/owner/config-repo.git",
			Name:       "podinfo",
			Cluster:    "prod",
			Namespace:  "weave-system",
			Version:    "latest",
		}
	})

	When("the config repository exists", func() {
		When("the version and HelmRepository name and namespace were discovered", func() {
			When("the HelmRelease was appended to profiles.yaml", func() {
				BeforeEach(func() {
					gitProviders.RepositoryExistsReturns(true, nil)
					gitProviders.GetDefaultBranchReturns("main", nil)
					gitProviders.GetRepoDirFilesReturns(makeTestFiles(), nil)
					clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
						return true, newFakeResponseWrapper(getProfilesResp), nil
					})
				})

				It("creates a helm release with the latest available version of the profile via a PR", func() {
					fakePR.GetReturns(gitprovider.PullRequestInfo{
						WebURL: "url",
					})
					gitProviders.CreatePullRequestReturns(fakePR, nil)
					Expect(profilesSvc.Add(context.TODO(), gitProviders, addOptions)).Should(Succeed())
					Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
					Expect(gitProviders.GetRepoDirFilesCallCount()).To(Equal(1))
					Expect(gitProviders.CreatePullRequestCallCount()).To(Equal(1))

					_, repoURL, prInfo := gitProviders.CreatePullRequestArgsForCall(0)
					Expect(repoURL.String()).To(Equal("ssh://git@github.com/owner/config-repo.git"))
					Expect(prInfo.Title).To(Equal("GitOps add podinfo"))
					Expect(prInfo.Description).To(Equal("Add manifest for podinfo profile"))
					Expect(prInfo.CommitMessage).To(Equal("Add profile manifests"))
					Expect(prInfo.TargetBranch).To(Equal("main"))
					Expect(prInfo.Files).To(HaveLen(1))
					Expect(*prInfo.Files[0].Path).To(Equal(".weave-gitops/clusters/prod/system/profiles.yaml"))
				})

				When("PR settings are configured", func() {
					It("opens a PR with the configuration", func() {
						addOptions = profiles.Options{
							ConfigRepo:  "ssh://git@github.com/owner/config-repo.git",
							Name:        "podinfo",
							Cluster:     "prod",
							Namespace:   "weave-system",
							Version:     "latest",
							HeadBranch:  "foo",
							BaseBranch:  "bar",
							Message:     "sup",
							Title:       "cool-title",
							Description: "so cool",
						}

						fakePR.GetReturns(gitprovider.PullRequestInfo{
							WebURL: "url",
						})
						gitProviders.CreatePullRequestReturns(fakePR, nil)

						Expect(profilesSvc.Add(context.TODO(), gitProviders, addOptions)).Should(Succeed())
						Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
						Expect(gitProviders.GetRepoDirFilesCallCount()).To(Equal(1))
						Expect(gitProviders.CreatePullRequestCallCount()).To(Equal(1))

						_, repoURL, prInfo := gitProviders.CreatePullRequestArgsForCall(0)
						Expect(repoURL.String()).To(Equal("ssh://git@github.com/owner/config-repo.git"))
						Expect(prInfo.Title).To(Equal("cool-title"))
						Expect(prInfo.Description).To(Equal("so cool"))
						Expect(prInfo.CommitMessage).To(Equal("sup"))
						Expect(prInfo.TargetBranch).To(Equal("foo"))
						Expect(prInfo.NewBranch).To(Equal("bar"))
						Expect(prInfo.Files).To(HaveLen(1))
						Expect(*prInfo.Files[0].Path).To(Equal(".weave-gitops/clusters/prod/system/profiles.yaml"))
					})
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
						Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
						Expect(gitProviders.GetRepoDirFilesCallCount()).To(Equal(1))
						Expect(gitProviders.CreatePullRequestCallCount()).To(Equal(1))
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
						gitProviders.CreatePullRequestReturns(fakePR, nil)
						err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
						Expect(err).NotTo(HaveOccurred())
						Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
						Expect(gitProviders.GetRepoDirFilesCallCount()).To(Equal(1))
						Expect(gitProviders.CreatePullRequestCallCount()).To(Equal(1))
					})
				})

				When("it fails to create a pull request to write the helm release to the config repo", func() {
					It("returns an error", func() {
						gitProviders.CreatePullRequestReturns(nil, fmt.Errorf("err"))
						err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
						Expect(err).To(MatchError("failed to create pull request: err"))
					})
				})
			})
		})

		When("profiles.yaml contains a HelmRelease with the same name in that namespace", func() {
			BeforeEach(func() {
				gitProviders.RepositoryExistsReturns(true, nil)
				gitProviders.GetDefaultBranchReturns("main", nil)

				existingRelease := helm.MakeHelmRelease(
					"podinfo", "6.0.1", "prod", "weave-system",
					types.NamespacedName{Name: "helm-repo-name", Namespace: "helm-repo-namespace"},
				)
				r, _ := yaml.Marshal(existingRelease)
				content := string(r)
				path := git.GetProfilesPath("prod", profiles.ManifestFileName)
				gitProviders.GetRepoDirFilesReturns([]*gitprovider.CommitFile{{
					Path:    &path,
					Content: &content,
				}}, nil)
				clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
					return true, newFakeResponseWrapper(getProfilesResp), nil
				})
			})

			It("fails to append the new HelmRelease to profiles.yaml", func() {
				err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
				Expect(err).To(MatchError("failed to add HelmRelease for profile 'podinfo' to profiles.yaml: found another HelmRelease for profile 'podinfo' in namespace weave-system"))
				Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
				Expect(gitProviders.GetRepoDirFilesCallCount()).To(Equal(1))
			})
		})

		Context("it fails to discover the HelmRepository name and namespace", func() {
			It("fails if it's unable to list available profiles from the cluster", func() {
				gitProviders.RepositoryExistsReturns(true, nil)
				clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
					return true, newFakeResponseWrapperWithErr("nope"), nil
				})
				err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
				Expect(err).To(MatchError("failed to discover HelmRepository: failed to get profiles from cluster: failed to make GET request to service weave-system/wego-app path \"/v1/profiles\": nope"))
				Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
			})

			It("fails to find an available profile with the given version", func() {
				gitProviders.RepositoryExistsReturns(true, nil)
				clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
					return true, newFakeResponseWrapper(getProfilesResp), nil
				})
				addOptions.Version = "7.0.0"
				err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
				Expect(err).To(MatchError("failed to discover HelmRepository: failed to get profiles from cluster: version '7.0.0' not found for profile 'podinfo' in prod/weave-system"))
				Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
			})
		})
	})

	When("the config repository does not exist", func() {
		It("fails if the --config-repo url format is wrong", func() {
			addOptions = profiles.Options{
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
