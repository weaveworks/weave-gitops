/**
* All tests related to 'wego add' will go into this file
 */

package acceptance

import (
	"fmt"
	"os"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var clusterName string

var _ = Describe("Weave GitOps Add Tests", func() {
	deleteWegoRuntime := false
	if os.Getenv("DELETE_WEGO_RUNTIME_ON_EACH_TEST") == "true" {
		deleteWegoRuntime = true
	}

	var _ = BeforeEach(func() {
		By("Given I have a brand new cluster", func() {
			var err error

			_, err = ResetOrCreateCluster(WEGO_DEFAULT_NAMESPACE, deleteWegoRuntime)
			Expect(err).ShouldNot(HaveOccurred())

			clusterName = getClusterName()
		})

		By("And I have a wego binary installed on my local machine", func() {
			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
		})
	})

	It("Verify private repo can be added to the cluster by running 'wego add . --auto-merge=true' ", func() {
		var repoAbsolutePath string
		private := true
		tip := generateTestInputs()

		addCommand := "app add . --auto-merge=true"
		appName := tip.appRepoName

		defer deleteRepo(tip.appRepoName)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(tip.workloadName, tip.workloadNamespace)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego add command", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed(tip.workloadName, tip.workloadNamespace)
		})

		By("And repos created have private visibility", func() {
			Expect(getRepoVisibility(GITHUB_ORG, tip.appRepoName)).Should(ContainSubstring("true"))
		})
	})

	It("SmokeTest - Verify helm repo with app-config-url=NONE can be added to the cluster by running 'wego app add --url=https://charts.kube-ops.io --chart=loki --app-config-url=NONE' ", func() {
		appName := "loki"
		workloadName := "loki-0"
		helmRepoURL := "https://charts.kube-ops.io"
		addCommand := "app add --url=" + helmRepoURL + " --chart=" + appName + " --app-config-url=NONE"

		defer uninstallWego()

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(workloadName, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego add command", func() {
			runWegoAddCommand(".", addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoHelmAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyHelmPodWorkloadIsDeployed(workloadName, WEGO_DEFAULT_NAMESPACE)
		})
	})

	It("Verify that wego can deploy an app after it is setup with an empty repo initially", func() {
		var repoAbsolutePath string
		private := true
		tip := generateTestInputs()

		addCommand := "app add . --auto-merge=true"
		appName := tip.appRepoName
		defer deleteRepo(tip.appRepoName)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(tip.workloadName, tip.workloadNamespace)
		})

		By("When I create an empty private repo", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, private)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego add command", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see wego add command linked the repo to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I git add-commit-push app workload to repo", func() {
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I should see workload is deployed to the cluster", func() {
			verifyWorkloadIsDeployed(tip.workloadName, tip.workloadNamespace)
		})

		By("And repos created have private visibility", func() {
			Expect(getRepoVisibility(GITHUB_ORG, tip.appRepoName)).Should(ContainSubstring("true"))
		})
	})

	It("SmokeTest - Verify that wego can deploy multiple apps one with private and other with public repo", func() {
		var repoAbsolutePath1 string
		var repoAbsolutePath2 string
		tip1 := generateTestInputs()
		tip2 := generateTestInputs()

		appName1 := tip1.appRepoName
		appName2 := tip2.appRepoName
		addCommand1 := "app add . --name=" + appName1 + " --auto-merge=true"
		addCommand2 := "app add . --name=" + appName2 + " --auto-merge=true"

		defer deleteRepo(tip1.appRepoName)
		defer deleteRepo(tip2.appRepoName)
		defer deleteWorkload(tip1.workloadName, tip1.workloadNamespace)
		defer deleteWorkload(tip2.workloadName, tip2.workloadNamespace)

		By("And application repos do not already exist", func() {
			deleteRepo(tip1.appRepoName)
			deleteRepo(tip2.appRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(tip1.workloadName, tip1.workloadNamespace)
			deleteWorkload(tip2.workloadName, tip2.workloadNamespace)
		})

		By("When I create an empty private repo for app1", func() {
			repoAbsolutePath1 = initAndCreateEmptyRepo(tip1.appRepoName, true)
		})

		By("When I create an empty private repo for app2", func() {
			repoAbsolutePath2 = initAndCreateEmptyRepo(tip2.appRepoName, false)
		})

		By("And I git add-commit-push for app1 with workload", func() {
			gitAddCommitPush(repoAbsolutePath1, tip1.appManifestFilePath)
		})

		By("And I git add-commit-push for app2 with workload", func() {
			gitAddCommitPush(repoAbsolutePath2, tip2.appManifestFilePath)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego app add command for 1st app", func() {
			runWegoAddCommand(repoAbsolutePath1, addCommand1, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I run wego app add command for 2nd app", func() {
			runWegoAddCommand(repoAbsolutePath2, addCommand2, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see wego app add command linked the repo1 to the cluster", func() {
			verifyWegoAddCommand(appName1, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I should see wego app add command linked the repo2 to the cluster", func() {
			verifyWegoAddCommand(appName2, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I should see workload for app1 is deployed to the cluster", func() {
			verifyWorkloadIsDeployed(tip1.workloadName, tip1.workloadNamespace)
		})

		By("And I should see workload for app2 is deployed to the cluster", func() {
			verifyWorkloadIsDeployed(tip2.workloadName, tip2.workloadNamespace)
		})

		By("And repos created have private visibility", func() {
			Expect(getRepoVisibility(GITHUB_ORG, tip1.appRepoName)).Should(ContainSubstring("true"))
			Expect(getRepoVisibility(GITHUB_ORG, tip2.appRepoName)).Should(ContainSubstring("false"))
		})
	})

	It("SmokeTest - Verify helm repo can be added to the cluster by running 'wego add . --deployment-type=helm --path=./hello-world'", func() {
		var repoAbsolutePath string
		private := true
		appManifestFilePath := "./data/helm-repo/hello-world"

		appName := "my-helm-app"
		addCommand := "app add . --deployment-type=helm --path=./hello-world --name=" + appName + " --auto-merge=true"
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

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego add command", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			Expect(waitForResource("apps", appName, WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
			Expect(waitForResource("configmaps", "helloworld-configmap", WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
		})
	})

	It("Verify helm repo can be added to the cluster by running 'wego add . --app-config-url=<URL> --deployment-type=helm --path=./hello-world --auto-merge=true'", func() {
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

		defer deleteRepo(appRepoName)
		defer deleteRepo(configRepoName)

		By("Application and config repo does not already exist", func() {
			deleteRepo(appRepoName)
			deleteRepo(configRepoName)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, appManifestFilePath)
		})

		By("When I create a private repo for my config files", func() {
			configRepoAbsolutePath = initAndCreateEmptyRepo(configRepoName, private)
			gitAddCommitPush(configRepoAbsolutePath, configRepoFiles)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego add command", func() {
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

			_, err = os.Stat(fmt.Sprintf("%s/targets/%s/%s/%s-gitops-runtime.yaml", configRepoAbsolutePath, clusterName, appName, appName))
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			Expect(waitForResource("apps", appName, WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
			Expect(waitForResource("configmaps", "helloworld-configmap", WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
		})

	})

	It("SmokeTest - Verify helm repo can be added to the cluster by running 'wego add . --app-config-url=NONE --deployment-type=helm --path=./hello-world'", func() {
		var repoAbsolutePath string
		private := true
		appManifestFilePath := "./data/helm-repo/hello-world"
		appName := "my-helm-app"
		appRepoName := "wego-test-app-" + RandString(8)
		addCommand := fmt.Sprintf("app add . --app-config-url=NONE --deployment-type=helm --path=./hello-world --name=%s", appName)

		defer deleteRepo(appRepoName)

		By("Application and config repo does not already exist", func() {
			deleteRepo(appRepoName)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, appManifestFilePath)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego add command", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("There is no .wego folder in the app repo", func() {
			_, err := os.Stat(repoAbsolutePath + "/.wego")
			Expect(os.IsNotExist(err)).To(Equal(true))
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			Expect(waitForResource("apps", appName, WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
			Expect(waitForResource("configmaps", "helloworld-configmap", WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
		})
	})

	It("Verify 'wego add' does not work without controllers installed", func() {
		var repoAbsolutePath string
		var addCommandOutput string
		var addCommandErr string
		private := true
		tip := generateTestInputs()
		addCommand := "app add . --auto-merge=true"

		defer deleteRepo(tip.appRepoName)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And WeGO runtime is not installed", func() {
			uninstallWegoRuntime(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I run wego add command", func() {
			addCommandOutput, addCommandErr = runWegoAddCommandWithOutput(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see relevant message in the console", func() {
			Eventually(addCommandOutput).Should(MatchRegexp(`No flux or wego installed`))
			Eventually(addCommandErr).Should(MatchRegexp(`Wego.*... exiting`))
		})
	})

	It("Verify 'wego app add' with --dry-run flag does not modify the cluster", func() {
		var repoAbsolutePath string
		var addCommandOutput string
		private := true
		tip := generateTestInputs()
		branchName := "test-branch-01"

		appRepoRemoteURL := "ssh://git@github.com/" + GITHUB_ORG + "/" + tip.appRepoName + ".git"
		addCommand := "app add --url=" + appRepoRemoteURL + " --branch=" + branchName + " --dry-run" + " --auto-merge=true"
		appName := tip.appRepoName
		appType := "Kustomization"

		defer deleteRepo(tip.appRepoName)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(tip.workloadName, tip.workloadNamespace)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I create a new branch", func() {
			createGitRepoBranch(repoAbsolutePath, branchName)
		})

		By("And I run 'wego app add dry-run' command", func() {
			addCommandOutput, _ = runWegoAddCommandWithOutput(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see dry-run output with summary: name, url, path, branch and type", func() {
			Eventually(addCommandOutput).Should(MatchRegexp(`Name: ` + appName))
			Eventually(addCommandOutput).Should(MatchRegexp(`URL: ` + appRepoRemoteURL))
			Eventually(addCommandOutput).Should(MatchRegexp(`Path: ./`))
			Eventually(addCommandOutput).Should(MatchRegexp(`Branch: ` + branchName))
			Eventually(addCommandOutput).Should(MatchRegexp(`Type: kustomize`))

			Eventually(addCommandOutput).Should(MatchRegexp(`✚ Generating Source manifest`))
			Eventually(addCommandOutput).Should(MatchRegexp(`✚ Generating GitOps automation manifests`))
			Eventually(addCommandOutput).Should(MatchRegexp(`✚ Generating Application spec manifest`))
			Eventually(addCommandOutput).Should(MatchRegexp(`► Applying manifests to the cluster`))

			Eventually(addCommandOutput).Should(MatchRegexp(
				`apiVersion:.*\nkind: GitRepository\nmetadata:\n\s*name: ` + appName + `\n\s*namespace: ` + WEGO_DEFAULT_NAMESPACE + `[a-z0-9:\n\s*]+branch: ` + branchName + `[a-zA-Z0-9:\n\s*-]+url: ` + appRepoRemoteURL))

			Eventually(addCommandOutput).Should(MatchRegexp(
				`apiVersion:.*\nkind: ` + appType + `\nmetadata:\n\s*name: ` + appName + `-wego-apps-dir\n\s*namespace: ` + WEGO_DEFAULT_NAMESPACE))
		})

		By("And I should not see any workload deployed to the cluster", func() {
			verifyWegoAddCommandWithDryRun(tip.appRepoName, WEGO_DEFAULT_NAMESPACE)
		})
	})

	// Eventually this test run will include all the remaining un-automated `wego app add` flags.
	It("Verify 'wego app add' works with user-specified branch, namespace, url", func() {
		var repoAbsolutePath string
		private := true
		tip := generateTestInputs()
		branchName := "test-branch-02"

		wegoNamespace := "my-space"
		appRepoRemoteURL := "ssh://git@github.com/" + GITHUB_ORG + "/" + tip.appRepoName + ".git"
		addCommand := "app add --url=" + appRepoRemoteURL + " --branch=" + branchName + " --namespace=" + wegoNamespace + " --auto-merge=true"
		appName := tip.appRepoName

		defer deleteRepo(tip.appRepoName)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)
		defer uninstallWegoRuntime(wegoNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(tip.workloadName, tip.workloadNamespace)
		})

		By("And namespace: "+wegoNamespace+" doesn't exist", func() {
			uninstallWegoRuntime(wegoNamespace)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I install wego under my namespace: "+wegoNamespace, func() {
			installAndVerifyWego(wegoNamespace)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I create a new branch", func() {
			createGitRepoBranch(repoAbsolutePath, branchName)
		})

		By("And I run wego add command with specified branch, namespace, url", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, wegoNamespace)
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, wegoNamespace)
			verifyWorkloadIsDeployed(tip.workloadName, tip.workloadNamespace)
		})

		By("And my app is deployed under specified branch name", func() {
			branchOutput, _ := runCommandAndReturnStringOutput(fmt.Sprintf("kubectl get -n %s GitRepositories", wegoNamespace))
			Eventually(branchOutput).Should(ContainSubstring(appName))
			Eventually(branchOutput).Should(ContainSubstring(branchName))
		})
	})

	It("Verify app repo can be added to the cluster by running 'wego add path/to/repo/dir' ", func() {
		var repoAbsolutePath string
		private := true
		tip := generateTestInputs()

		addCommand := "app add " + tip.appRepoName + "/" + " --auto-merge=true"
		appName := tip.appRepoName

		defer deleteRepo(tip.appRepoName)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(tip.workloadName, tip.workloadNamespace)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego add command from repo parent dir", func() {
			pathToRepoParentDir := repoAbsolutePath + "/../"
			runWegoAddCommand(pathToRepoParentDir, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed(tip.workloadName, tip.workloadNamespace)
		})

		By("And repos created have private visibility", func() {
			Expect(getRepoVisibility(GITHUB_ORG, tip.appRepoName)).Should(ContainSubstring("true"))
		})
	})

	It("Verify app repo can be added to the cluster by running 'wego add . --app-config-url=<git ssh url> --auto-merge=true' ", func() {
		var repoAbsolutePath string
		var configRepoRemoteURL string
		private := true
		tip := generateTestInputs()

		appConfigRepoName := "wego-config-repo-" + RandString(8)
		configRepoRemoteURL = "ssh://git@github.com/" + GITHUB_ORG + "/" + appConfigRepoName + ".git"
		addCommand := "app add . --app-config-url=" + configRepoRemoteURL + " --auto-merge=true"
		appName := tip.appRepoName

		defer deleteRepo(tip.appRepoName)
		defer deleteRepo(appConfigRepoName)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
			deleteRepo(appConfigRepoName)

		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(tip.workloadName, tip.workloadNamespace)
		})

		By("When I create a private repo for wego app config", func() {
			appConfigRepoAbsPath := initAndCreateEmptyRepo(appConfigRepoName, private)
			gitAddCommitPush(appConfigRepoAbsPath, tip.appManifestFilePath)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego add command with --app-config-url param", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed(tip.workloadName, tip.workloadNamespace)
		})
	})

	It("Verify app repo can be added to the cluster by running 'wego add --url=<app repo url> --app-config-url=<git ssh url> --auto-merge=true' ", func() {
		var repoAbsolutePath string
		var configRepoRemoteURL string
		private := true
		tip := generateTestInputs()

		appConfigRepoName := "wego-config-repo-" + RandString(8)
		appRepoRemoteURL := "ssh://git@github.com/" + GITHUB_ORG + "/" + tip.appRepoName + ".git"
		configRepoRemoteURL = "ssh://git@github.com/" + GITHUB_ORG + "/" + appConfigRepoName + ".git"
		addCommand := "app add --url=" + appRepoRemoteURL + " --app-config-url=" + configRepoRemoteURL + " --auto-merge=true"
		appName := tip.appRepoName

		defer deleteRepo(tip.appRepoName)
		defer deleteRepo(appConfigRepoName)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
			deleteRepo(appConfigRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(tip.workloadName, tip.workloadNamespace)
		})

		By("When I create a private repo for wego app config", func() {
			appConfigRepoAbsPath := initAndCreateEmptyRepo(appConfigRepoName, private)
			gitAddCommitPush(appConfigRepoAbsPath, tip.appManifestFilePath)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego add command with --url and --app-config-url params", func() {
			runWegoAddCommand(repoAbsolutePath+"/../", addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed(tip.workloadName, tip.workloadNamespace)
		})
	})

	It("Verify app repo can be added to the cluster by running 'wego add --url=<app repo url> --auto-merge=true' ", func() {
		var repoAbsolutePath string
		private := false
		tip := generateTestInputs()

		appRepoRemoteURL := "ssh://git@github.com/" + GITHUB_ORG + "/" + tip.appRepoName + ".git"
		addCommand := "app add --url=" + appRepoRemoteURL + " --auto-merge=true"
		appName := tip.appRepoName

		defer deleteRepo(tip.appRepoName)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(tip.workloadName, tip.workloadNamespace)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego add command with --url", func() {
			runWegoAddCommand(repoAbsolutePath+"/../", addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed(tip.workloadName, tip.workloadNamespace)
		})
	})

	It("Verify that wego can deploy multiple workloads from a single app repo", func() {
		var repoAbsolutePath string
		tip1 := generateTestInputs()
		tip2 := generateTestInputs()

		appRepoName := "wego-test-app-" + RandString(8)
		appName := appRepoName
		addCommand := "app add . --name=" + appName + " --auto-merge=true"

		defer deleteRepo(appRepoName)
		defer deleteWorkload(tip1.workloadName, tip1.workloadNamespace)
		defer deleteWorkload(tip2.workloadName, tip2.workloadNamespace)

		By("And application repos do not already exist", func() {
			deleteRepo(appRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(tip1.workloadName, tip1.workloadNamespace)
			deleteWorkload(tip2.workloadName, tip2.workloadNamespace)
		})

		By("When I create an empty private repo for app1", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, true)
		})

		By("And I git add-commit-push for app with multiple workloads", func() {
			gitAddCommitPush(repoAbsolutePath, tip1.appManifestFilePath)
			gitAddCommitPush(repoAbsolutePath, tip2.appManifestFilePath)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego add command for 1st app", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see wego add command linked the repo  to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I should see workload for app1 is deployed to the cluster", func() {
			verifyWorkloadIsDeployed(tip1.workloadName, tip1.workloadNamespace)
			verifyWorkloadIsDeployed(tip2.workloadName, tip2.workloadNamespace)
		})
	})

	It("Verify multiple apps dir can be added to the cluster using single repo for wego config", func() {
		var repoAbsolutePath string
		var configRepoRemoteURL string
		private := true
		tip1 := generateTestInputs()
		tip2 := generateTestInputs()
		readmeFilePath := "./data/README.md"

		appRepoName1 := "wego-test-app-" + RandString(8)
		appRepoName2 := "wego-test-app-" + RandString(8)
		appConfigRepoName := "wego-config-repo-" + RandString(8)
		configRepoRemoteURL = "ssh://git@github.com/" + GITHUB_ORG + "/" + appConfigRepoName + ".git"
		addCommand := "app add . --app-config-url=" + configRepoRemoteURL + " --auto-merge=true"
		appName1 := appRepoName1
		appName2 := appRepoName2

		defer deleteRepo(appRepoName1)
		defer deleteRepo(appRepoName2)
		defer deleteRepo(appConfigRepoName)
		defer deleteWorkload(tip1.workloadName, tip1.workloadNamespace)
		defer deleteWorkload(tip2.workloadName, tip2.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(appRepoName1)
			deleteRepo(appRepoName2)
			deleteRepo(appConfigRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(tip1.workloadName, tip1.workloadNamespace)
			deleteWorkload(tip2.workloadName, tip2.workloadNamespace)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("When I create a private repo for wego app config", func() {
			appConfigRepoAbsPath := initAndCreateEmptyRepo(appConfigRepoName, private)
			gitAddCommitPush(appConfigRepoAbsPath, readmeFilePath)
		})

		By("And I create a repo with my app1 workload and run the add the command on it", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName1, private)
			gitAddCommitPush(repoAbsolutePath, tip1.appManifestFilePath)
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I create a repo with my app2 workload and run the add the command on it", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName2, private)
			gitAddCommitPush(repoAbsolutePath, tip2.appManifestFilePath)
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see should see my workloads for app1 and app2 are deployed to the cluster", func() {
			verifyWegoAddCommand(appName1, WEGO_DEFAULT_NAMESPACE)
			verifyWegoAddCommand(appName2, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed(tip1.workloadName, tip1.workloadNamespace)
			verifyWorkloadIsDeployed(tip2.workloadName, tip2.workloadNamespace)
		})
	})

	It("Verify multiple apps dir can be added to the cluster using single app and wego config repos", func() {
		var repoAbsolutePath string
		private := true
		tip1 := generateTestInputs()
		tip2 := generateTestInputs()

		appRepoName := "wego-test-app-" + RandString(8)
		appName1 := "app1"
		appName2 := "app2"
		addCommand1 := "app add . --path=./" + appName1 + " --name=" + appName1 + " --auto-merge=true"
		addCommand2 := "app add . --path=./" + appName2 + " --name=" + appName2 + " --auto-merge=true"

		defer deleteRepo(appRepoName)
		defer deleteWorkload(tip1.workloadName, tip1.workloadNamespace)
		defer deleteWorkload(tip2.workloadName, tip2.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(appRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(tip1.workloadName, tip1.workloadNamespace)
			deleteWorkload(tip2.workloadName, tip2.workloadNamespace)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I create a repo with my app1 and app2 workloads and run the add the command for each app", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, private)
			app1Path := createSubDir(appName1, repoAbsolutePath)
			app2Path := createSubDir(appName2, repoAbsolutePath)
			gitAddCommitPush(app1Path, tip1.appManifestFilePath)
			gitAddCommitPush(app2Path, tip2.appManifestFilePath)
			runWegoAddCommand(repoAbsolutePath, addCommand1, WEGO_DEFAULT_NAMESPACE)
			runWegoAddCommand(repoAbsolutePath, addCommand2, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see should see my workloads for app1 and app2 are deployed to the cluster", func() {
			verifyWegoAddCommand(appName1, WEGO_DEFAULT_NAMESPACE)
			verifyWegoAddCommand(appName2, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed(tip1.workloadName, tip1.workloadNamespace)
			verifyWorkloadIsDeployed(tip2.workloadName, tip2.workloadNamespace)
		})
	})

	It("SmokeTest - Verify wego can add kustomize-based app with 'app-config-url=NONE'", func() {
		var repoAbsolutePath string
		var statusOutput *gexec.Session
		private := true
		tip := generateTestInputs()
		DEFAULT_SSH_KEY_PATH := "~/.ssh/id_rsa"
		addCommand := "app add . --app-config-url=NONE"
		appName := tip.appRepoName

		defer deleteRepo(tip.appRepoName)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(tip.workloadName, tip.workloadNamespace)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I install wego under my namespace: "+WEGO_DEFAULT_NAMESPACE, func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego add command with app-config-url set to 'none'", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed(tip.workloadName, tip.workloadNamespace)
		})

		By("And I should not see wego components in the remote git repo", func() {
			pullGitRepo(repoAbsolutePath)
			folderOutput, _ := runCommandAndReturnStringOutput(fmt.Sprintf("cd %s && ls -al", repoAbsolutePath))
			Expect(folderOutput).ShouldNot(ContainSubstring(".wego"))
			Expect(folderOutput).ShouldNot(ContainSubstring("apps"))
			Expect(folderOutput).ShouldNot(ContainSubstring("targets"))
		})

		By("When I run wego app status", func() {
			statusOutput = runCommandAndReturnSessionOutput(fmt.Sprintf("%s app status %s", WEGO_BIN_PATH, appName))
		})

		By("Then I can see app status", func() {
			Eventually(statusOutput).Should(gbytes.Say(`Last successful reconciliation:`))
			Eventually(statusOutput).Should(gbytes.Say(`gitrepository/` + appName))
			Eventually(statusOutput).Should(gbytes.Say(`kustomization/` + appName))
		})
	})

	It("Verify wego can add kustomize-based app with 'app-config-url=NONE' and a URL", func() {
		var repoAbsolutePath string
		private := true
		tip := generateTestInputs()
		DEFAULT_SSH_KEY_PATH := "~/.ssh/id_rsa"
		appName := tip.appRepoName
		addCommand := fmt.Sprintf("app add --url='ssh://git@github.com/%s' --app-config-url=NONE", GITHUB_ORG+"/"+appName)

		defer deleteRepo(tip.appRepoName)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(tip.workloadName, tip.workloadNamespace)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I install wego under my namespace: "+WEGO_DEFAULT_NAMESPACE, func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego add command with app-config-url set to 'none'", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed(tip.workloadName, tip.workloadNamespace)
		})

		By("And I should not see wego components in the remote git repo", func() {
			pullGitRepo(repoAbsolutePath)
			folderOutput, _ := runCommandAndReturnStringOutput(fmt.Sprintf("cd %s && ls -al", repoAbsolutePath))
			Expect(folderOutput).ShouldNot(ContainSubstring(".wego"))
			Expect(folderOutput).ShouldNot(ContainSubstring("apps"))
			Expect(folderOutput).ShouldNot(ContainSubstring("targets"))
		})
	})

	It("Verify url and directory validation", func() {
		var repoAbsolutePath string
		appRepoName := "wego-test-app-" + RandString(8)
		uniqueSuffix := RandString(6)
		appManifestFilePath := getUniqueWorkload("xxyyzz", uniqueSuffix)
		addCommand := "app add . --url=https://github.com/foo/bar --auto-merge=true"

		repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, true)
		gitAddCommitPush(repoAbsolutePath, appManifestFilePath)

		_, errOut := runWegoAddCommandWithOutput(repoAbsolutePath, addCommand, "wego-system")
		Expect(errOut).To(ContainSubstring("you should choose either --url or the app directory"))
	})

	It("Verify a PR is created in user repo using `wego app add .", func() {
		var repoAbsolutePath string
		tip := generateTestInputs()

		appName := tip.appRepoName
		addCommand := "app add . --name=" + appName
		prLink := ""

		defer deleteRepo(tip.appRepoName)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
		})

		By("When I create an empty private repo for app", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, true)
		})

		By("And I git add-commit-push for app with workload", func() {
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego app add command for app", func() {
			output, _ := runWegoAddCommandWithOutput(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
			re := regexp.MustCompile(`(http|ftp|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
			prLink = re.FindAllString(output, -1)[0]
		})

		By("And I should see a PR is created in user repo", func() {
			verifyPRCreated(repoAbsolutePath, appName)
		})

		By("And I merge the created PR", func() {
			mergePR(repoAbsolutePath, prLink)
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed(tip.workloadName, tip.workloadNamespace)
		})
	})

	It("MyTest - Verify PR can be created in external repo 'wego add . --app-config-url=<git ssh url>' ", func() {
		var repoAbsolutePath string
		var configRepoRemoteURL string
		var appCofigRepoAbsPath string
		private := true
		tip := generateTestInputs()

		appConfigRepoName := "wego-config-repo-" + RandString(8)
		configRepoRemoteURL = "ssh://git@github.com/" + GITHUB_ORG + "/" + appConfigRepoName + ".git"
		addCommand := "app add . --app-config-url=" + configRepoRemoteURL
		appName := tip.appRepoName
		prLink := ""

		defer deleteRepo(tip.appRepoName)
		defer deleteRepo(appConfigRepoName)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
			deleteRepo(appConfigRepoName)

		})

		By("When I create a private repo for wego app config", func() {
			appCofigRepoAbsPath = initAndCreateEmptyRepo(appConfigRepoName, private)
			gitAddCommitPush(appCofigRepoAbsPath, tip.appManifestFilePath)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego add command with --app-config-url param", func() {
			output, _ := runWegoAddCommandWithOutput(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
			re := regexp.MustCompile(`(http|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
			prLink = re.FindAllString(output, 1)[0]
		})

		By("And I should see a PR is created for external repo", func() {
			verifyPRCreated(appCofigRepoAbsPath, appName)
		})

		By("And I merge the created PR", func() {
			mergePR(appCofigRepoAbsPath, prLink)
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed(tip.workloadName, tip.workloadNamespace)
		})
	})

	It("Verify PR fails if branch exists", func() {
		var repoAbsolutePath string
		tip := generateTestInputs()
		tip2 := generateTestInputs()

		appName := tip.appRepoName
		appName2 := tip2.appRepoName
		addCommand := "app add . --name=" + appName
		addCommand2 := "app add . --name=" + appName2

		defer deleteRepo(tip.appRepoName)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
		})

		By("When I create an empty private repo for app", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, true)
		})

		By("And I git add-commit-push for app with workload", func() {
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego app add command for app", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I should see a PR is created", func() {
			verifyPRCreated(repoAbsolutePath, appName)
		})

		By("And I fail to create a PR with the same app", func() {
			_, addCommandErr := runWegoAddCommandWithOutput(repoAbsolutePath, addCommand2, WEGO_DEFAULT_NAMESPACE)
			Expect(addCommandErr).Should(ContainSubstring("422 Reference already exists"))
		})
	})

	It("Verify PR fails if app exists", func() {
		var repoAbsolutePath string
		tip := generateTestInputs()
		tip2 := generateTestInputs()

		appName := tip.appRepoName
		appName2 := tip2.appRepoName
		addCommand := "app add . --name=" + appName + " --auto-merge=true"
		addCommand2 := "app add . --name=" + appName2

		defer deleteRepo(tip.appRepoName)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
		})

		By("When I create an empty private repo for app", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, true)
		})

		By("And I git add-commit-push for app with workload", func() {
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego app add command for app", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed(tip.workloadName, tip.workloadNamespace)
		})

		By("And I fail to create a PR with the same app", func() {
			_, addCommandErr := runWegoAddCommandWithOutput(repoAbsolutePath, addCommand2, WEGO_DEFAULT_NAMESPACE)
			Expect(addCommandErr).Should(ContainSubstring("unable to create resource, resource already exists in cluster"))
		})
	})
})
