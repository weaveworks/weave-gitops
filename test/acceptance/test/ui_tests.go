package acceptance

import (
	. "github.com/onsi/ginkgo"
	"github.com/sclevine/agouti"
)

var webDriver *agouti.Page

//func initializeUISteps() {
//
//	//By("And I install gitops to my active cluster", func() {
//	//	installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
//	//})
//	//
//	//By("And I run gitops ui", func() {
//	//	_ = runCommandAndReturnSessionOutput(fmt.Sprintf("%s ui run &", WEGO_BIN_PATH))
//	//})
//}

var _ = XDescribe("Weave GitOps UI Test", func() {

	//deleteWegoRuntime := false
	//if os.Getenv("DELETE_WEGO_RUNTIME_ON_EACH_TEST") == "true" {
	//	deleteWegoRuntime = true
	//}

	BeforeEach(func() {
		By("Given I have a brand new cluster", func() {
			//var err error
			//_, err = ResetOrCreateCluster(WEGO_DEFAULT_NAMESPACE, deleteWegoRuntime)
			//Expect(err).ShouldNot(HaveOccurred())

			//Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
			//initializeUISteps()

			//By("When I open up a browser", func() {
			//	var err error
			//	webDriver, err = agouti.NewPage(SELENIUM_SERVICE_URL, agouti.Desired(agouti.Capabilities{
			//		"chromeOptions": map[string][]string{
			//			"args": {
			//				"--disable-gpu",
			//				"--no-sandbox",
			//			}}}.Browser("chrome")))
			//	Expect(err).NotTo(HaveOccurred())
			//	if err != nil {
			//		fmt.Println("Error creating new page: " + err.Error())
			//		return
			//	}
			//})
		})
	})

	//AfterEach(func() {
	//	takeScreenshot()
	//	Expect(webDriver.Destroy()).To(Succeed())
	//})
	//
	//It("SmokeTest - Verify gitops can run UI without apps installed", func() {
	//
	//	dashboardPage := pages.Dashboard(webDriver)
	//	expectedTitle := "Weave GitOps"
	//
	//	By("Then I should be able to navigate to WeGO dashboard", func() {
	//		Expect(webDriver.Navigate(WEGO_UI_URL)).To(Succeed())
	//		str, _ := webDriver.Title()
	//		Expect(str).To(ContainSubstring(expectedTitle))
	//		Expect(dashboardPage.ApplicationTab).Should(BeFound())
	//	})
	//})
})
