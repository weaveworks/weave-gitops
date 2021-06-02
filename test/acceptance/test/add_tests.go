// +build smoke acceptance

/**
* All tests related to 'wego install' will go into this file
 */

package acceptance

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/prometheus/common/log"
	"os"
	"os/exec"
	"strings"
)

func getClusterName() string {
	command := exec.Command("kubectl", "config", "current-context")
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session, INSTALL_PODS_READY_TIMEOUT).Should(gexec.Exit())
	return strings.TrimSuffix(string(session.Wait().Out.Contents()), "\n")
}

func deleteRepos(appRepoName string, wegoRepoName string) {
	log.Infof("Delete application repo: %s", os.Getenv("GITHUB_ORG")+"/"+appRepoName)
	_ = runCommandPassThrough([]string{}, "hub", "delete", "-y", os.Getenv("GITHUB_ORG")+"/"+appRepoName)
	log.Infof("Delete application repo: %s", os.Getenv("GITHUB_ORG")+"/"+wegoRepoName)
	_ = runCommandPassThrough([]string{}, "hub", "delete", "-y", os.Getenv("GITHUB_ORG")+"/"+wegoRepoName)
	log.Infof("Delete Repo from %s/.wego/repositories/%s", os.Getenv("HOME"), wegoRepoName)
	_ = os.RemoveAll(fmt.Sprintf("%s/.wego/repositories/%s", os.Getenv("HOME"), wegoRepoName))
}

func initAndCreateEmptyRepo(appRepoName string, IsPrivateRepo bool) string {
	repoAbsolutePath := "/tmp/" + appRepoName
	privateRepo := ""
	if IsPrivateRepo {
		privateRepo = "-p"
	}
	command := exec.Command("sh", "-c", fmt.Sprintf(`
							mkdir %s && 
							cd %s && 
							git init && 
							hub create %s %s`, repoAbsolutePath, repoAbsolutePath, os.Getenv("GITHUB_ORG")+"/"+appRepoName, privateRepo))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	return repoAbsolutePath
}

func gitAddCommitPush(repoAbsolutePath string, appManifestFilePath string) {
	command := exec.Command("sh", "-c", fmt.Sprintf(`
							cp %s %s &&
							cd %s &&
							git add . &&
							git commit -m 'add workload manifest' &&
							git push -u origin main`, appManifestFilePath, repoAbsolutePath, repoAbsolutePath))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
}

func getRepoVisibility(org string, repo string) string {
	command := exec.Command("sh", "-c", fmt.Sprintf("hub api --flat repos/%s/%s|grep -i private|cut -f2", org, repo))
	session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
	Expect(err).ShouldNot(HaveOccurred())
	Eventually(session).Should(gexec.Exit())
	visibilityStr := strings.TrimSpace(string(session.Wait().Out.Contents()))
	log.Infof("Repo visibility private=%s", visibilityStr)
	return visibilityStr
}

func setupSSHKey() {
	sshKeyPath := os.Getenv("HOME") + "/.ssh/id_rsa_wego"
	if _, err := os.Stat(sshKeyPath); os.IsNotExist(err) {
		command := exec.Command("sh", "-c", fmt.Sprintf(`
							echo "%s" >> %s &&
							chmod 0600 %s &&
							ls -la %s`, os.Getenv("GITHUB_KEY"), sshKeyPath, sshKeyPath, sshKeyPath))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit())
	}
}

func installAndVerifyWego(wegoNamespace string) {
	By("And I run 'wego install' command with namespace "+wegoNamespace, func() {
		command := exec.Command("sh", "-c", fmt.Sprintf("%s install --namespace=%s| kubectl apply -f -", WEGO_BIN_PATH, wegoNamespace))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit())
		VerifyControllersInCluster(session, wegoNamespace)
	})
}
func runWegoAddCommand(repoAbsolutePath string, private bool, wegoNamespace string) {
	var session *gexec.Session
	By("And I have my ssh key on path ~/.ssh/id_rsa_wego", func() {
		setupSSHKey()
	})

	if private {
		By("And I run `wego add . --private-key=~/.ssh/id_rsa_wego`", func() {
			command := exec.Command("sh", "-c", fmt.Sprintf("cd %s && %s add . --private-key=%s/.ssh/id_rsa_wego", repoAbsolutePath, WEGO_BIN_PATH, os.Getenv("HOME")))
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})
	} else {
		By("And I run `wego add . --private=false --private-key=~/.ssh/id_rsa_wego`", func() {
			command := exec.Command("sh", "-c", fmt.Sprintf("cd %s && %s add . --private=false --private-key=%s/.ssh/id_rsa_wego", repoAbsolutePath, WEGO_BIN_PATH, os.Getenv("HOME")))
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})
	}
}

