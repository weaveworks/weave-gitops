/**
* All tests related to 'wego add' will go into this file
 */

package acceptance

import (
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

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(workloadName, workloadNamespace)
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

		By("Then I should see my workload deployed to the cluster", func() {
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

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(workloadName, workloadNamespace)
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

		// var statusOutput string

		defer deleteRepo(appRepoName1)
		defer deleteRepo(appRepoName2)
		defer deleteWorkload(workloadName1, workloadNamespace1)
		defer deleteWorkload(workloadName2, workloadNamespace2)

		By("And application repos do not already exist", func() {
			deleteRepo(appRepoName1)
			deleteRepo(appRepoName2)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(workloadName1, workloadNamespace1)
			deleteWorkload(workloadName2, workloadNamespace2)
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
			verifyWorkloadIsDeployed(workloadName1, workloadNamespace1)
		})

		By("And I should see workload for app2 is deployed to the cluster", func() {
			verifyWorkloadIsDeployed(workloadName2, workloadNamespace2)
		})

		By("And repos created have private visibility", func() {
			Expect(getRepoVisibility(os.Getenv("GITHUB_ORG"), appRepoName1)).Should(ContainSubstring("true"))
			Expect(getRepoVisibility(os.Getenv("GITHUB_ORG"), appRepoName2)).Should(ContainSubstring("false"))
		})

		// By("When I run wego app status for running apps", func() {
		// 	statusOutput, _ := runCommandAndReturnOutput()
		// })
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

	FIt("SmokeTest - Verify 'wego app add' with --dry-run flag does not modify the cluster", func() {

		var repoAbsolutePath string
		var addCommandOutput string
		private := true
		branchName := "test-branch-01"
		appManifestFilePath := "./data/nginx.yaml"
		defaultSshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		appRepoName := "wego-test-app-" + RandString(8)
		workloadName := "nginx"
		workloadNamespace := "my-nginx"
		url := "ssh://git@github.com/" + os.Getenv("GITHUB_ORG") + "/" + appRepoName + ".git"
		addCommand := "app add . --url=" + url + " --branch=" + branchName + " --dry-run"
		appName := appRepoName
		appType := "Kustomization"

		defer deleteRepo(appRepoName)
		defer deleteWorkload(workloadName, workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(appRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(workloadName, workloadNamespace)
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
			createGitRepoBranch(repoAbsolutePath, branchName)
		})

		By("And I run 'wego app add dry-run' command", func() {
			addCommandOutput, _ = runWegoAddCommandWithOutput(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see dry-run output with specified: url, namespace, branch", func() {
			Eventually(addCommandOutput).Should(MatchRegexp(`using URL: '` + url + `'`))
			Eventually(addCommandOutput).Should(MatchRegexp(`Checking cluster status... WeGOInstalled`))

			Eventually(addCommandOutput).Should(MatchRegexp(
				`Generating deploy key for repo ` + url + ` ...\nGenerating Source manifest...\nGenerating GitOps automation manifests...\nGenerating Application spec manifest...\nApplying manifests to the cluster...`))

			Eventually(addCommandOutput).Should(MatchRegexp(
				`apiVersion:.*\nkind: GitRepository\nmetadata:\n\s*name: ` + appName + `\n\s*namespace: ` + WEGO_DEFAULT_NAMESPACE + `[a-z0-9:\n\s*]+branch: ` + branchName + `[a-zA-Z0-9:\n\s*-]+url: ` + url))

			Eventually(addCommandOutput).Should(MatchRegexp(
				`apiVersion:.*\nkind: ` + appType + `\nmetadata:\n\s*name: ` + appName + `\n\s*namespace: ` + WEGO_DEFAULT_NAMESPACE + `[\w\d\W\n\s*]+kind: GitRepository\n\s*name: ` + appName))
		})

		By("And I should not see any workload deployed to the cluster", func() {
			verifyWegoAddCommandWithDryRun(appRepoName, WEGO_DEFAULT_NAMESPACE)
		})
	})

	// Eventually this test run will include all the remaining un-automated `wego app add` flags.
	It("Verify 'wego app add' works with user-specified branch, namespace", func() {

		var repoAbsolutePath string
		private := true
		appRepoName := "wego-test-app-" + RandString(8)
		branchName := "test-branch-02"
		appManifestFilePath := "./data/nginx.yaml"
		workloadName := "nginx"
		workloadNamespace := "my-nginx"
		defaultSshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		wegoNamespace := "my-space"
		addCommand := "app add . --branch=" + branchName + " --namespace=" + wegoNamespace
		appName := appRepoName

		defer deleteRepo(appRepoName)
		defer deleteWorkload(workloadName, workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(appRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(workloadName, workloadNamespace)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, appManifestFilePath)
		})

		By("And I install wego under my specified namespace", func() {
			installAndVerifyWego(wegoNamespace)
			VerifyControllersInCluster(wegoNamespace)
		})

		By("And I have my default ssh key on path "+defaultSshKeyPath, func() {
			setupSSHKey(defaultSshKeyPath)
		})

		By("And I create a new branch", func() {
			createGitRepoBranch(repoAbsolutePath, branchName)
		})

		By("And I run wego add command with specified branch", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, wegoNamespace)
		})

		By("And I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, wegoNamespace)
			verifyWorkloadIsDeployed("nginx", "my-nginx")
		})

		By("And I verify my specified branch name", func() {
			branchOutput := checkGitBranch(repoAbsolutePath)
			Eventually(branchOutput).Should(ContainSubstring(branchName))
		})
	})

	It("Verify app repo can be added to the cluster by running 'wego add path/to/repo/dir' ", func() {
		var repoAbsolutePath string
		private := true
		appManifestFilePath := "./data/nginx.yaml"
		workloadName := "nginx"
		workloadNamespace := "my-nginx"
		defaultSshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		appRepoName := "wego-test-app-" + RandString(8)
		addCommand := "app add " + appRepoName + "/"
		appName := appRepoName

		defer deleteRepo(appRepoName)
		defer deleteWorkload(workloadName, workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(appRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(workloadName, workloadNamespace)
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

		By("And I run wego add command from repo parent dir", func() {
			pathToRepoParentDir := repoAbsolutePath + "/../"
			runWegoAddCommand(pathToRepoParentDir, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed("nginx", "my-nginx")
		})

		By("And repos created have private visibility", func() {
			Expect(getRepoVisibility(os.Getenv("GITHUB_ORG"), appRepoName)).Should(ContainSubstring("true"))
		})
	})

	It("Verify app repo can be added to the cluster by running 'wego add . --app-config-url=<git ssh url>' ", func() {
		var repoAbsolutePath string
		var configRepoRemoteURL string
		private := true
		appManifestFilePath := "./data/nginx.yaml"
		workloadName := "nginx"
		workloadNamespace := "my-nginx"
		defaultSshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		appRepoName := "wego-test-app-" + RandString(8)
		appConfigRepoName := "wego-config-repo-" + RandString(8)
		configRepoRemoteURL = "ssh://git@github.com/" + os.Getenv("GITHUB_ORG") + "/" + appConfigRepoName + ".git"
		addCommand := "app add . --app-config-url=" + configRepoRemoteURL
		appName := appRepoName

		defer deleteRepo(appRepoName)
		defer deleteRepo(appConfigRepoName)
		defer deleteWorkload(workloadName, workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(appRepoName)
			deleteRepo(appConfigRepoName)

		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(workloadName, workloadNamespace)
		})

		By("When I create a private repo for wego app config", func() {
			appCofigRepoAbsPath := initAndCreateEmptyRepo(appConfigRepoName, private)
			gitAddCommitPush(appCofigRepoAbsPath, appManifestFilePath)
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

		By("And I run wego add command with --app-config-url param", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed("nginx", "my-nginx")
		})
	})

	It("Verify app repo can be added to the cluster by running 'wego add --url=<app repo url> --app-config-url=<git ssh url>' ", func() {
		var repoAbsolutePath string
		var configRepoRemoteURL string
		private := true
		appManifestFilePath := "./data/nginx.yaml"
		workloadName := "nginx"
		workloadNamespace := "my-nginx"
		defaultSshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		appRepoName := "wego-test-app-" + RandString(8)
		appConfigRepoName := "wego-config-repo-" + RandString(8)
		appRepoRemoteURL := "ssh://git@github.com/" + os.Getenv("GITHUB_ORG") + "/" + appRepoName + ".git"
		configRepoRemoteURL = "ssh://git@github.com/" + os.Getenv("GITHUB_ORG") + "/" + appConfigRepoName + ".git"
		addCommand := "app add --url=" + appRepoRemoteURL + " --app-config-url=" + configRepoRemoteURL
		appName := appRepoName

		defer deleteRepo(appRepoName)
		defer deleteRepo(appConfigRepoName)
		defer deleteWorkload(workloadName, workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(appRepoName)
			deleteRepo(appConfigRepoName)

		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(workloadName, workloadNamespace)
		})

		By("When I create a private repo for wego app config", func() {
			appCofigRepoAbsPath := initAndCreateEmptyRepo(appConfigRepoName, private)
			gitAddCommitPush(appCofigRepoAbsPath, appManifestFilePath)
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

		By("And I run wego add command with --url and --app-config-url params", func() {
			runWegoAddCommand(repoAbsolutePath+"/../", addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed("nginx", "my-nginx")
		})
	})

	It("Verify app repo can be added to the cluster by running 'wego add --url=<app repo url>' ", func() {
		var repoAbsolutePath string
		private := false
		appManifestFilePath := "./data/nginx.yaml"
		workloadName := "nginx"
		workloadNamespace := "my-nginx"
		defaultSshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		appRepoName := "wego-test-app-" + RandString(8)
		appRepoRemoteURL := "ssh://git@github.com/" + os.Getenv("GITHUB_ORG") + "/" + appRepoName + ".git"
		addCommand := "app add --url=" + appRepoRemoteURL
		appName := appRepoName

		defer deleteRepo(appRepoName)
		defer deleteWorkload(workloadName, workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(appRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(workloadName, workloadNamespace)
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

		By("And I run wego add command with --url", func() {
			runWegoAddCommand(repoAbsolutePath+"/../", addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed("nginx", "my-nginx")
		})
	})

	It("Verify that wego can deploy multiple workloads from a single app repo", func() {
		var repoAbsolutePath string
		appManifestFilePath1 := "./data/nginx.yaml"
		appManifestFilePath2 := "./data/nginx2.yaml"
		workloadName1 := "nginx"
		workloadName2 := "nginx2"
		workloadNamespace1 := "my-nginx"
		workloadNamespace2 := "my-nginx2"
		defaultSshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa"
		appRepoName := "wego-test-app-" + RandString(8)
		appName := appRepoName
		addCommand := "app add . --name=" + appName

		defer deleteRepo(appRepoName)
		defer deleteWorkload(workloadName1, workloadNamespace1)
		defer deleteWorkload(workloadName2, workloadNamespace2)

		By("And application repos do not already exist", func() {
			deleteRepo(appRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deleteWorkload(workloadName1, workloadNamespace1)
			deleteWorkload(workloadName2, workloadNamespace2)
		})

		By("When I create an empty private repo for app1", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, true)
		})

		By("And I git add-commit-push for app with multiple workloads", func() {
			gitAddCommitPush(repoAbsolutePath, appManifestFilePath1)
			gitAddCommitPush(repoAbsolutePath, appManifestFilePath2)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+defaultSshKeyPath, func() {
			setupSSHKey(defaultSshKeyPath)
		})

		By("And I run wego add command for 1st app", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see wego add command linked the repo  to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I should see workload for app1 is deployed to the cluster", func() {
			verifyWorkloadIsDeployed(workloadName1, workloadNamespace1)
			verifyWorkloadIsDeployed(workloadName2, workloadNamespace2)
		})
	})

})
