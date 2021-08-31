package acceptance

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
)

var webDriver *agouti.Page

func initializeUISteps() {

	By("And I install wego to my active cluster", func() {
		installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
	})

	By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
		setupSSHKey(DEFAULT_SSH_KEY_PATH)
	})

	By("And I run wego ui", func() {
		_ = runCommandAndReturnSessionOutput(fmt.Sprintf("%s ui run &", WEGO_BIN_PATH))
	})
}

var _ = Describe("Weave GitOps UI Test", func() {

	deleteWegoRuntime := false
	if os.Getenv("DELETE_WEGO_RUNTIME_ON_EACH_TEST") == "true" {
		deleteWegoRuntime = true
	}

	BeforeEach(func() {
		By("Given I have all the setup ready", func() {
			var err error
			_, err = ResetOrCreateCluster(WEGO_DEFAULT_NAMESPACE, deleteWegoRuntime)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
			initializeUISteps()

			By("When I open up a browser", func() {
				webDriver, err = agouti.NewPage(SELENIUM_SERVICE_URL, agouti.Desired(agouti.Capabilities{
					"chromeOptions": map[string][]string{
						"args": {
							"--disable-gpu",
							"--no-sandbox",
						}}}.Browser("chrome")))
				Expect(err).NotTo(HaveOccurred())
				if err != nil {
					fmt.Println("Error creating new page: " + err.Error())
					return
				}
			})
		})
	})

	AfterEach(func() {
		Expect(webDriver.Destroy()).To(Succeed())
	})

	It("SmokeTest - Verify wego can run UI without apps installed", func() {

		// var repoAbsolutePath string
		// private := true
		// tip := generateTestInputs()
		// appName := tip.appRepoName

		// addCommand := "app add . --auto-merge=true"

		By("Then I should be able to navigate to WeGO dashboard", func() {
			Expect(webDriver.Navigate(WEGO_UI_URL)).To(Succeed())
			Expect(webDriver.Title()).To(ContainSubstring("Weave GitOps"))
		})

		// By("When I create a private repo with my app workload", func() {
		// 	repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, private)
		// 	gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		// })

		// By("And I run wego app add command", func() {
		// 	runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		// 	verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
		// 	verifyWorkloadIsDeployed(tip.workloadName, tip.workloadNamespace)
		// })

		// By("Then I should see my app in wego ui dashboard", func() {

		// })
	})
})
