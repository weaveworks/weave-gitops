package acceptance

import (
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Weave GitOps Profiles API", func() {
	var (
		testNamespace    = "test-namespace"
		appRepoRemoteURL string
		tip              TestInputs
	)

	BeforeEach(func() {
		Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
		Expect(GITHUB_ORG).NotTo(BeEmpty())

		_, _, err := ResetOrCreateCluster(testNamespace, true)
		Expect(err).ShouldNot(HaveOccurred())

		private := true
		tip = generateTestInputs()
		_ = initAndCreateEmptyRepo(tip.appRepoName, gitproviders.GitProviderGitHub, private, GITHUB_ORG)
	})

	AfterEach(func() {
		deleteRepo(tip.appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
	})

	FIt("Gets a profile", func() {
		By("Installing the Profiles API")
		appRepoRemoteURL = "git@github.com:" + GITHUB_ORG + "/" + tip.appRepoName + ".git"
		installAndVerifyWego(testNamespace, appRepoRemoteURL)

		// get profile
	})
})