func verifyWegoAddCommand(appRepoName string, wegoRepoName string, wegoNamespace string) {

	By("Then I should see remote wego and public app repos are created and linked to the cluster", func() {
		command := exec.Command("sh", "-c", fmt.Sprintf(" kubectl wait --for=condition=Ready --timeout=60s -n %s GitRepositories --all", wegoNamespace))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, INSTALL_PODS_READY_TIMEOUT).Should(gexec.Exit())
		Expect(waitForResource("GitRepositories", "wego", wegoNamespace, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
		Expect(waitForResource("GitRepositories", wegoRepoName+"-"+appRepoName, wegoNamespace, INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
	})

	By("And I should see nginx deployment appears in the cluster", func() {
		Expect(waitForResource("deploy", "nginx", "my-nginx", INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
		Expect(waitForResource("pods", "", "my-nginx", INSTALL_PODS_READY_TIMEOUT)).To(Succeed())
	})

	By("And I wait for the nginx pods to be ready", func() {
		command := exec.Command("sh", "-c", fmt.Sprintf("kubectl wait --for=condition=Ready --timeout=60s -n %s --all pods", "my-nginx"))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, INSTALL_PODS_READY_TIMEOUT).Should(gexec.Exit())
	})
}

var _ = Describe("Weave GitOps Add Tests", func() {
	var appRepoName string
	var wegoRepoName string

	BeforeEach(func() {

		By("Given I have a brand new cluster", func() {
			_, err := ResetOrCreateCluster(WEGO_DEFAULT_NAMESPACE)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("And I have a wego binary installed on my local machine", func() {
			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
			appRepoName = "wego-test-app-" + RandString(8)
			wegoRepoName = getClusterName() + "-wego"
		})

		By("And wego and application repos do not already exist", func() {
			deleteRepos(appRepoName, wegoRepoName)
		})
	})

	AfterEach(func() {
		By("Clean up", func() {
			deleteRepos(appRepoName, wegoRepoName)
		})
	})

	It("Verify private repo can be added to the cluster by running 'wego add . --private-key' ", func() {
		var repoAbsolutePath string
		private := true

		By("When I create a private repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, "./data/nginx.yaml")
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I run wego add command", func() {
			runWegoAddCommand(repoAbsolutePath, private, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appRepoName, wegoRepoName, WEGO_DEFAULT_NAMESPACE)
		})

		By("And repos created have private visibility", func() {
			Expect(getRepoVisibility(os.Getenv("GITHUB_ORG"), appRepoName)).Should(ContainSubstring("true"))
			Expect(getRepoVisibility(os.Getenv("GITHUB_ORG"), wegoRepoName)).Should(ContainSubstring("true"))
		})
	})

	It("Verify public repo can be added to the cluster by running 'wego add . --private=false --private-key'", func() {
		var repoAbsolutePath string
		private := false

		By("When I create a public repo with my app workload", func() {
			repoAbsolutePath = initAndCreateEmptyRepo(appRepoName, private)
			gitAddCommitPush(repoAbsolutePath, "./data/nginx.yaml")
		})

		By("And I install wego to my active cluster", func() {
			installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)
		})

		By("And I run wego add command", func() {
			runWegoAddCommand(repoAbsolutePath, private, WEGO_DEFAULT_NAMESPACE)
		})

		By("Then I should see should see my workload deployed to the cluster", func() {
			verifyWegoAddCommand(appRepoName, wegoRepoName, WEGO_DEFAULT_NAMESPACE)
		})
		By("And repos created have public visibility", func() {
			Expect(getRepoVisibility(os.Getenv("GITHUB_ORG"), appRepoName)).Should(ContainSubstring("false"))
			Expect(getRepoVisibility(os.Getenv("GITHUB_ORG"), wegoRepoName)).Should(ContainSubstring("false"))
		})
	})
})
