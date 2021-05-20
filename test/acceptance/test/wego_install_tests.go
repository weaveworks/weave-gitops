// +build smoke acceptance

package test

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

func VerifyControllersInCluster(session *gexec.Session) {

	By(" Then I should see flux controllers present in the cluster", func() {
		Eventually(string(session.Wait().Out.Contents())).Should(ContainSubstring("helm-controller"))
		Eventually(string(session.Wait().Out.Contents())).Should(ContainSubstring("kustomize-controller"))
		Eventually(string(session.Wait().Out.Contents())).Should(ContainSubstring("notification-controller"))
		Eventually(string(session.Wait().Out.Contents())).Should(ContainSubstring("source-controller"))
	})
}

//Reseting namespace is an expensive operation, only use this when absolutely necessary
func ResetNamespace(namespace string) {
	By("And there's no previous wego installation", func() {
		command := exec.Command("sh", "-c", fmt.Sprintf("%s install --namespace %s| kubectl --ignore-not-found=true delete -f -", WEGO_BIN_PATH, namespace))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, INSTALL_RESET_TIMEOUT).Should(gexec.Exit())
	})
}

var _ = Describe("WEGO Acceptance Tests", func() {

	var session *gexec.Session
	var err error

	BeforeEach(func() {
		By("Given I have a wego binary installed on my local machine", func() {
			Expect(utils.Exists(WEGO_BIN_PATH)).To(BeTrue())
		})
	})

	It("Verify that wego displays error message when provided with the wrong flag", func() {

		By("When I run 'wego foo'", func() {
			command := exec.Command(WEGO_BIN_PATH, "foo")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Then I should see wego error message", func() {
			Eventually(session.Err).Should(gbytes.Say("Error: unknown command \"foo\" for \"wego\""))
			Eventually(session.Err).Should(gbytes.Say("Run 'wego --help' for usage."))
		})
	})

	It("Verify that wego can install required controllers under default namespace `wego-system`", func() {

		By("And I have a brand new cluster", func() {
			_, err := ResetOrCreateCluster()
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("When I run the command 'wego install | kubectl apply -f -'", func() {
			command := exec.Command("sh", "-c", fmt.Sprintf("%s install | kubectl apply -f -", WEGO_BIN_PATH))
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		By("And I search for the controllers with 'kubectl'", func() {
			command := exec.Command("kubectl", "get", "deploy", "-n", "wego-system")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		VerifyControllersInCluster(session)
	})

	It("Validate wego can add flux controllers with specified namespace", func() {

		namespace := "test-namespace"

		By("And I have a brand new cluster", func() {
			_, err := ResetOrCreateCluster()
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("And I create a namespace for my controllers", func() {
			command := exec.Command("kubectl", "create", "namespace", namespace)
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		By("When I run 'wego install --namespace test-namespace' command with specified namespace", func() {
			command := exec.Command("sh", "-c", fmt.Sprintf("%s install --namespace %s | kubectl apply -f -", WEGO_BIN_PATH, namespace))
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		By("When I search for the controllers with 'kubectl'", func() {
			command := exec.Command("kubectl", "get", "deploy", "-n", namespace)
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		VerifyControllersInCluster(session)
	})

})
