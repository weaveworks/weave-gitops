package profiles_test

import (
	"context"
	"fmt"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/services/profiles"
	"github.com/weaveworks/weave-gitops/pkg/vendorfakes/fakegitprovider"
	"sigs.k8s.io/yaml"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/testing"
)

var updateOptions profiles.Options

var _ = Describe("Update Profile(s)", func() {
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

		updateOptions = profiles.Options{
			ConfigRepo: "ssh://git@github.com/owner/config-repo.git",
			Name:       "podinfo",
			Cluster:    "prod",
			Namespace:  "weave-system",
			Version:    "latest",
		}
	})

	When("the config repository exists", func() {
		BeforeEach(func() {
			gitProviders.RepositoryExistsReturns(true, nil)
			gitProviders.GetDefaultBranchReturns("main", nil)
		})

		AfterEach(func() {
			Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
		})

		When("the version and HelmRepository name and namespace were discovered", func() {
			BeforeEach(func() {
				clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
					return true, newFakeResponseWrapper(getProfilesResp), nil
				})
			})

			When("the file containing the HelmReleases is not empty", func() {
				When("a matching HelmRelease is found", func() {
					When("the existing HelmRelease is a different version than the want to update to", func() {
						BeforeEach(func() {
							existingRelease := helm.MakeHelmRelease(
								"podinfo", "6.0.0", "prod", "weave-system",
								types.NamespacedName{Name: "helm-repo-name", Namespace: "helm-repo-namespace"},
							)
							r, _ := yaml.Marshal(existingRelease)
							content := string(r)
							path := git.GetProfilesPath("prod", models.WegoProfilesPath)
							gitProviders.GetRepoDirFilesReturns([]*gitprovider.CommitFile{{
								Path:    &path,
								Content: &content,
							}}, nil)
						})

						AfterEach(func() {
							Expect(gitProviders.CreatePullRequestCallCount()).To(Equal(1))
						})

						It("opens a PR to update the profiles HelmRelease version", func() {
							fakePR.GetReturns(gitprovider.PullRequestInfo{
								WebURL: "url",
							})
							gitProviders.CreatePullRequestReturns(fakePR, nil)
							Expect(profilesSvc.Update(context.TODO(), gitProviders, updateOptions)).To(Succeed())
							Expect(gitProviders.GetRepoDirFilesCallCount()).To(Equal(1))
							_, repoURL, prInfo := gitProviders.CreatePullRequestArgsForCall(0)
							Expect(repoURL.String()).To(Equal("ssh://git@github.com/owner/config-repo.git"))
							Expect(prInfo.Title).To(Equal("GitOps update podinfo"))
							Expect(prInfo.Description).To(Equal("Update manifest for podinfo profile"))
							Expect(prInfo.CommitMessage).To(Equal("Update profile manifests"))
							Expect(prInfo.TargetBranch).To(Equal("main"))
							Expect(prInfo.Files).To(HaveLen(1))
							Expect(*prInfo.Files[0].Path).To(Equal(".weave-gitops/clusters/prod/system/profiles.yaml"))
						})

						When("PR settings are configured", func() {
							It("opens a PR with the configuration", func() {
								updateOptions = profiles.Options{
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
								Expect(profilesSvc.Update(context.TODO(), gitProviders, updateOptions)).To(Succeed())
								Expect(gitProviders.GetRepoDirFilesCallCount()).To(Equal(1))
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
								updateOptions.AutoMerge = true
								Expect(profilesSvc.Update(context.TODO(), gitProviders, updateOptions)).Should(Succeed())
								Expect(gitProviders.GetRepoDirFilesCallCount()).To(Equal(1))
							})

							When("the PR fails to be merged", func() {
								It("returns an error", func() {
									fakePR.GetReturns(gitprovider.PullRequestInfo{
										WebURL: "url",
									})
									gitProviders.CreatePullRequestReturns(fakePR, nil)
									gitProviders.MergePullRequestReturns(fmt.Errorf("err"))
									updateOptions.AutoMerge = true
									err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
									Expect(err).To(MatchError("error auto-merging PR: err"))
								})
							})
						})

						When("a version other than 'latest' is specified", func() {
							It("creates a helm release with that version", func() {
								updateOptions.Version = "6.0.1"
								fakePR.GetReturns(gitprovider.PullRequestInfo{
									WebURL: "url",
								})
								gitProviders.CreatePullRequestReturns(fakePR, nil)
								err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
								Expect(err).To(BeNil())
							})
						})

						When("the PR fails to be merged", func() {
							It("returns an error", func() {
								gitProviders.CreatePullRequestReturns(nil, fmt.Errorf("err"))
								err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
								Expect(err).To(MatchError("failed to create pull request: err"))
							})
						})
					})

					When("an existing HelmRelease is the same version as the one to update to", func() {
						It("returns an error", func() {
							existingRelease := helm.MakeHelmRelease(
								"podinfo", "6.0.1", "prod", "weave-system",
								types.NamespacedName{Name: "helm-repo-name", Namespace: "helm-repo-namespace"},
							)
							r, _ := yaml.Marshal(existingRelease)
							content := string(r)
							path := git.GetProfilesPath("prod", models.WegoProfilesPath)
							gitProviders.GetRepoDirFilesReturns([]*gitprovider.CommitFile{{
								Path:    &path,
								Content: &content,
							}}, nil)

							err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
							Expect(err).To(MatchError("failed to update HelmRelease for profile 'podinfo' in profiles.yaml: version 6.0.1 of HelmRelease 'prod-podinfo' already installed in namespace 'weave-system'"))
						})
					})
				})

				When("the file containing the HelmReleases does not contain a HelmRelease with the given name and namespace", func() {
					It("returns an error", func() {
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

						err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
						Expect(err).To(MatchError("failed to update HelmRelease for profile 'podinfo' in profiles.yaml: failed to find HelmRelease 'prod-podinfo' in namespace 'weave-system'"))
					})
				})

				When("the file containing the HelmRelease contains something other than a HelmRelease", func() {
					It("returns an error", func() {
						content := "content"
						path := git.GetProfilesPath("prod", models.WegoProfilesPath)
						gitProviders.GetRepoDirFilesReturns([]*gitprovider.CommitFile{{
							Path:    &path,
							Content: &content,
						}}, nil)

						err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
						Expect(err).To(MatchError("failed to update HelmRelease for profile 'podinfo' in profiles.yaml: error splitting into YAML: error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type v2beta1.HelmRelease"))
					})
				})

				When("the file containing the HelmReleases is empty", func() {
					It("returns an error", func() {
						gitProviders.GetRepoDirFilesReturns(makeTestFiles(), nil)

						err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
						Expect(err).To(MatchError(ContainSubstring("failed to find installed profiles in '.weave-gitops/clusters/prod/system/profiles.yaml'")))

						Expect(gitProviders.GetRepoDirFilesCallCount()).To(Equal(1))
					})
				})
			})
		})

		Context("it fails to discover the HelmRepository name and namespace", func() {
			It("fails if it's unable to list available profiles from the cluster", func() {
				clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
					return true, newFakeResponseWrapperWithErr("nope"), nil
				})
				err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
				Expect(err).To(MatchError("failed to discover HelmRepository: failed to get profiles from cluster: failed to make GET request to service weave-system/wego-app path \"/v1/profiles\": nope"))
			})

			It("fails to find an available profile with the given version", func() {
				clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
					return true, newFakeResponseWrapper(getProfilesResp), nil
				})
				updateOptions.Version = "7.0.0"
				err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
				Expect(err).To(MatchError("failed to discover HelmRepository: failed to get profiles from cluster: version '7.0.0' not found for profile 'podinfo' in prod/weave-system"))
			})
		})
	})

	When("the config repository does not exist", func() {
		It("fails if the --config-repo url format is wrong", func() {
			updateOptions = profiles.Options{
				Name:       "foo",
				ConfigRepo: "{http:/-*wrong-url-827",
				Cluster:    "prod",
			}

			err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
			Expect(err).To(MatchError(ContainSubstring("failed to parse url: could not get provider name from URL {http:/-*wrong-url-827")))
		})

		It("fails if the config repo does not exist", func() {
			gitProviders.RepositoryExistsReturns(false, nil)
			err := profilesSvc.Update(context.TODO(), gitProviders, updateOptions)
			Expect(err).To(MatchError("repository \"ssh://git@github.com/owner/config-repo.git\" could not be found"))
			Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
		})
	})
})
