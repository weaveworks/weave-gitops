/**
* All tests related to 'wego add' will go into this file
 */

package acceptance

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
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

	It("Verify that wego cannot work without wego components installed in the cluster", func() {
		var repoAbsolutePath string
		var addCommandErr string
		var addCommandOut string
		private := true
		tip := generateTestInputs()

		addCommand := "app add . --auto-merge=true"

		defer deleteRepo(tip.appRepoName)

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

		By("And WeGO runtime is not installed", func() {
			uninstallWegoRuntime(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego add command", func() {
			addCommandOut, addCommandErr = runWegoAddCommandWithOutput(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see relevant message in the console", func() {
			Eventually(addCommandOut).Should(MatchRegexp(`✔ No flux or wego installed`))
			Eventually(addCommandErr).Should(ContainSubstring("Wego not installed... exiting"))
		})
	})

	It("Verify that wego cannot work with both url and directory provided", func() {
		var repoAbsolutePath string
		var errOutput string
		tip := generateTestInputs()
		appRepoRemoteURL := "ssh://git@github.com/" + GITHUB_ORG + "/" + tip.appRepoName + ".git"

		addCommand := "app add . --url=" + appRepoRemoteURL + " --auto-merge=true"

		defer deleteRepo(tip.appRepoName)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, true)
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I run add command with both directory path and url specified", func() {
			_, errOutput = runWegoAddCommandWithOutput(repoAbsolutePath, addCommand, "wego-system")
		})

		By("Then I should see an error", func() {
			Expect(errOutput).To(ContainSubstring("you should choose either --url or the app directory"))
		})
	})

	//deployment-type=default k | repo=private | url=giturl | branch| dry-run | app-config-url=""
	It("SmokeTest - Verify that wego does not modify the cluster when run with --dry-run flag", func() {
		var repoAbsolutePath string
		var addCommandOutput string
		private := true
		tip := generateTestInputs()
		branchName := "test-branch-01"
		appRepoRemoteURL := "ssh://git@github.com/" + GITHUB_ORG + "/" + tip.appRepoName + ".git"
		appName := tip.appRepoName
		appType := "Kustomization"

		fmt.Printf("DEFAULT_SSH_KEY_PATH[%s]\n", DEFAULT_SSH_KEY_PATH)
		fmt.Printf("SSH_AUTH_SOCK[%s]\n", os.Getenv("SSH_AUTH_SOCK"))
		c := exec.Command("ls", "-lha", os.Getenv("HOME")+"/.ssh")
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		err := c.Run()
		fmt.Println("error on command: ", err)
		body, err := ioutil.ReadFile(DEFAULT_SSH_KEY_PATH)
		fmt.Println("Error reading ssh key file", err)
		fmt.Println("Content:", string(body))
		addCommand := "app add --url=" + appRepoRemoteURL + " --branch=" + branchName + " --private-key=" + DEFAULT_SSH_KEY_PATH + " --dry-run" + " --auto-merge=true"

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

	//deployment-type=default k | repo=private, initially empty | app-config-url=""
	It("Verify that wego can deploy an app after it is setup with an empty repo initially", func() {
		var repoAbsolutePath string
		private := true
		tip := generateTestInputs()
		appName := tip.appRepoName

		addCommand := "app add . --auto-merge=true"

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

	//deployment-type=default k | repo=private | url=giturl | branch | namespace | private-key=~/ | app-config-url=NONE
	// Eventually this test run will include all the remaining un-automated `wego app add` flags.
	It("SmokeTest - Verify that wego can deploy app when user specifies branch, namespace, url, private-key, deployment-type", func() {
		var repoAbsolutePath string
		private := true
		DEFAULT_SSH_KEY_PATH := "~/.ssh/id_rsa"
		tip := generateTestInputs()
		branchName := "test-branch-02"
		wegoNamespace := "my-space"
		appName := tip.appRepoName
		appRepoRemoteURL := "ssh://git@github.com/" + GITHUB_ORG + "/" + appName + ".git"

		addCommand := "app add --url=" + appRepoRemoteURL + " --branch=" + branchName + " --namespace=" + wegoNamespace + " --deployment-type=kustomize --private-key=" + DEFAULT_SSH_KEY_PATH + " --app-config-url=NONE"

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

		By("And I create a new branch", func() {
			createGitRepoBranch(repoAbsolutePath, branchName)
		})

		By("And I run wego add command with specified branch, namespace, url, deplyment-type, private-key", func() {
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

		By("And I should not see wego components in the remote git repo", func() {
			pullGitRepo(repoAbsolutePath)
			folderOutput, _ := runCommandAndReturnStringOutput(fmt.Sprintf("cd %s && ls -al", repoAbsolutePath))
			Expect(folderOutput).ShouldNot(ContainSubstring(".wego"))
			Expect(folderOutput).ShouldNot(ContainSubstring("apps"))
			Expect(folderOutput).ShouldNot(ContainSubstring("targets"))
		})
	})

	//deployment-type=default k | repo=private | app-config-url=url
	It("Verify that wego can deploy an app with app-config-url set to <url>", func() {
		var repoAbsolutePath string
		var configRepoRemoteURL string
		private := true
		tip := generateTestInputs()
		appName := tip.appRepoName
		appConfigRepoName := "wego-config-repo-" + RandString(8)
		configRepoRemoteURL = "ssh://git@github.com/" + GITHUB_ORG + "/" + appConfigRepoName + ".git"

		addCommand := "app add . --app-config-url=" + configRepoRemoteURL + " --auto-merge=true"

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

	//deployment-type=default k | repo=private | url | app-config-url=url
	It("Verify that wego can deploy an app with specified config-url and app-config-url set to <url>", func() {
		var repoAbsolutePath string
		var configRepoRemoteURL string
		private := true
		tip := generateTestInputs()
		appName := tip.appRepoName
		appConfigRepoName := "wego-config-repo-" + RandString(8)
		appRepoRemoteURL := "ssh://git@github.com/" + GITHUB_ORG + "/" + tip.appRepoName + ".git"
		configRepoRemoteURL = "ssh://git@github.com/" + GITHUB_ORG + "/" + appConfigRepoName + ".git"

		addCommand := "app add --url=" + appRepoRemoteURL + " --app-config-url=" + configRepoRemoteURL + " --auto-merge=true"

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

	//deployment-type=default k | repo=private | app-config-url=url
	It("Verify that wego can deploy an app with specified config-url and app-config-url set to default", func() {
		var repoAbsolutePath string
		private := false
		tip := generateTestInputs()
		appName := tip.appRepoName
		appRepoRemoteURL := "ssh://git@github.com/" + GITHUB_ORG + "/" + tip.appRepoName + ".git"

		addCommand := "app add --url=" + appRepoRemoteURL + " --auto-merge=true"

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

	//deployment-type=default k | repo=private | path/to/repo/dir | app-config-url=""
	It("Verify that wego can deploy an app when provided with relative path: 'path/to/repo/dir'", func() {
		var repoAbsolutePath string
		private := true
		tip := generateTestInputs()
		appName := tip.appRepoName

		addCommand := "app add " + tip.appRepoName + "/" + " --auto-merge=true"

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

	//deployment-type=default k | repo=private | name | workload=1,2 | app-config-url=""
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

	//deployment-type=default k | repo=private | workload=1,2 | app-config-url=url
	It("Verify that wego can add multiple apps dir to the cluster using single repo for wego config", func() {
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
		appName1 := appRepoName1
		appName2 := appRepoName2

		addCommand := "app add . --app-config-url=" + configRepoRemoteURL + " --auto-merge=true"

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

	//deployment-type=default k | repo=private | workload=1,2 | app-config-url=url
	It("Verify that wego can add multiple apps dir to the cluster using single app and wego config repo", func() {
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

	//deployment-type=default k | repo=private + public | app status | app list | app-config-url="" | e2e
	It("SmokeTest - Verify that wego can deploy multiple apps one with private and other with public repo", func() {
		var listOutput string
		var appStatus1 *gexec.Session
		var appStatus2 *gexec.Session
		var repoAbsolutePath1 string
		var repoAbsolutePath2 string
		tip1 := generateTestInputs()
		tip2 := generateTestInputs()
		appName1 := tip1.appRepoName
		appName2 := tip2.appRepoName
		private := true
		public := false

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
			repoAbsolutePath1 = initAndCreateEmptyRepo(tip1.appRepoName, private)
		})

		By("When I create an empty private repo for app2", func() {
			repoAbsolutePath2 = initAndCreateEmptyRepo(tip2.appRepoName, public)
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

		By("And repos created have proper visibility", func() {
			Eventually(getRepoVisibility(GITHUB_ORG, tip1.appRepoName)).Should(ContainSubstring("true"))
			Eventually(getRepoVisibility(GITHUB_ORG, tip2.appRepoName)).Should(ContainSubstring("false"))
		})

		By("When I check the app status for "+appName1, func() {
			appStatus1 = runCommandAndReturnSessionOutput(fmt.Sprintf("%s app status %s", WEGO_BIN_PATH, appName1))
		})

		By("Then I should see the status for "+appName1, func() {
			Eventually(appStatus1).Should(gbytes.Say(`Last successful reconciliation:`))
			Eventually(appStatus1).Should(gbytes.Say(`gitrepository/` + appName1))
			Eventually(appStatus1).Should(gbytes.Say(`kustomization/` + appName1))
		})

		By("When I check the app status for "+appName2, func() {
			appStatus2 = runCommandAndReturnSessionOutput(fmt.Sprintf("%s app status %s", WEGO_BIN_PATH, appName2))
		})

		By("Then I should see the status for "+appName2, func() {
			Eventually(appStatus2).Should(gbytes.Say(`Last successful reconciliation:`))
			Eventually(appStatus2).Should(gbytes.Say(`gitrepository/` + appName2))
			Eventually(appStatus2).Should(gbytes.Say(`kustomization/` + appName2))
		})

		By("When I check for apps list", func() {
			listOutput, _ = runCommandAndReturnStringOutput(WEGO_BIN_PATH + " app list")
		})

		By("Then I should see appNames for both apps listed", func() {
			Eventually(listOutput).Should(ContainSubstring(appName1))
			Eventually(listOutput).Should(ContainSubstring(appName2))
		})
	})

	//deployment-type=helm | repo=private | path=helmchart | name=appname | app-config-url=NONE
	It("SmokeTest - Verify that wego can deploy a helm app from a git repo with app-config-url set to NONE", func() {
		var repoAbsolutePath string
		private := true
		appManifestFilePath := "./data/helm-repo/hello-world"
		appName := "my-helm-app"
		appRepoName := "wego-test-app-" + RandString(8)

		addCommand := "app add . --deployment-type=helm --path=./hello-world --name=" + appName + " --app-config-url=NONE"

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

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			Expect(waitForResource("apps", appName, WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
			Expect(waitForResource("configmaps", "helloworld-configmap", WEGO_DEFAULT_NAMESPACE, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
		})

		By("And I should not see wego components in the remote git repo", func() {
			pullGitRepo(repoAbsolutePath)
			folderOutput, _ := runCommandAndReturnStringOutput(fmt.Sprintf("cd %s && ls -al", repoAbsolutePath))
			Expect(folderOutput).ShouldNot(ContainSubstring(".wego"))
			Expect(folderOutput).ShouldNot(ContainSubstring("apps"))
			Expect(folderOutput).ShouldNot(ContainSubstring("targets"))
		})
	})

	//deployment-type=helm | repo=public | path=helmchart | name=appname | app-config-url=""
	It("SmokeTest - Verify that wego can deploy a helm app from a git repo with app-config-url set to default", func() {
		var repoAbsolutePath string
		public := false
		appName := "my-helm-app"
		appManifestFilePath := "./data/helm-repo/hello-world"
		appRepoName := "wego-test-app-" + RandString(8)

		addCommand := "app add . --deployment-type=helm --path=./hello-world --name=" + appName + " --auto-merge=true"

		defer deleteRepo(appRepoName)

		By("And application repo does not already exist", func() {
			deleteRepo(appRepoName)
		})

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, public)
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

		By("And repo created has public visibility", func() {
			Eventually(getRepoVisibility(GITHUB_ORG, appRepoName)).Should(ContainSubstring("false"))
		})
	})

	//deployment-type=helm | repo=private | path=helmchart | name=appname | app-config-url=url
	It("Verify that wego can deploy a helm app from a git repo with app-config-url set to <url>", func() {
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

	//deployment-type=default h | repo=private | url=helmrepo | chart=helmchart | app-config-url=url
	It("SmokeTest - Verify that wego can deploy multiple helm apps from a helm repo with app-config-url set to <url>", func() {
		var repoAbsolutePath string
		var listOutput string
		var appStatus1 *gexec.Session
		var appStatus2 *gexec.Session
		private := true
		appName1 := "rabbitmq"
		appName2 := "zookeeper"
		workloadName1 := "rabbitmq-0"
		workloadName2 := "zookeeper-0"
		readmeFilePath := "./data/README.md"
		appRepoName := "wego-test-app-" + RandString(8)
		appRepoRemoteURL := "ssh://git@github.com/" + GITHUB_ORG + "/" + appRepoName + ".git"
		helmRepoURL := "https://charts.bitnami.com/bitnami"

		addCommand1 := "app add --url=" + helmRepoURL + " --chart=" + appName1 + " --app-config-url=" + appRepoRemoteURL + " --auto-merge=true"
		addCommand2 := "app add --url=" + helmRepoURL + " --chart=" + appName2 + " --app-config-url=" + appRepoRemoteURL + " --auto-merge=true"

		defer deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName1, EVENTUALLY_DEFAULT_TIME_OUT)
		defer deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName2, EVENTUALLY_DEFAULT_TIME_OUT)
		defer deleteRepo(appRepoName)

		By("And application repo does not already exist", func() {
			deleteRepo(appRepoName)
		})

		By("And application workload is not already deployed to cluster", func() {
			deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName1, EVENTUALLY_DEFAULT_TIME_OUT)
			deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName2, EVENTUALLY_DEFAULT_TIME_OUT)
		})

		By("When I create a private git repo", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, readmeFilePath)
		})

		By("And I install wego under my namespace: "+WEGO_DEFAULT_NAMESPACE, func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("And I run wego add command for 1st app", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand1, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I run wego add command for 2nd app", func() {
			runWegoAddCommand(repoAbsolutePath, addCommand2, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see workload1 deployed to the cluster", func() {
			verifyWegoHelmAddCommand(appName1, WEGO_DEFAULT_NAMESPACE)
			verifyHelmPodWorkloadIsDeployed(workloadName1, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I should see workload2 deployed to the cluster", func() {
			verifyWegoHelmAddCommand(appName2, WEGO_DEFAULT_NAMESPACE)
			verifyHelmPodWorkloadIsDeployed(workloadName2, WEGO_DEFAULT_NAMESPACE)
		})

		By("And I should see wego components in the remote git repo", func() {
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
			appStatus1 = runCommandAndReturnSessionOutput(fmt.Sprintf("%s app status %s", WEGO_BIN_PATH, appName1))
		})

		By("Then I should see the status for "+appName1, func() {
			Eventually(appStatus1).Should(gbytes.Say(`Last successful reconciliation:`))
			Eventually(appStatus1).Should(gbytes.Say(`helmrepository/` + appName1))
			Eventually(appStatus1).Should(gbytes.Say(`helmrelease/` + appName1))
		})

		By("When I check the app status for "+appName2, func() {
			appStatus2 = runCommandAndReturnSessionOutput(fmt.Sprintf("%s app status %s", WEGO_BIN_PATH, appName2))
		})

		By("Then I should see the status for "+appName2, func() {
			Eventually(appStatus2).Should(gbytes.Say(`Last successful reconciliation:`))
			Eventually(appStatus2).Should(gbytes.Say(`helmrepository/` + appName2))
			Eventually(appStatus2).Should(gbytes.Say(`helmrelease/` + appName2))
		})
	})
	//deployment-type=default h | url=helmrepo | chart=helmchart | app-config-url=NONE
	It("Verify that wego can deploy a helm app from a helm repo with app-config-url set to NONE", func() {
		appName := "loki"
		workloadName := "loki-0"
		helmRepoURL := "https://charts.kube-ops.io"

		addCommand := "app add --url=" + helmRepoURL + " --chart=" + appName + " --app-config-url=NONE"

		defer deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName, EVENTUALLY_DEFAULT_TIME_OUT)

		By("And application workload is not already deployed to cluster", func() {
			deletePersistingHelmApp(WEGO_DEFAULT_NAMESPACE, workloadName, EVENTUALLY_DEFAULT_TIME_OUT)
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

	It("Verify that a PR is raised against a user repo when skipping auto-merge", func() {
		var repoAbsolutePath string
		tip := generateTestInputs()
		appName := tip.appRepoName
		prLink := ""

		addCommand := "app add . --name=" + appName + " --auto-merge=false"

		defer deleteRepo(tip.appRepoName)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
		})

		By("When I create an empty private repo for app", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(tip.appRepoName, true)
		})

		By("And I git add-commit-push app manifest", func() {
			gitAddCommitPush(repoAbsolutePath, tip.appManifestFilePath)
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I have my default ssh key on path "+DEFAULT_SSH_KEY_PATH, func() {
			setupSSHKey(DEFAULT_SSH_KEY_PATH)
		})

		By("When I run wego app add command for app", func() {
			output, _ := runWegoAddCommandWithOutput(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
			re := regexp.MustCompile(`(http|ftp|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
			prLink = re.FindAllString(output, -1)[0]
		})

		By("Then I should see a PR created in user repo", func() {
			verifyPRCreated(repoAbsolutePath, appName)
		})

		By("When I merge the created PR", func() {
			mergePR(repoAbsolutePath, prLink)
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed(tip.workloadName, tip.workloadNamespace)
		})
	})

	It("Verify that a PR can be raised against an external repo with app-config-url set to <url>", func() {
		var repoAbsolutePath string
		var configRepoRemoteURL string
		var appConfigRepoAbsPath string
		prLink := ""
		private := true
		tip := generateTestInputs()
		appName := tip.appRepoName
		appConfigRepoName := "wego-config-repo-" + RandString(8)
		configRepoRemoteURL = "ssh://git@github.com/" + GITHUB_ORG + "/" + appConfigRepoName + ".git"

		addCommand := "app add . --app-config-url=" + configRepoRemoteURL

		defer deleteRepo(tip.appRepoName)
		defer deleteRepo(appConfigRepoName)
		defer deleteWorkload(tip.workloadName, tip.workloadNamespace)

		By("And application repo does not already exist", func() {
			deleteRepo(tip.appRepoName)
			deleteRepo(appConfigRepoName)
		})

		By("When I create a private repo for wego app config", func() {
			appConfigRepoAbsPath = initAndCreateEmptyRepo(appConfigRepoName, private)
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
			output, _ := runWegoAddCommandWithOutput(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
			re := regexp.MustCompile(`(http|https):\/\/([\w\-_]+(?:(?:\.[\w\-_]+)+))([\w\-\.,@?^=%&amp;:/~\+#]*[\w\-\@?^=%&amp;/~\+#])?`)
			prLink = re.FindAllString(output, 1)[0]
		})

		By("Then I should see a PR created for external repo", func() {
			verifyPRCreated(appConfigRepoAbsPath, appName)
		})

		By("When I merge the created PR", func() {
			mergePR(appConfigRepoAbsPath, prLink)
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed(tip.workloadName, tip.workloadNamespace)
		})
	})

	It("Verify that a PR fails when raised against the same app-repo with different branch and app", func() {
		var repoAbsolutePath string
		tip := generateTestInputs()
		tip2 := generateTestInputs()
		appName := tip.appRepoName
		appName2 := tip2.appRepoName
		prLink := "https://github.com/" + GITHUB_ORG + "/" + tip.appRepoName + "/pull/1"

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

		By("And I run app add command for "+appName, func() {
			runWegoAddCommand(repoAbsolutePath, addCommand, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see a PR created for "+appName, func() {
			verifyPRCreated(repoAbsolutePath, appName)
		})

		By("And I should fail to create a PR with the same app repo consecutively", func() {
			_, addCommandErr := runWegoAddCommandWithOutput(repoAbsolutePath, addCommand2, WEGO_DEFAULT_NAMESPACE)
			Expect(addCommandErr).Should(ContainSubstring("422 Reference already exists"))
		})

		By("When I merge the previous PR", func() {
			mergePR(repoAbsolutePath, prLink)
		})

		By("Then I should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appName, WEGO_DEFAULT_NAMESPACE)
			verifyWorkloadIsDeployed(tip.workloadName, tip.workloadNamespace)
		})

		By("And I should fail to create another PR with the same app", func() {
			_, addCommandErr := runWegoAddCommandWithOutput(repoAbsolutePath, addCommand2, WEGO_DEFAULT_NAMESPACE)
			Expect(addCommandErr).Should(ContainSubstring("unable to create resource, resource already exists in cluster"))
		})
	})
})
