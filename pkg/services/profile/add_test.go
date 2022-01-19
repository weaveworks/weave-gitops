package profile_test

import (
	"fmt"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/services/profile"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var addParams profile.AddParams

var _ = Describe("Add Profile", func() {
	BeforeEach(func() {
		addParams = profile.AddParams{
			ConfigRepo: "ssh://git@github.com/owner/config-repo.git",
			Name:       "foo",
		}
	})

	It("adds a profile", func() {
		gitProviders.RepositoryExistsReturns(true, nil)
		gitProviders.GetRepoFilesReturns(makeTestFiles(), nil)
		Expect(profileSvc.Add(gitProviders, addParams)).Should(Succeed())
		Expect(gitProviders.RepositoryExistsCallCount()).To(Equal(1))
	})

	It("fails if the config repo does not exist", func() {
		gitProviders.RepositoryExistsReturns(false, nil)
		err := profileSvc.Add(gitProviders, addParams)
		Expect(err).NotTo(BeNil())
		Expect(err).To(MatchError("repository 'ssh://git@github.com/owner/config-repo.git' could not be found"))
	})

	It("fails if the --config-repo url format is wrong", func() {
		addParams = profile.AddParams{
			Name:       "foo",
			ConfigRepo: "{http:/-*wrong-url-827",
		}

		err := profileSvc.Add(gitProviders, addParams)
		Expect(err).NotTo(BeNil())
		Expect(err).To(MatchError("could not get provider name from URL {http:/-*wrong-url-827: could not parse git repo url \"{http:/-*wrong-url-827\": parse \"{http:/-*wrong-url-827\": first path segment in URL cannot contain colon"))
	})

	It("fails if the config repo's filesystem could not be fetched", func() {
		gitProviders.RepositoryExistsReturns(true, nil)
		gitProviders.GetRepoFilesReturns(nil, fmt.Errorf("err"))
		err := profileSvc.Add(gitProviders, addParams)
		Expect(err).NotTo(BeNil())
		Expect(err).To(MatchError("failed to get files of config repository 'ssh://git@github.com/owner/config-repo.git': err"))
	})

	It("fails if the cluster name is not found", func() {
		gitProviders.RepositoryExistsReturns(true, nil)
		path, content := "", ""
		file := &gitprovider.CommitFile{
			Path:    &path,
			Content: &content,
		}
		gitProviders.GetRepoFilesReturns([]*gitprovider.CommitFile{file}, nil)
		err := profileSvc.Add(gitProviders, addParams)
		Expect(err).NotTo(BeNil())
		Expect(err).To(MatchError("failed to find cluster in '/.weave-gitops/clusters/'"))
	})
})

var _ = Describe("ValidateAddParams", func() {
	It("fails if --config-repo is not provided", func() {
		_, err := profileSvc.ValidateAddParams(profile.AddParams{})
		Expect(err).NotTo(BeNil())
		Expect(err).To(MatchError("--config-repo should be provided"))
	})

	When("--name is specified", func() {
		It("fails if --name value is <= 63 characters in length", func() {
			addParams = profile.AddParams{
				Name:       "a234567890123456789012345678901234567890123456789012345678901234",
				ConfigRepo: "ssh://git@github.com/owner/config-repo.git",
			}

			_, err := profileSvc.ValidateAddParams(addParams)
			Expect(err).NotTo(BeNil())
			Expect(err).To(MatchError("--name value is too long: a234567890123456789012345678901234567890123456789012345678901234; must be <= 63 characters"))
		})

		It("fails if --name is prefixed by 'wego'", func() {
			addParams = profile.AddParams{
				Name:       "wego-app",
				ConfigRepo: "ssh://git@github.com/owner/config-repo.git",
			}

			_, err := profileSvc.ValidateAddParams(addParams)
			Expect(err).NotTo(BeNil())
			Expect(err).To(MatchError("the prefix 'wego' is used by weave gitops and is not allowed for a profile name"))
		})
	})

	When("--name is not specified", func() {
		It("fails", func() {
			addParams = profile.AddParams{
				ConfigRepo: "ssh://git@github.com/owner/config-repo.git",
			}

			_, err := profileSvc.ValidateAddParams(addParams)
			Expect(err).NotTo(BeNil())
			Expect(err).To(MatchError("--name should be provided"))
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
