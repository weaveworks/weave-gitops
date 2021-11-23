package acceptance

import (
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	. "github.com/sclevine/agouti/matchers"
	log "github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/test/acceptance/test/pages"
)

var err error
var imgSrc string
var pageTitle string
var pageHeader string
var webDriver *agouti.Page
var httpResponse *http.Response
var dashboardPage *pages.DashboardPageElements

var _ = Describe("Weave GitOps UI Test", func() {

	addApplicationPageHeader := "Add Application"

	deleteWegoRuntime := false
	if os.Getenv("DELETE_WEGO_RUNTIME_ON_EACH_TEST") == "true" {
		deleteWegoRuntime = true
	}

	BeforeEach(func() {

		os := runtime.GOOS
		log.Infof("Running tests on OS: " + os)

		By("Given I have a brand new cluster", func() {
			_, _, err = ResetOrCreateCluster(WEGO_DEFAULT_NAMESPACE, deleteWegoRuntime)
			Expect(err).ShouldNot(HaveOccurred())

			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
		})

		By("And I run gitops ui", func() {
			_ = runCommandAndReturnSessionOutput(fmt.Sprintf("%s ui run &", WEGO_BIN_PATH))
		})

		By("And I open up a browser", func() {
			initializeWebDriver(os)
		})

		By("When I navigate to the dashboard", func() {
			Expect(webDriver.Navigate(WEGO_UI_URL)).To(Succeed())
		})

		By("Then I should see application page", func() {
			dashboardPage = pages.GetDashboardPageElements(webDriver)

			pageTitle, _ = webDriver.Title()
			Eventually(pageTitle, THIRTY_SECOND_TIMEOUT).Should(ContainSubstring(WEGO_DASHBOARD_TITLE))

			imgSrc, _ = dashboardPage.LogoImage.Attribute("src")

			httpResponse, err = http.Get(imgSrc)
			Expect(err).ShouldNot(HaveOccurred(), "Logo image is broken")
			Expect(httpResponse.StatusCode).Should(Equal(200))

			Eventually(dashboardPage.ApplicationsHeader).Should(BeFound())

			pageHeader, _ = dashboardPage.ApplicationsHeader.Text()
			Eventually(pageHeader).Should(ContainSubstring(APP_PAGE_HEADER))
		})
	})

	AfterEach(func() {
		takeScreenshot()
		Expect(webDriver.Destroy()).To(Succeed())
	})

	It("UITest - Verify gitops can add apps from the UI to an empty cluster", func() {
		var addAppPage *pages.AddAppPageElements
		var repoAbsolutePath string
		tip := generateTestInputs()
		appName := tip.appRepoName
		private := true
		appRepoRemoteURL := "https://github.com/" + GITHUB_ORG + "/" + tip.appRepoName + ".git"

		dashboardPage = pages.GetDashboardPageElements(webDriver)
		addAppPage = pages.GetAddAppPageElements(webDriver)

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

		By("And I install gitops to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE, appRepoRemoteURL)
		})

		By("And I add application details in Application form", func() {
			_ = addAppPage.AppName.Fill(appName)
			_ = addAppPage.AppRepoURL.Fill(appRepoRemoteURL)
			_ = addAppPage.PathToManifests.Fill("./")
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
	})

	It("UITest - Verify gitops UI can list details of apps running in the cluster", func() {
		var appPageURL string
		var repoAbsolutePath string
		var linkToApp *pages.AppListElements
		public := false
		tip := generateTestInputs()
		appName := tip.appRepoName
		workloadName := tip.workloadName
		deploymentType := "Kustomize"
		appManifestFilePath := tip.appManifestFilePath
		appRepoRemoteURL := "ssh://git@github.com/" + GITHUB_ORG + "/" + tip.appRepoName + ".git"
		appDetailsPage := pages.GetAppDetailsPageElements(webDriver)

		addCommand := "add app . --deployment-type=kustomize --auto-merge=true"

		defer deleteRepo(tip.appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
		defer deleteWorkload(workloadName, tip.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
		})

		By("And application workload is not already deployed to cluster", func() {
			deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName, EVENTUALLY_DEFAULT_TIMEOUT)
		})

		By("When I create a public repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, gitproviders.GitProviderGitHub, public, GITHUB_ORG)
			gitAddCommitPush(repoAbsolutePath, appManifestFilePath)
		})

		By("And I install gitops to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE, appRepoRemoteURL)
		})

		By("And I run gitops add command for app", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I should see workload deployed to the cluster", func() {
			verifyWorkloadIsDeployed(workloadName, tip.workloadNamespace)
		})

		By("And I should see app names listed on the UI", func() {
			_ = webDriver.Refresh()

			linkToApp = pages.GetAppListElements(webDriver, appName)

			Eventually(linkToApp.AppList, 5*time.Second).Should(BeFound())

			Expect(linkToApp.AppList.Attribute("href")).To(ContainSubstring(appName))
		})

		By("When I click on appName: "+appName, func() {
			Expect(linkToApp.AppList.Click()).To(Succeed())
		})

		verifyAppDetailsPage := func(appName string) {
			appPageURL, _ = webDriver.URL()
			Eventually(appPageURL).Should(MatchRegexp(WEGO_UI_URL + `/application_detail.*` + appName))
			Eventually(appDetailsPage.ApplicationsHeader).Should(BeFound())
			Eventually(appDetailsPage.NameSubheader).Should(BeFound())
			Eventually(appDetailsPage.DeploymentTypeSubheader).Should(BeFound())
			Eventually(appDetailsPage.URLSubheader).Should(BeFound())
			Eventually(appDetailsPage.PathSubheader).Should(BeFound())
		}

		By("Then I should be able to navigate to app details page for app", func() {
			verifyAppDetailsPage(appName)

			app2 := pages.GetAppNameElements(webDriver, appName)
			Expect(app2.AppNameHeader.Text()).To(ContainSubstring(appName))
			Expect(app2.AppName.Text()).To(ContainSubstring(appName))

			appDeployment := pages.GetAppTypeElement(webDriver, deploymentType)
			Expect(appDeployment.AppType.Text()).Should(ContainSubstring(deploymentType))

			appURL := pages.GetURLElement(webDriver, appRepoRemoteURL)
			Expect(appURL.AppURL.Text()).Should(ContainSubstring(appRepoRemoteURL))

			appPath := pages.GetPathElement(webDriver, "./")
			Expect(appPath.AppPathToManifests.Text()).Should(ContainSubstring("./"))
		})
	})
})
