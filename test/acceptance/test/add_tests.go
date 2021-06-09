/**
* All tests related to 'wego add' will go into this file
 */

package acceptance

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
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
		workloadName := "nginx"
		workloadNamespace := "my-nginx"
		defaultSshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		addCommand := "app add . "
		appRepoName := "wego-test-app-" + RandString(8)
		appName := appRepoName

		defer deleteRepo(appRepoName)
		defer deleteWorkload(workloadName, workloadNamespace)

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
			verifyWorkloadIsDeployed("nginx", "my-nginx")
		})

		By("And repos created have private visibility", func() {
			Expect(getRepoVisibility(os.Getenv("GITHUB_ORG"), appRepoName)).Should(ContainSubstring("true"))
		})
	})

	It("Verify that wego can deploy an app after it is setup with an empty repo initially", func() {
		var repoAbsolutePath string
		private := true
		appManifestFilePath := "./data/nginx.yaml"
		workloadName := "nginx"
		workloadNamespace := "my-nginx"
		defaultSshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		addCommand := "app add ."
		appRepoName := "wego-test-app-" + RandString(8)
		appName := appRepoName
		defer deleteRepo(appRepoName)
		defer deleteWorkload(workloadName, workloadNamespace)

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
			verifyWorkloadIsDeployed(workloadName, workloadNamespace)
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
		workloadName1 := "nginx"
		workloadName2 := "nginx2"
		workloadNamespace1 := "my-nginx"
		workloadNamespace2 := "my-nginx2"

		defaultSshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		appRepoName1 := "wego-test-app-" + RandString(8)
		appRepoName2 := "wego-test-app-" + RandString(8)
		appName1 := appRepoName1
		appName2 := appRepoName2
		addCommand1 := "app add . --name=" + appName1
		addCommand2 := "app add . --name=" + appName2

		defer deleteRepo(appRepoName1)
		defer deleteRepo(appRepoName2)
		defer deleteWorkload(workloadName1, workloadNamespace1)
		defer deleteWorkload(workloadName2, workloadNamespace2)

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
			runWegoAddCommand(repoAbsolutePath1, addCommand1, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I run wego add command for 2nd app", func() {
			runWegoAddCommand(repoAbsolutePath2, addCommand2, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see wego add command linked the repo1 to the cluster", func() {
			verifyWegoAddCommand(appName1, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I should see wego add command linked the repo2 to the cluster", func() {
			verifyWegoAddCommand(appName2, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I should see workload for app1 is deployed to the cluster", func() {
			verifyWorkloadIsDeployed(workloadName1, workloadNamespace1)
		})

		By("And I should see workload for app2 is deployed to the cluster", func() {
			verifyWorkloadIsDeployed(workloadName2, workloadNamespace2)
		})

		By("And repos created have private visibility", func() {
			Expect(getRepoVisibility(os.Getenv("GITHUB_ORG"), appRepoName1)).Should(ContainSubstring("true"))
			Expect(getRepoVisibility(os.Getenv("GITHUB_ORG"), appRepoName2)).Should(ContainSubstring("false"))
		})
	})

	It("SmokeTest - Verify helm repo can be added to the cluster by running 'wego add . --deployment-type=helm --path=./hello-world'", func() {
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
			verifyWegoAddCommand(appRepoName, WEGO_DEFAULT_NAMESPACE)
			Expect(waitForResource("apps", appName, "default", INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
			Expect(waitForResource("configmaps", "helloworld-configmap", WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
		})
	})

	It("Verify 'wego add' does not work without controllers installed", func() {

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
			Eventually(addCommandErr).Should(MatchRegexp(`WeGO.*... exiting`))
		})
	})

	It("Smoke - Verify 'wego app add' with --dry-run flag does not modify the cluster", func() {
		var repoAbsolutePath string
		var session *gexec.Session
		private := true
		branchName := "test-branch"
		appManifestFilePath := "./data/nginx.yaml"
		defaultSshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		appRepoName := "wego-test-app-" + RandString(8)
		workloadName := "nginx"
		workloadNamespace := "my-nginx"
		url := "ssh://git@github.com/weaveworks-gitops-test/" + appRepoName + ".git"
		addCommand := "app add . --url=" + url + " --branch=" + branchName + " --dry-run"
		appName := appRepoName
		appType := "Kustomization"

		defer deleteRepo(appRepoName)
		defer deleteWorkload(workloadName, workloadNamespace)

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

		By("And I create a new branch", func() {
			createGitRepoBranch(branchName)
		})

		By("And I run 'wego app add dry-run' command", func() {
			session = runWegoAddCommandAndReturnSession(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see dry-run output with specified: url, namespace, branch", func() {
			Eventually(session).Should(gbytes.Say("using URL: '" + url + "'"))
			Eventually(session).Should(gbytes.Say("Checking cluster status... FluxInstalled"))
			Eventually(session).Should(gbytes.Say(`apiVersion:.*\nkind: GitRepository\nmetadata:\n\s*name: ` + appName + `\n\s*namespace: ` + WEGO_DEFAULT_NAMESPACE + `[a-z0-9:\n\s*]+branch: ` + branchName + `\n\s*.*\n\s*name: ` + appName + `\n\s*url: ` + url))
			Eventually(session).Should(gbytes.Say(
				`apiVersion:.*\nkind: ` + appType + `\nmetadata:\n\s*name: ` + appName + `\n\s*namespace: ` + WEGO_DEFAULT_NAMESPACE + `[\w\d\W\n\s*]+kind: GitRepository\n\s*name: ` + appName))
		})

		By("And I should not see any workload deployed to the cluster", func() {
			verifyWegoAddCommandWithDryRun(appRepoName, WEGO_DEFAULT_NAMESPACE)
		})
	})

	// Eventually this test run will include all the remaining un-automated `wego app add` flags.
	It("Smoke - Verify 'wego app add' works when --url flag is specified", func() {
		var repoAbsolutePath string
		private := true
		appRepoName := "wego-test-app-" + RandString(8)
		url := "ssh://git@github.com/weaveworks-gitops-test/" + appRepoName + ".git"
		appManifestFilePath := "./data/nginx.yaml"
		workloadName := "nginx"
		workloadNamespace := "my-nginx"
		defaultSshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		addCommand := "app add . --url=" + url
		appName := appRepoName
		var addCommandOutput string

		defer deleteRepo(appRepoName)
		defer deleteWorkload(workloadName, workloadNamespace)

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

		By("And I run wego add command with url flag specified", func() {
			addCommandOutput, _ = runWegoAddCommandWithOutput(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see wego using the specified url", func() {
			Eventually(addCommandOutput).Should(ContainSubstring("using URL: '" + url + "'"))
		})

		By("And I should see should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed("nginx", "my-nginx")
		})
	})
})
