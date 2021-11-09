package acceptance

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/test/acceptance/test/pages"
)

var err error
var webDriver *agouti.Page

var _ = Describe("Weave GitOps UI Test", func() {

	applicationPageHeader := "Applications"
	addApplicationPageHeader := "Add Application"

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
		})

		By("And I install gitops to my active cluster", func() {
			_ = runCommandPassThrough([]string{}, WEGO_BIN_PATH, "install")
			VerifyControllersInCluster(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I run gitops ui", func() {
			_ = runCommandAndReturnSessionOutput(fmt.Sprintf("%s ui run &", WEGO_BIN_PATH))
		})

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

		By("When I navigate to the dashboard", func() {
			Expect(webDriver.Navigate(WEGO_UI_URL)).To(Succeed())
		})

		By("Then I should see Application page", func() {
			dashboardPage := pages.GetDashboardPageElements(webDriver)
			pageHeader, _ := dashboardPage.ApplicationsHeader.Text()

			title, _ := webDriver.Title()
			Eventually(title, THIRTY_SECOND_TIMEOUT).Should(ContainSubstring(WEGO_DASHBOARD_TITLE))

			src, _ := dashboardPage.LogoImage.Attribute("src")

			response, err := http.Get(src)
			Expect(err).ShouldNot(HaveOccurred(), "Logo image is broken")
			Expect(response.StatusCode).Should(Equal(200))

			Eventually(pageHeader, THIRTY_SECOND_TIMEOUT).Should(ContainSubstring(applicationPageHeader))
		})
	})

	AfterEach(func() {
		takeScreenshot()
		Expect(webDriver.Destroy()).To(Succeed())
	})

	It("SmokeTest - Verify gitops can add apps from the UI to an empty cluster", func() {
		var repoAbsolutePath string
		tip := generateTestInputs()
		appName := tip.appRepoName
		private := true
		appRepoRemoteURL := "ssh://git@github.com/" + GITHUB_ORG + "/" + tip.appRepoName + ".git"

		dashboardPage := pages.GetDashboardPageElements(webDriver)
		addAppPage := pages.GetAddAppPageElements(webDriver)

		defer deleteRepo(tip.appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("When I navigate to Add Application page", func() {
			Expect(dashboardPage.AddAppButton.Click()).To(Succeed())
		})

		By("Then I should see Add Applcation page", func() {
			Expect(addAppPage.AddAppHeader.Text()).Should(ContainSubstring(addApplicationPageHeader))
		})

		By("When I create an app repo with workload that does not already exist", func() {
			deleteRepo(tip.appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, gitproviders.GitProviderGitHub, private, GITHUB_ORG)
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I add application details in Application form", func() {
			_ = addAppPage.AppName.Fill(appName)
			_ = addAppPage.AppRepoURL.Fill(appRepoRemoteURL)
			_ = addAppPage.PathToManifests.Fill("/")
		})

		By("And default form values are present", func() {
			Expect(addAppPage.AppNamespace.Attribute("value")).Should(ContainSubstring(WEGO_DEFAULT_NAMESPACE))
			Expect(addAppPage.Branch.Attribute("value")).Should(ContainSubstring("main"))
		})

		By("And auto-merge is turned on", func() {
			_ = addAppPage.AutoMergeCheck.Check()
		})

		By("And I submit the Add App form", func() {
			Expect(addAppPage.SubmitButton.Click()).To(Succeed())
		})

		By("Then I should see Authentication Button pop up", func() {
			_ = Expect(addAppPage.AuthenticationButton.Visible()).To(BeTrue())
		})
	})

	It("SmokeTest - Verify gitops UI can list details of apps running in the cluster", func() {
		var repoAbsolutePath string
		tip := generateTestInputs()
		public := false
		appName1 := "loki"
		appName2 := tip.appRepoName
		workloadName1 := "loki-0"
		workloadName2 := tip.workloadName
		appRepoRemoteURL := "https://github.com/" + GITHUB_ORG + "/" + tip.appRepoName + ".git"
		helmRepoURL := "https://charts.kube-ops.io"

		addCommand1 := "add app --url=" + helmRepoURL + " --chart=" + appName1 + " --app-config-url=" + appRepoRemoteURL + " --auto-merge=true"
		addCommand2 := "add app . --auto-merge=true"

		defer deleteRepo(tip.appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
		defer deleteWorkload(workloadName2, tip.workloadNamespace)
		defer deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName1, EVENTUALLY_DEFAULT_TIMEOUT)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
		})

		By("And application workload is not already deployed to cluster", func() {
			deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName1, EVENTUALLY_DEFAULT_TIMEOUT)
		})

		By("When I create a public repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, gitproviders.GitProviderGitHub, public, GITHUB_ORG)
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I run gitops add command for app1", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand1, WEGO_DEFAULT_NAMESPACE)
			verifyWegoHelmAddCommand(appName1, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I run gitops add command for app2", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand2, WEGO_DEFAULT_NAMESPACE)
			verifyWegoAddCommand(appName2, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see workload1 deployed to the cluster", func() {
			verifyHelmPodWorkloadIsDeployed(workloadName1, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I should see workload2 deployed to the cluster", func() {
			verifyWorkloadIsDeployed(workloadName2, tip.workloadNamespace)
		})

		By("And I should see app names listed on the UI", func() {

			_ = webDriver.Refresh()

		})

		By("When I click on appName1: " + appName2)
	})
})
