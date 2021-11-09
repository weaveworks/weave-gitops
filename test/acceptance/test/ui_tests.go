package acceptance

import (
	"fmt"
	"os"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops/test/acceptance/test/pages"
)

var err error
var webDriver *agouti.Page

func initializeUISteps() {
	By("And I install gitops to my active cluster", func() {
		_ = runCommandPassThrough([]string{}, WEGO_BIN_PATH, "install")
		VerifyControllersInCluster(WEGO_DEFAULT_NAMESPACE)
	})

	By("And I run gitops ui", func() {
		_ = runCommandAndReturnSessionOutput(fmt.Sprintf("%s ui run &", WEGO_BIN_PATH))
	})
}

var _ = Describe("Weave GitOps UI Test", func() {

	deleteWegoRuntime := false
	if os.Getenv("DELETE_WEGO_RUNTIME_ON_EACH_TEST") == "true" {
		deleteWegoRuntime = true
	}

	BeforeEach(func() {
		os := runtime.GOOS
		log.Infof("Running tests on OS: " + os)

		By("Given I have a brand new cluster", func() {

			_, err = ResetOrCreateCluster(WEGO_DEFAULT_NAMESPACE, deleteWegoRuntime)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
			initializeUISteps()

			By("And I open up a browser", func() {

				if os == "linux" {
					webDriver, err = agouti.NewPage(SELENIUM_SERVICE_URL, agouti.Desired(agouti.Capabilities{
						"chromeOptions": map[string][]string{
							"args": {
								"--disable-gpu",
								"--no-sandbox",
							}}}))
					Expect(err).NotTo(HaveOccurred(), "Error creating new page")
				}

				if os == "darwin" {

					chromeDriver := agouti.ChromeDriver(agouti.ChromeOptions("args", []string{"--disable-gpu", "--no-sandbox"}))
					err = chromeDriver.Start()
					Expect(err).NotTo(HaveOccurred())
					webDriver, err = chromeDriver.NewPage()
					Expect(err).NotTo(HaveOccurred(), "Error creating new page")
				}
			})
		})
	})

	AfterEach(func() {
		takeScreenshot()
		Expect(webDriver.Destroy()).To(Succeed())
	})

	It("SmokeTest - Verify gitops can run UI without apps installed", func() {

		dashboardPage := pages.Dashboard(webDriver)
		expectedTitle := "Weave GitOps"

		By("Then I should be able to navigate to WeGO dashboard", func() {
			Expect(webDriver.Navigate(WEGO_UI_URL)).To(Succeed())
			str, _ := webDriver.Title()
			Eventually(str, THIRTY_SECOND_TIMEOUT).Should(ContainSubstring(expectedTitle))
			Expect(dashboardPage.ApplicationTab).Should(BeFound())
		})
	})
})
