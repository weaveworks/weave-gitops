/**
* All tests related to 'gitops add' will go into this file
 */

package acceptance

import (
	"fmt"
	"os"
	"os/exec"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
)

var helmClusterName string

var _ = Describe("Weave GitOps Helm App Add Tests", func() {
	deleteWegoRuntime := false
	if os.Getenv("DELETE_WEGO_RUNTIME_ON_EACH_TEST") == "true" {
		deleteWegoRuntime = true
	}

	var _ = BeforeEach(func() {
		By("Given I have a brand new cluster", func() {
			var err error

			_, err = ResetOrCreateCluster(WEGO_DEFAULT_NAMESPACE, deleteWegoRuntime)
			Expect(err).ShouldNot(HaveOccurred())

			helmClusterName = getClusterName()
		})

		By("And I have a gitops binary installed on my local machine", func() {
			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
		})
	})

	It("Verify that gitops cannot work without gitops components installed OR with both url and directory provided", func() {
		var repoAbsolutePath string
		var errOutput string
		var exitCode int
		private := true
		tip := generateTestInputs()
		appRepoRemoteURL := "ssh://git@github.com/" + GITHUB_ORG + "/" + tip.appRepoName + ".git"

		addCommand1 := "app add . --auto-merge=true"
		addCommand2 := "app add . --url=" + appRepoRemoteURL + " --auto-merge=true"

		defer deleteRepo(tip.appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(tip.workloadName, tip.workloadNamespace)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, gitproviders.GitProviderGitHub, private, GITHUB_ORG)
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And Gitops runtime is not installed", func() {
			uninstallWegoRuntime(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I run gitops add command", func() {
			command := exec.Command("sh", "-c", fmt.Sprintf("cd %s && %s %s", repoAbsolutePath, WEGO_BIN_PATH, addCommand1))
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
			exitCode = session.Wait().ExitCode()
		})

		By("Then I should see relevant message in the console", func() {
			// Should  be a failure
			Eventually(exitCode).ShouldNot(Equal(0))
		})

		By("When I run add command with both directory path and url specified", func() {
			_, errOutput = runWegoAddCommandWithOutput(repoAbsolutePath, addCommand2, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see an error", func() {
			Expect(errOutput).To(ContainSubstring("you should choose either --url or the app directory"))
		})
	})

	It("Verify that gitops can deploy an app with app-config-url set to <url>", func() {
		var repoAbsolutePath string
		var configRepoRemoteURL string
		var listOutput string
		var appStatus1 string
		var appStatus2 string
		var appStatus3 string
		var commitList1 string
		var commitList2 string
		private := true
		readmeFilePath := "./data/README.md"
		tip := generateTestInputs()
		appFilesRepoName := tip.appRepoName
		appConfigRepoName := "wego-config-repo-" + RandString(8)
		configRepoRemoteURL = "ssh://git@github.com/" + GITHUB_ORG + "/" + appConfigRepoName + ".git"
		helmRepoURL := "https://charts.kube-ops.io"
		appName1 := appFilesRepoName
		workloadName1 := tip.workloadName
		workloadNamespace1 := tip.workloadNamespace
		appManifestFilePath1 := tip.appManifestFilePath
		appName2 := "my-helm-app"
		appManifestFilePath2 := "./data/helm-repo/hello-world"
		appName3 := "loki"
		workloadName3 := "loki-0"

		addCommand1 := "app add . --app-config-url=" + configRepoRemoteURL + " --auto-merge=true"
		addCommand2 := "app add . --deployment-type=helm --path=./hello-world --name=" + appName2 + " --app-config-url=" + configRepoRemoteURL + " --auto-merge=true"
		addCommand3 := "app add --url=" + helmRepoURL + " --chart=" + appName3 + " --app-config-url=" + configRepoRemoteURL + " --auto-merge=true"

		defer deleteRepo(appFilesRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
		defer deleteRepo(appConfigRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
		defer deleteWorkload(workloadName1, workloadNamespace1)
		defer deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName3, EVENTUALLY_DEFAULT_TIMEOUT)

		By("And application repo does not already exist", func() {
			deleteRepo(appFilesRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
			deleteRepo(appConfigRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(workloadName1, workloadNamespace1)
			deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName3, EVENTUALLY_DEFAULT_TIMEOUT)
		})

		By("When I create a private repo for gitops app config", func() {
			appConfigRepoAbsPath := initAndCreateEmptyRepo(appConfigRepoName, gitproviders.GitProviderGitHub, private, GITHUB_ORG)
			gitAddCommitPush(appConfigRepoAbsPath, readmeFilePath)
		})

		By("When I create a private repo with app1 workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appFilesRepoName, gitproviders.GitProviderGitHub, private, GITHUB_ORG)
			gitAddCommitPush(repoAbsolutePath, appManifestFilePath1)
		})

		By("And I install gitops to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I run gitops app add command for app1: "+appName1, func() {
			runWegoAddCommand(repoAbsolutePath, addCommand1, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see my workload deployed for app1", func() {
			verifyWegoAddCommand(appName1, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed(workloadName1, workloadNamespace1)
		})

		By("When I add manifests for app2", func() {
			gitAddCommitPush(repoAbsolutePath, appManifestFilePath2)
		})

		By("And I run gitops app add command for app2: "+appName2, func() {
			runWegoAddCommand(repoAbsolutePath, addCommand2, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see my workload deployed for app2", func() {
			verifyWegoAddCommand(appName2, WEGO_DEFAULT_NAMESPACE)
			Expect(waitForResource("apps", appName2, WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
			Expect(waitForResource("configmaps", "helloworld-configmap", WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
		})

		By("When I run gitops app add command for app3: "+appName3, func() {
			runWegoAddCommand(repoAbsolutePath, addCommand3, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see my workload deployed for app3", func() {
			verifyWegoHelmAddCommand(appName3, WEGO_DEFAULT_NAMESPACE)
			verifyHelmPodWorkloadIsDeployed(workloadName3, WEGO_DEFAULT_NAMESPACE)
		})

		By("When I check the app status for app1", func() {
			appStatus1, _ = runCommandAndReturnStringOutput(WEGO_BIN_PATH + " app status " + appName1)
		})

		By("Then I should see the status for "+appName1, func() {
			Eventually(appStatus1).Should(ContainSubstring(`Last successful reconciliation:`))
			Eventually(appStatus1).Should(ContainSubstring(`gitrepository/` + appName1))
			Eventually(appStatus1).Should(ContainSubstring(`kustomization/` + appName1))
		})

		By("When I check the app status for app2", func() {
			appStatus2, _ = runCommandAndReturnStringOutput(WEGO_BIN_PATH + " app status " + appName2)
		})

		By("Then I should see the status for "+appName2, func() {
			Eventually(appStatus2).Should(ContainSubstring(`Last successful reconciliation:`))
			Eventually(appStatus2).Should(ContainSubstring(`gitrepository/` + appName2))
			Eventually(appStatus2).Should(ContainSubstring(`helmrelease/` + appName2))
		})

		By("When I check the app status for app3", func() {
			appStatus3, _ = runCommandAndReturnStringOutput(WEGO_BIN_PATH + " app status " + appName3)
		})

		By("Then I should see the status for "+appName3, func() {
			Eventually(appStatus3).Should(ContainSubstring(`Last successful reconciliation:`))
			Eventually(appStatus3).Should(ContainSubstring(`helmrepository/` + appName3))
			Eventually(appStatus3).Should(ContainSubstring(`helmrelease/` + appName3))
		})

		By("When I check for apps list", func() {
			listOutput, _ = runCommandAndReturnStringOutput(WEGO_BIN_PATH + " app list")
		})

		By("Then I should see appNames for all apps listed", func() {
			Eventually(listOutput).Should(ContainSubstring(appName1))
			Eventually(listOutput).Should(ContainSubstring(appName2))
			Eventually(listOutput).Should(ContainSubstring(appName3))
		})

		By("And I should not see gitops components in app repo: "+appFilesRepoName, func() {
			pullGitRepo(repoAbsolutePath)
			folderOutput, _ := runCommandAndReturnStringOutput(fmt.Sprintf("cd %s && ls -al", repoAbsolutePath))
			Expect(folderOutput).ShouldNot(ContainSubstring(".wego"))
			Expect(folderOutput).ShouldNot(ContainSubstring("apps"))
			Expect(folderOutput).ShouldNot(ContainSubstring("targets"))
		})

		By("And I should see gitops components in config repo: "+appConfigRepoName, func() {
			folderOutput, _ := runCommandAndReturnStringOutput(fmt.Sprintf("cd %s && git clone %s && cd %s && ls -al", repoAbsolutePath, configRepoRemoteURL, appConfigRepoName))
			Expect(folderOutput).ShouldNot(ContainSubstring(".wego"))
			Expect(folderOutput).Should(ContainSubstring("apps"))
			Expect(folderOutput).Should(ContainSubstring("targets"))
		})

		By("When I check for list of commits for app1", func() {
			commitList1, _ = runCommandAndReturnStringOutput(fmt.Sprintf("%s app %s get commits", WEGO_BIN_PATH, appName1))
		})

		By("Then I should see the list of commits for app1", func() {
			Eventually(commitList1).Should(MatchRegexp(`COMMIT HASH\s*CREATED AT\s*AUTHOR\s*MESSAGE\s*URL`))
			Eventually(commitList1).Should(MatchRegexp(`[\w]{7}\s*202\d-[0,1][0-9]-[0-3][0-9]T[0-2][0-9]:[0-5][0-9]:[0-5][0-9]Z`))
			Eventually(commitList1).Should(MatchRegexp(`[\w]{7}\s*202\d-[0,1][0-9]-[0-3][0-9]T[0-2][0-9]:[0-5][0-9]:[0-5][0-9]Z`))
		})

		By("When I check for list of commits for app2", func() {
			commitList2, _ = runCommandAndReturnStringOutput(fmt.Sprintf("%s app %s get commits", WEGO_BIN_PATH, appName2))
		})

		By("Then I should see the list of commits for app2", func() {
			Eventually(commitList2).Should(MatchRegexp(`COMMIT HASH\s*CREATED AT\s*AUTHOR\s*MESSAGE\s*URL`))
			Eventually(commitList2).Should(MatchRegexp(`[\w]{7}\s*202\d-[0,1][0-9]-[0-3][0-9]T[0-2][0-9]:[0-5][0-9]:[0-5][0-9]Z`))
			Eventually(commitList2).Should(MatchRegexp(`[\w]{7}\s*202\d-[0,1][0-9]-[0-3][0-9]T[0-2][0-9]:[0-5][0-9]:[0-5][0-9]Z`))
		})
	})

	It("SmokeTest - Verify that gitops can deploy a helm app from a git repo with app-config-url set to NONE", func() {
		var repoAbsolutePath string
		var reAddOutput string
		var removeOutput *gexec.Session
		private := true
		appManifestFilePath := "./data/helm-repo/hello-world"
		appName := "my-helm-app"
		appRepoName := "wego-test-app-" + RandString(8)
		badAppName := "foo"

		addCommand := "app add . --deployment-type=helm --path=./hello-world --name=" + appName + " --app-config-url=NONE"

		defer deleteRepo(appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)

		By("Application and config repo does not already exist", func() {
			deleteRepo(appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, gitproviders.GitProviderGitHub, private, GITHUB_ORG)
			gitAddCommitPush(repoAbsolutePath, appManifestFilePath)
		})

		By("And I install gitops to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I run gitops add command", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			Expect(waitForResource("apps", appName, WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
			Expect(waitForResource("configmaps", "helloworld-configmap", WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
		})

		By("And I should not see gitops components in the remote git repo", func() {
			pullGitRepo(repoAbsolutePath)
			folderOutput, _ := runCommandAndReturnStringOutput(fmt.Sprintf("cd %s && ls -al", repoAbsolutePath))
			Expect(folderOutput).ShouldNot(ContainSubstring(".wego"))
			Expect(folderOutput).ShouldNot(ContainSubstring("apps"))
			Expect(folderOutput).ShouldNot(ContainSubstring("targets"))
		})

		By("When I rerun gitops install", func() {
			_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("%s install", WEGO_BIN_PATH))
		})

		By("Then I should not see any errors", func() {
			VerifyControllersInCluster(WEGO_DEFAULT_NAMESPACE)
		})

		By("When I rerun gitops app add command", func() {
			_, reAddOutput = runCommandAndReturnStringOutput(fmt.Sprintf("cd %s && %s %s", repoAbsolutePath, WEGO_BIN_PATH, addCommand))
		})

		By("Then I should see an error", func() {
			Eventually(reAddOutput).Should(ContainSubstring("Error: failed to add the app " + appName + ": unable to create resource, resource already exists in cluster"))
		})

		By("And app status should remain same", func() {
			out := runCommandAndReturnSessionOutput(WEGO_BIN_PATH + " app status " + appName)
			Eventually(out).Should(gbytes.Say(`helmrelease/` + appName + `\s*True\s*.*False`))
		})

		By("When I run gitops app remove", func() {
			_ = runCommandPassThrough([]string{}, "sh", "-c", fmt.Sprintf("%s app remove %s", WEGO_BIN_PATH, appName))
		})

		By("Then I should see app removed from the cluster", func() {
			_ = waitForAppRemoval(appName, THIRTY_SECOND_TIMEOUT)
		})

		By("When I run gitops app remove for a non-existent app", func() {
			removeOutput = runCommandAndReturnSessionOutput(WEGO_BIN_PATH + " app remove " + badAppName)
		})

		By("Then I should get an error", func() {
			Eventually(removeOutput.Err).Should(gbytes.Say(`Error: failed to create app service: error getting git clients: could not retrieve application "` + badAppName + `": could not get application: apps.wego.weave.works "` + badAppName + `" not found`))
		})
	})

	It("Verify that gitops can deploy a helm app from a git repo with app-config-url set to default", func() {
		var repoAbsolutePath string
		public := false
		appName := "my-helm-app"
		appManifestFilePath := "./data/helm-repo/hello-world"
		appRepoName := "wego-test-app-" + RandString(8)

		addCommand := "app add . --deployment-type=helm --path=./hello-world --name=" + appName + " --auto-merge=true"

		defer deleteRepo(appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)

		By("And application repo does not already exist", func() {
			deleteRepo(appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, gitproviders.GitProviderGitHub, public, GITHUB_ORG)
			gitAddCommitPush(repoAbsolutePath, appManifestFilePath)
		})

		By("And I install gitops to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I run gitops add command", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			Expect(waitForResource("apps", appName, WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
			Expect(waitForResource("configmaps", "helloworld-configmap", WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
		})

		By("And repo created has public visibility", func() {
			Eventually(getGitRepoVisibility(GITHUB_ORG, appRepoName, gitproviders.GitProviderGitHub)).Should(ContainSubstring("public"))
		})
	})

	It("Verify that gitops can deploy a helm app from a git repo with app-config-url set to <url>", func() {
		var repoAbsolutePath string
		var configRepoAbsolutePath string
		private := true
		appManifestFilePath := "./data/helm-repo/hello-world"
		configRepoFiles := "./data/config-repo"
		appName := "my-helm-app"
		appRepoName := "wego-test-app-" + RandString(8)
		configRepoName := "wego-test-config-repo-" + RandString(8)
		configRepoUrl := fmt.Sprintf("ssh://git@github.com/%s/%s.git", os.Getenv("GITHUB_ORG"), configRepoName)

		addCommand := fmt.Sprintf("app add . --app-config-url=%s --deployment-type=helm --path=./hello-world --name=%s --auto-merge=true", configRepoUrl, appName)

		defer deleteRepo(appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
		defer deleteRepo(configRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)

		By("Application and config repo does not already exist", func() {
			deleteRepo(appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
			deleteRepo(configRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, gitproviders.GitProviderGitHub, private, GITHUB_ORG)
			gitAddCommitPush(repoAbsolutePath, appManifestFilePath)
		})

		By("When I create a private repo for my config files", func() {
			configRepoAbsolutePath = initAndCreateEmptyRepo(configRepoName, gitproviders.GitProviderGitHub, private, GITHUB_ORG)
			gitAddCommitPush(configRepoAbsolutePath, configRepoFiles)
		})

		By("And I install gitops to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I run gitops add command", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("There is no .wego folder in the app repo", func() {
			_, err := os.Stat(repoAbsolutePath + "/.wego")
			Expect(os.IsNotExist(err)).To(Equal(true))
		})

		By("The manifests are present in the config repo", func() {
			pullBranch(configRepoAbsolutePath, "main")

			_, err := os.Stat(fmt.Sprintf("%s/apps/%s/app.yaml", configRepoAbsolutePath, appName))
			Expect(err).ShouldNot(HaveOccurred())

			_, err = os.Stat(fmt.Sprintf("%s/targets/%s/%s/%s-gitops-source.yaml", configRepoAbsolutePath, helmClusterName, appName, appName))
			Expect(err).ShouldNot(HaveOccurred())

			_, err = os.Stat(fmt.Sprintf("%s/targets/%s/%s/%s-gitops-deploy.yaml", configRepoAbsolutePath, helmClusterName, appName, appName))
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			Expect(waitForResource("apps", appName, WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
			Expect(waitForResource("configmaps", "helloworld-configmap", WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
		})

	})

	It("Verify that gitops can deploy multiple helm apps from a helm repo with app-config-url set to <url>", func() {
		var repoAbsolutePath string
		var listOutput string
		var appStatus1 string
		var appStatus2 string
		private := true
		appName1 := "rabbitmq"
		appName2 := "zookeeper"
		workloadName1 := "rabbitmq-0"
		workloadName2 := "test-space-zookeeper-0"
		workloadNamespace2 := "test-space"
		readmeFilePath := "./data/README.md"
		appRepoName := "wego-test-app-" + RandString(8)
		appRepoRemoteURL := "ssh://git@github.com/" + GITHUB_ORG + "/" + appRepoName + ".git"
		helmRepoURL := "https://charts.bitnami.com/bitnami"

		addCommand1 := "app add --url=" + helmRepoURL + " --chart=" + appName1 + " --app-config-url=" + appRepoRemoteURL + " --auto-merge=true"
		addCommand2 := "app add --url=" + helmRepoURL + " --chart=" + appName2 + " --app-config-url=" + appRepoRemoteURL + " --auto-merge=true --helm-release-target-namespace=" + workloadNamespace2

		defer deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName1, EVENTUALLY_DEFAULT_TIMEOUT)
		defer deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName2, EVENTUALLY_DEFAULT_TIMEOUT)
		defer deleteRepo(appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
		defer deleteNamespace(workloadNamespace2)

		By("And application repo does not already exist", func() {
			deleteRepo(appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
		})

		By("And application workload is not already deployed to cluster", func() {
			deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName1, EVENTUALLY_DEFAULT_TIMEOUT)
			deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName2, EVENTUALLY_DEFAULT_TIMEOUT)
		})

		By("When I create a private git repo", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, gitproviders.GitProviderGitHub, private, GITHUB_ORG)
			gitAddCommitPush(repoAbsolutePath, readmeFilePath)
		})

		By("And I install gitops under my namespace: "+WEGO_DEFAULT_NAMESPACE, func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I create a namespace for helm-app", func() {
			out, _ := runCommandAndReturnStringOutput("kubectl create ns " + workloadNamespace2)
			Eventually(out).Should(ContainSubstring("namespace/" + workloadNamespace2 + " created"))
		})

		By("And I add a invalid entry without --app-config-url set", func() {
			addCommand := "app add --url=" + helmRepoURL + " --chart=" + appName1 + " --auto-merge=true"
			_, err := runWegoAddCommandWithOutput(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
			Eventually(err).Should(ContainSubstring("--app-config-url should be provided or set to NONE"))
		})

		By("And I run gitops app add command for 1st app", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand1, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I run gitops app add command for 2nd app", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand2, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see workload1 deployed to the cluster", func() {
			verifyWegoHelmAddCommand(appName1, WEGO_DEFAULT_NAMESPACE)
			verifyHelmPodWorkloadIsDeployed(workloadName1, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I should see workload2 deployed to the cluster", func() {
			verifyWegoHelmAddCommand(appName2, WEGO_DEFAULT_NAMESPACE)
			verifyHelmPodWorkloadIsDeployed(workloadName2, workloadNamespace2)
		})

		By("And I should see gitops components in the remote git repo", func() {
			pullGitRepo(repoAbsolutePath)
			folderOutput, _ := runCommandAndReturnStringOutput(fmt.Sprintf("cd %s && ls -al", repoAbsolutePath))
			Expect(folderOutput).ShouldNot(ContainSubstring(".wego"))
			Expect(folderOutput).Should(ContainSubstring("apps"))
			Expect(folderOutput).Should(ContainSubstring("targets"))
		})

		By("When I check for apps list", func() {
			listOutput, _ = runCommandAndReturnStringOutput(WEGO_BIN_PATH + " app list")
		})

		By("Then I should see appNames for both apps listed", func() {
			Eventually(listOutput).Should(ContainSubstring(appName1))
			Eventually(listOutput).Should(ContainSubstring(appName2))
		})

		By("When I check the app status for "+appName1, func() {
			appStatus1, _ = runCommandAndReturnStringOutput(WEGO_BIN_PATH + " app status " + appName1)
		})

		By("Then I should see the status for app1", func() {
			Eventually(appStatus1).Should(ContainSubstring(`Last successful reconciliation:`))
			Eventually(appStatus1).Should(ContainSubstring(`helmrepository/` + appName1))
			Eventually(appStatus1).Should(ContainSubstring(`helmrelease/` + appName1))
		})

		By("When I check the app status for "+appName2, func() {
			appStatus2, _ = runCommandAndReturnStringOutput(WEGO_BIN_PATH + " app status " + appName2)
		})

		By("Then I should see the status for app2", func() {
			Eventually(appStatus2).Should(ContainSubstring(`Last successful reconciliation:`))
			Eventually(appStatus2).Should(ContainSubstring(`helmrepository/` + appName2))
			Eventually(appStatus2).Should(ContainSubstring(`helmrelease/` + appName2))
		})
	})

	It("Verify that gitops can deploy a helm app from a helm repo with app-config-url set to NONE", func() {
		appName := "loki"
		workloadName := "loki-0"
		helmRepoURL := "https://charts.kube-ops.io"

		addCommand := "app add --url=" + helmRepoURL + " --chart=" + appName + " --app-config-url=NONE"

		defer deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName, EVENTUALLY_DEFAULT_TIMEOUT)

		By("And application workload is not already deployed to cluster", func() {
			deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName, EVENTUALLY_DEFAULT_TIMEOUT)
		})

		By("And I install gitops to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I run gitops add command", func() {
			runWegoAddCommand(".", addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoHelmAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyHelmPodWorkloadIsDeployed(workloadName, WEGO_DEFAULT_NAMESPACE)
		})
	})

	It("Verify that a PR is raised against a user repo when skipping auto-merge", func() {
		var repoAbsolutePath string
		tip := generateTestInputs()
		appName := tip.appRepoName
		prLink := ""

		addCommand := "app add . --name=" + appName + " --auto-merge=false"

		defer deleteRepo(tip.appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName, gitproviders.GitProviderGitHub, GITHUB_ORG)
		})

		By("When I create an empty private repo for app", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, gitproviders.GitProviderGitHub, true, GITHUB_ORG)
		})

		By("And I git add-commit-push app manifest", func() {
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I install gitops to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("When I run gitops app add command for app", func() {
			output, _ := runWegoAddCommandWithOutput(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
			re := regexp.MustCompile(`(http|ftp|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
			prLink = re.FindAllString(output, -1)[0]
		})

		By("Then I should see a PR created in user repo", func() {
			verifyPRCreated(repoAbsolutePath, appName, gitproviders.GitProviderGitHub)
		})

		By("When I merge the created PR", func() {
			mergePR(repoAbsolutePath, prLink, gitproviders.GitProviderGitHub)
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed(tip.workloadName, tip.workloadNamespace)
		})
	})
})
