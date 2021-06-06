/**
* All tests related to 'wego add' will go into this file
 */

package acceptance

import (
	"fmt"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Weave GitOps Add Tests", func() {

	BeforeEach(func() {
		By("Given I have a brand new cluster", func() {
			_, err := ResetOrCreateCluster(WEGO_DEFAULT_NAMESPACE)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("And I have a wego binary installed on my local machine", func() {
			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
		})
	})

	It("Verify private repo can be added to the cluster by running 'wego add .' ", func() {
		var repoAbsolutePath string
		private := true
		appManifestFilePath := "./data/nginx.yaml"
		defaultSshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		addCommand := "app add . "
		appRepoName := "wego-test-app-" + RandString(8)

		defer deleteRepo(appRepoName)

		By("And application repo does not already exist", func() {
			deleteRepo(appRepoName)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, private)
			appName := appRepoName
			gitAddCommitPush(repoAbsolutePath, appManifestFilePath)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+defaultSshKeyPath, func() {
			setupSSHKey(defaultSshKeyPath)
		})

		By("And I run wego add command", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed("nginx", "my-nginx")
		})

		By("And repos created have private visibility", func() {
			Expect(getRepoVisibility(os.Getenv("GITHUB_ORG"), appRepoName)).Should(ContainSubstring("true"))
		})
	})

	It("Verify public repo can be added to the cluster by running 'wego add . --private=false --private-key --deployment-type=kustomize'", func() {
		var repoAbsolutePath string
		private := false
		appManifestFilePath := "./data/nginx.yaml"
		sshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa_wego"
		addCommand := fmt.Sprintf("app add . --private=false --private-key=%s --deployment-type=kustomize ", sshKeyPath)
		appRepoName := "wego-test-app-" + RandString(8)
		appName := appRepoName

		defer deleteRepo(appRepoName)

		By("And application repo does not already exist", func() {
			deleteRepos(appRepoName)
		})

		By("When I create a public repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, appManifestFilePath)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my ssh key on path ~/.ssh/id_rsa_wego", func() {
			setupSSHKey(sshKeyPath)
		})

		By("And I run wego add command", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see workload is deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed("nginx", "my-nginx")
		})

		By("And repos created have public visibility", func() {
			Expect(getRepoVisibility(os.Getenv("GITHUB_ORG"), appRepoName)).Should(ContainSubstring("false"))
		})
	})

	It("Verify that wego can deploy an app after it is setup with an empty repo initially", func() {
		var repoAbsolutePath string
		private := true
		appManifestFilePath := "./data/nginx.yaml"
		defaultSshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		addCommand := "app add . --private=true"
		appRepoName := "wego-test-app-" + RandString(8)
		appName := appRepoName
		defer deleteRepo(appRepoName)

		By("And application repo does not already exist", func() {
			deleteRepo(appRepoName)
		})

		By("When I create an empty private repo", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, private)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+defaultSshKeyPath, func() {
			setupSSHKey(defaultSshKeyPath)
		})

		By("And I run wego add command", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see wego add command linked the repo to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I git add-commit-push app workload to repo", func() {
			gitAddCommitPush(repoAbsolutePath, appManifestFilePath)
		})

		By("And I should see workload is deployed to the cluster", func() {
			verifyWorkloadIsDeployed("nginx", "my-nginx")
		})

		By("And repos created have private visibility", func() {
			Expect(getRepoVisibility(os.Getenv("GITHUB_ORG"), appRepoName)).Should(ContainSubstring("true"))
		})
	})

	It("SmokeTest - Verify that wego can deploy multiple apps one with private and other with public repo", func() {
		var repoAbsolutePath1 string
		var repoAbsolutePath2 string
		appManifestFilePath1 := "./data/nginx.yaml"
		appManifestFilePath2 := "./data/nginx2.yaml"
		defaultSshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		addCommand := "app add . "
		appRepoName1 := "wego-test-app-" + RandString(8)
		appRepoName2 := "wego-test-app-" + RandString(8)
		appName1 := appRepoName1
		appName2 := appRepoName2

		defer deleteRepo(appRepoName1)
		defer deleteRepo(appRepoName2)

		By("And application repos do not already exist", func() {
			deleteRepo(appRepoName1)
			deleteRepo(appRepoName2)
		})

		By("When I create an empty private repo for app1", func() {
			repoAbsolutePath1 = initAndCreateEmptyRepo(appRepoName1, true)
		})

		By("When I create an empty private repo for app2", func() {
			repoAbsolutePath2 = initAndCreateEmptyRepo(appRepoName2, false)
		})

		By("And I git add-commit-push for app1 with workload", func() {
			gitAddCommitPush(repoAbsolutePath1, appManifestFilePath1)
		})

		By("And I git add-commit-push for app2 with workload", func() {
			gitAddCommitPush(repoAbsolutePath2, appManifestFilePath2)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+defaultSshKeyPath, func() {
			setupSSHKey(defaultSshKeyPath)
		})

		By("And I run wego add command for 1st app", func() {
			runWegoAddCommand(repoAbsolutePath1, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I run wego add command for 2nd app", func() {
			runWegoAddCommand(repoAbsolutePath2, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see wego add command linked the repo1 to the cluster", func() {
			verifyWegoAddCommand(appName1, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I should see wego add command linked the repo2 to the cluster", func() {
			verifyWegoAddCommand(appName2, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I should see workload for app1 is deployed to the cluster", func() {
			verifyWorkloadIsDeployed("nginx", "my-nginx")
		})

		By("And I should see workload for app2 is deployed to the cluster", func() {
			verifyWorkloadIsDeployed("nginx2", "my-nginx2")
		})

		By("And repos created have private visibility", func() {
			Expect(getRepoVisibility(os.Getenv("GITHUB_ORG"), appRepoName1)).Should(ContainSubstring("true"))
			Expect(getRepoVisibility(os.Getenv("GITHUB_ORG"), appRepoName2)).Should(ContainSubstring("false"))
		})
	})

	It("Verify helm repo can be added to the cluster by running 'wego add . --deployment-type=helm --path=./hello-world'", func() {
		var repoAbsolutePath string
		private := true
		appManifestFilePath := "./data/helm-repo/hello-world"
		defaultSshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		appName := "my-helm-app"
		addCommand := "app add . --deployment-type=helm --path=./hello-world --name=" + appName
		appRepoName := "wego-test-app-" + RandString(8)

		defer deleteRepo(appRepoName)

		By("And application repo does not already exist", func() {
			deleteRepo(appRepoName)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, appManifestFilePath)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+defaultSshKeyPath, func() {
			setupSSHKey(defaultSshKeyPath)
		})

		By("And I run wego add command", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			Expect(waitForResource("apps", appName, "default", INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
			Expect(waitForResource("configmaps", "helloworld-configmap", WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
		})
	})

	It("Verify 'wego add' cannot work without controllers installed", func() {

		var repoAbsolutePath string
		private := true
		appManifestFilePath := "./data/nginx.yaml"
		defaultSshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		addCommand := "app add . "
		appRepoName := "wego-test-app-" + RandString(8)
		var addCommandOutput string
		var addCommandErr string

		defer deleteRepo(appRepoName)

		By("And application repo does not already exist", func() {
			deleteRepo(appRepoName)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, appManifestFilePath)
		})

		By("And I have my default ssh key on path "+defaultSshKeyPath, func() {
			setupSSHKey(defaultSshKeyPath)
		})

		By("And I run wego add command", func() {
			addCommandOutput, addCommandErr = runWegoAddCommandWithOutput(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see relevant message in the console", func() {

			Eventually(addCommandOutput).Should(MatchRegexp(`Checking cluster status[.?]+ (Unknown|Unmodified)`))
			Eventually(addCommandErr).Should(ContainSubstring("WeGO not installed... exiting"))
		})
	})

})
