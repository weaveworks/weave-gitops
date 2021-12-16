package acceptance

import (
	"fmt"
	"net/http"
	"os"
	"runtime"

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

var _ = XDescribe("Weave GitOps UI Test", func() {

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

			Expect(FileExists(gitopsBinaryPath)).To(BeTrue())
		})

		By("And I run gitops ui", func() {
			_ = runCommandAndReturnSessionOutput(fmt.Sprintf("%s ui run &", gitopsBinaryPath))
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
		appRepoRemoteURL := "https://github.com/" + githubOrg + "/" + tip.appRepoName + ".git"

		dashboardPage = pages.GetDashboardPageElements(webDriver)
		addAppPage = pages.GetAddAppPageElements(webDriver)

		defer deleteRepo(tip.appRepoName, gitproviders.GitProviderGitHub, githubOrg)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("When I navigate to Add Application page", func() {
			Expect(dashboardPage.AddAppButton.Click()).To(Succeed())
		})

		By("Then I should see Add Applcation page", func() {
			Expect(addAppPage.AddAppHeader.Text()).Should(ContainSubstring(addApplicationPageHeader))
		})

		By("When I create an app repo with workload that does not already exist", func() {
			deleteRepo(tip.appRepoName, gitproviders.GitProviderGitHub, githubOrg)
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, gitproviders.GitProviderGitHub, private, githubOrg)
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
		var linkToApp1 *pages.AppListElements
		var linkToApp2 *pages.AppListElements
		public := false
		tip := generateTestInputs()
		appName1 := "loki"
		appName2 := tip.appRepoName
		workloadName1 := "loki-0"
		workloadName2 := tip.workloadName
		deploymentType1 := "Helm"
		deploymentType2 := "Kustomize"
		appManifestFilePath := tip.appManifestFilePath
		appRepoRemoteURL := "ssh://git@github.com/" + githubOrg + "/" + tip.appRepoName + ".git"
		helmRepoURL := "https://charts.kube-ops.io"
		appDetailsPage := pages.GetAppDetailsPageElements(webDriver)

		addCommand1 := "add app --url=" + helmRepoURL + " --chart=" + appName1 + " --config-repo=" + appRepoRemoteURL + " --auto-merge=true"
		addCommand2 := "add app . --deployment-type=kustomize --auto-merge=true"

		defer deleteRepo(tip.appRepoName, gitproviders.GitProviderGitHub, githubOrg)
		defer deleteWorkload(workloadName2, tip.workloadNamespace)
		defer deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName1, EVENTUALLY_DEFAULT_TIMEOUT)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName, gitproviders.GitProviderGitHub, githubOrg)
		})

		By("And application workload is not already deployed to cluster", func() {
			deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName1, EVENTUALLY_DEFAULT_TIMEOUT)
		})

		By("When I create a public repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, gitproviders.GitProviderGitHub, public, githubOrg)
			gitAddCommitPush(repoAbsolutePath, appManifestFilePath)
		})

		By("And I install gitops to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE, appRepoRemoteURL)
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

			linkToApp1 = pages.GetAppListElements(webDriver, appName1)
			linkToApp2 = pages.GetAppListElements(webDriver, appName2)

			Eventually(linkToApp1.AppList).Should(BeFound())
			Eventually(linkToApp2.AppList).Should(BeFound())

			Expect(linkToApp1.AppList.Attribute("href")).To(ContainSubstring(appName1))
			Expect(linkToApp2.AppList.Attribute("href")).To(ContainSubstring(appName2))
		})

		By("When I click on appName2: "+appName2, func() {
			Expect(linkToApp2.AppList.Click()).To(Succeed())
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

		By("Then I should be able to navigate to app details page for app2", func() {
			verifyAppDetailsPage(appName2)

			app2 := pages.GetAppNameElements(webDriver, appName2)
			Expect(app2.AppNameHeader.Text()).To(ContainSubstring(appName2))
			Expect(app2.AppName.Text()).To(ContainSubstring(appName2))

			appDeployment := pages.GetAppTypeElement(webDriver, deploymentType2)
			Expect(appDeployment.AppType.Text()).Should(ContainSubstring(deploymentType2))

			appURL := pages.GetURLElement(webDriver, appRepoRemoteURL)
			Expect(appURL.AppURL.Text()).Should(ContainSubstring(appRepoRemoteURL))

			appPath := pages.GetPathElement(webDriver, "./")
			Expect(appPath.AppPathToManifests.Text()).Should(ContainSubstring("./"))
		})

		By("And I should be able to see status for app2: "+appName2, func() {
			successMsg := pages.GetMessageElements(webDriver, "ReconciliationSucceeded")
			Eventually(successMsg.KustomizeSuccessMessage).Should(BeFound())
		})

		By("And I should be able to navigate back to Applications page", func() {
			Expect(appDetailsPage.ApplicationsHeader.Click()).To(Succeed())
			appPageURL, _ = webDriver.URL()
			Expect(appPageURL).To(ContainSubstring(WEGO_UI_URL + "/applications"))
		})

		By("When I click on appName1: "+appName1, func() {
			_ = webDriver.Refresh()
			linkToApp1 = pages.GetAppListElements(webDriver, appName1)
			Eventually(linkToApp1.AppList).Should(BeFound())
			Expect(linkToApp1.AppList.Click()).To(Succeed())
		})

		By("Then I should be able to navigate to app details page for app1", func() {
			verifyAppDetailsPage(appName1)

			app1 := pages.GetAppNameElements(webDriver, appName1)
			Expect(app1.AppNameHeader.Text()).To(ContainSubstring(appName1))
			Expect(app1.AppName.Text()).To(ContainSubstring(appName1))

			appDeployment := pages.GetAppTypeElement(webDriver, deploymentType1)
			Expect(appDeployment.AppType.Text()).Should(ContainSubstring(deploymentType1))

			appURL := pages.GetURLElement(webDriver, helmRepoURL)
			Expect(appURL.AppURL.Text()).Should(ContainSubstring(helmRepoURL))

			appPath := pages.GetPathElement(webDriver, appName1)
			Expect(appPath.AppPathToManifests.Text()).Should(ContainSubstring(appName1))
		})
	})
})
