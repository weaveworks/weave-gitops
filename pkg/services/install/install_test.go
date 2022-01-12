package install

import (
	"errors"

	"github.com/fluxcd/go-git-providers/gitprovider"

	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"

	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"

	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"

	"github.com/weaveworks/weave-gitops/pkg/flux/fluxfakes"

	"github.com/weaveworks/weave-gitops/pkg/services/gitopswriter"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/logger"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Installer", func() {

	var installer Installer
	var fakeFluxClient *fluxfakes.FakeFlux
	var fakeKubeClient *kubefakes.FakeKube
	var fakeGitClient *gitfakes.FakeGit
	var fakeGitProvider *gitprovidersfakes.FakeGitProvider
	var repoWriter gitopswriter.RepoWriter
	var log logger.Logger
	var namespace string
	var configRepo gitproviders.RepoURL
	var err error
	var _ = BeforeEach(func() {
		namespace = "test-namespace"
		configRepo, err = gitproviders.NewRepoURL("ssh://git@github.com/test-user/test-repo")
		Expect(err).ShouldNot(HaveOccurred())
		fakeFluxClient = &fluxfakes.FakeFlux{}
		fakeKubeClient = &kubefakes.FakeKube{}
		fakeGitClient = &gitfakes.FakeGit{}
		fakeGitProvider = &gitprovidersfakes.FakeGitProvider{}
		log = &loggerfakes.FakeLogger{}
		repoWriter = gitopswriter.NewRepoWriter(log, fakeGitClient, fakeGitProvider)
		installer = NewInstaller(fakeFluxClient, fakeKubeClient, fakeGitClient, fakeGitProvider, log, repoWriter)
	})

	// Should I include more specific error messages matches
	// Or maybe create a template of the errors and reuse it here
	Context("error paths", func() {
		someError := errors.New("some error")

		It("should fail getting cluster name", func() {
			fakeKubeClient.GetClusterNameReturns("", someError)

			err := installer.Install(namespace, configRepo, true)
			Expect(err.Error()).Should(ContainSubstring(someError.Error()))
		})

		It("should fail installing flux", func() {
			fakeKubeClient.GetClusterNameReturns(namespace, nil)

			fakeFluxClient.InstallReturns(nil, someError)

			err := installer.Install(namespace, configRepo, true)
			Expect(err.Error()).Should(ContainSubstring(someError.Error()))
		})

		It("should fail getting bootstrap manifests", func() {
			fakeKubeClient.GetClusterNameReturns(namespace, nil)

			fakeFluxClient.InstallReturnsOnCall(0, nil, nil)
			fakeFluxClient.InstallReturnsOnCall(1, nil, someError)

			err := installer.Install(namespace, configRepo, true)
			Expect(err.Error()).Should(ContainSubstring(someError.Error()))
		})

		It("should fail applying bootstrap manifests", func() {
			fakeKubeClient.GetClusterNameReturns(namespace, nil)

			fakeFluxClient.InstallReturns(nil, nil)

			fakeKubeClient.ApplyReturns(someError)

			err := installer.Install(namespace, configRepo, true)
			Expect(err.Error()).Should(ContainSubstring(someError.Error()))
		})

		It("should fail getting gitops manifests", func() {
			fakeKubeClient.GetClusterNameReturns(namespace, nil)

			fakeFluxClient.InstallReturnsOnCall(0, nil, nil)
			fakeFluxClient.InstallReturnsOnCall(1, nil, nil)

			fakeKubeClient.ApplyReturns(nil)

			fakeFluxClient.InstallReturnsOnCall(2, nil, someError)

			err := installer.Install(namespace, configRepo, true)
			Expect(err.Error()).Should(ContainSubstring(someError.Error()))
		})

		It("should fail getting default branch", func() {
			fakeKubeClient.GetClusterNameReturns(namespace, nil)

			fakeFluxClient.InstallReturns(nil, nil)

			fakeKubeClient.ApplyReturns(nil)

			privateVisibility := gitprovider.RepositoryVisibilityPrivate
			fakeGitProvider.GetRepoVisibilityReturns(&privateVisibility, nil)

			fakeGitProvider.GetDefaultBranchReturnsOnCall(0, "main", nil)

			fakeGitProvider.GetDefaultBranchReturnsOnCall(1, "", someError)

			err := installer.Install(namespace, configRepo, true)
			Expect(err.Error()).Should(ContainSubstring(someError.Error()))
		})

		It("should fail writing directly to branch", func() {
			fakeKubeClient.GetClusterNameReturns(namespace, nil)

			fakeFluxClient.InstallReturns(nil, nil)

			fakeKubeClient.ApplyReturns(nil)

			privateVisibility := gitprovider.RepositoryVisibilityPrivate
			fakeGitProvider.GetRepoVisibilityReturns(&privateVisibility, nil)

			fakeGitProvider.GetDefaultBranchReturns("main", nil)

			fakeGitClient.CloneReturns(false, someError)

			err := installer.Install(namespace, configRepo, true)
			Expect(err.Error()).Should(ContainSubstring(someError.Error()))
		})

		It("should fail creating a pull requests", func() {
			fakeKubeClient.GetClusterNameReturns(namespace, nil)

			fakeFluxClient.InstallReturns(nil, nil)

			fakeKubeClient.ApplyReturns(nil)

			privateVisibility := gitprovider.RepositoryVisibilityPrivate
			fakeGitProvider.GetRepoVisibilityReturns(&privateVisibility, nil)

			fakeGitProvider.GetDefaultBranchReturns("main", nil)

			fakeGitProvider.CreatePullRequestReturns(nil, someError)

			err := installer.Install(namespace, configRepo, false)
			Expect(err.Error()).Should(ContainSubstring(someError.Error()))
		})
	})

})
