package profiles_test

import (
	"context"
	"fmt"

	pb "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	"github.com/weaveworks/weave-gitops/pkg/services/profiles"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/fluxcd/helm-controller/api/v2beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
	)

	BeforeEach(func() {
		gitProviders = &gitprovidersfakes.FakeGitProvider{}
		clientSet = fake.NewSimpleClientset()
		fakeLogger = &loggerfakes.FakeLogger{}
		profilesSvc = profiles.NewService(clientSet, fakeLogger)

		addOptions = profiles.AddOptions{
			ConfigRepo: "ssh://git@github.com/owner/config-repo.git",
			Name:       "podinfo",
			Cluster:    "prod",
			Namespace:  "weave-system",
			Version:    "latest",
		}
	})

	It("creates a helm release with the latest available version of the profile", func() {
		gitProviders.RepositoryExistsReturns(true, nil)
		gitProviders.GetDefaultBranchReturns("main", nil)
		gitProviders.GetRepoFilesReturns(makeTestFiles(), nil)
		clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
			return true, newFakeResponseWrapper(getProfilesResp), nil
		})
		gitProviders.CreatePullRequestReturns(nil, nil)
		Expect(profilesSvc.Add(context.TODO(), gitProviders, addOptions)).Should(Succeed())
		Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
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

	It("fails if the config repo does not exist", func() {
		gitProviders.RepositoryExistsReturns(false, nil)
		err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
		Expect(err).NotTo(BeNil())
		Expect(err).To(MatchError("repository 'ssh://git@github.com/owner/config-repo.git' could not be found"))
	})

	It("fails if it's unable to get a matching available profile from the cluster", func() {
		gitProviders.RepositoryExistsReturns(true, nil)
		gitProviders.GetRepoFilesReturns(makeTestFiles(), nil)
		clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
			return true, newFakeResponseWrapperWithErr("nope"), nil
		})
		err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
		Expect(err).NotTo(BeNil())
		Expect(err).To(MatchError("failed to make GET request to service weave-system/wego-app path \"/v1/profiles\": nope"))
	})

	It("fails if the config repo's filesystem could not be fetched", func() {
		gitProviders.RepositoryExistsReturns(true, nil)
		clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
			return true, newFakeResponseWrapper(getProfilesResp), nil
		})
		gitProviders.GetRepoFilesReturns(nil, fmt.Errorf("err"))
		err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
		Expect(err).NotTo(BeNil())
		Expect(err).To(MatchError("failed to get files in '.weave-gitops/clusters/prod/system' for config repository 'ssh://git@github.com/owner/config-repo.git': err"))
	})

	When("an existing version other than 'latest' is specified", func() {
		JustBeforeEach(func() {
			gitProviders.RepositoryExistsReturns(true, nil)
			gitProviders.GetRepoFilesReturns(makeTestFiles(), nil)
			clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
				return true, newFakeResponseWrapper(getProfilesResp), nil
			})
		})

		It("fails if the given version was not found", func() {
			addOptions.Version = "7.0.0"
			err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
			Expect(err).NotTo(BeNil())
			Expect(err).To(MatchError("version '7.0.0' not found for profile 'podinfo' in prod/weave-system"))
		})

		It("creates a helm release with that version", func() {
			addOptions.Version = "6.0.0"
			gitProviders.CreatePullRequestReturns(nil, nil)
			err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
			Expect(err).To(BeNil())
		})
	})

	It("fails to create a pull request to write the helm release to the config repo", func() {
		gitProviders.RepositoryExistsReturns(true, nil)
		gitProviders.GetRepoFilesReturns(makeTestFiles(), nil)
		clientSet.AddProxyReactor("services", func(action testing.Action) (handled bool, ret restclient.ResponseWrapper, err error) {
			return true, newFakeResponseWrapper(getProfilesResp), nil
		})
		gitProviders.CreatePullRequestReturns(nil, fmt.Errorf("err"))
		err := profilesSvc.Add(context.TODO(), gitProviders, addOptions)
		Expect(err).NotTo(BeNil())
		Expect(err).To(MatchError("failed to create a pull request: err"))
	})
})

var _ = Describe("MakeManifestFile", func() {
	var (
		newRelease   *v2beta1.HelmRelease
		existingFile *gitprovider.CommitFile
		path         string
		content      string
	)

	BeforeEach(func() {
		newRelease = helm.MakeHelmRelease(&pb.Profile{
			Name:              "podinfo",
			AvailableVersions: []string{"6.0.0"},
			HelmRepository: &pb.HelmRepository{
				Name:      "helm-repo-name",
				Namespace: "helm-repo-namespace",
			},
		}, "prod", "weave-system")
		path = git.GetProfilesPath("prod")
	})

	When("profiles.yaml does not exist", func() {
		It("creates one with the new helm release", func() {
			file, err := profiles.MakeManifestFile(makeTestFiles(), newRelease, path)
			Expect(err).NotTo(HaveOccurred())
			r, err := yaml.Marshal(newRelease)
			Expect(err).NotTo(HaveOccurred())
			Expect(*file[0].Content).To(ContainSubstring(string(r)))
		})
	})

	When("profiles.yaml exists", func() {
		When("the manifest contain a release with the same name in that namespace", func() {
			When("the version is different", func() {
				It("appends the release to the manifest", func() {
					existingRelease := helm.MakeHelmRelease(&pb.Profile{
						Name:              "podinfo",
						AvailableVersions: []string{"6.0.1"},
						HelmRepository: &pb.HelmRepository{
							Name:      "helm-repo-name",
							Namespace: "helm-repo-namespace",
						},
					}, "prod", "weave-system")
					r, _ := yaml.Marshal(existingRelease)
					content = string(r)
					file, err := profiles.MakeManifestFile([]*gitprovider.CommitFile{{
						Path:    &path,
						Content: &content,
					}}, newRelease, path)
					Expect(err).NotTo(HaveOccurred())
					Expect(*file[0].Content).To(ContainSubstring(string(r)))
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
					_, err := profiles.MakeManifestFile([]*gitprovider.CommitFile{existingFile}, newRelease, path)
					Expect(err).To(HaveOccurred())
					Expect(err).To(MatchError("version 6.0.0 of profile 'prod-podinfo' already exists in namespace weave-system"))
				})
			})
		})
		It("fails if the manifest contains a resource that is not a HelmRelease", func() {
			content = "content"
			_, err := profiles.MakeManifestFile([]*gitprovider.CommitFile{{
				Path:    &path,
				Content: &content,
			}}, newRelease, path)
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError("error unmarshaling JSON: while decoding JSON: json: cannot unmarshal string into Go value of type v2beta1.HelmRelease"))
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
