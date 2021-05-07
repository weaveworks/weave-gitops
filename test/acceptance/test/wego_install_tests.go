// +build smoke acceptance

package acceptance

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("WEGO Acceptance Tests", func() {

	var session *gexec.Session
	var err error

	BeforeEach(func() {
		WEGO_BIN_PATH = "/Users/rokshana/weaveworks/weave-gitops/bin/wego"

		By("Given I have a wego binary installed on my local machine", func() {
			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
		})
	})

	It("Verify that wego displays error message when provided with the wrong flag", func() {

		By("When I run 'wego abcd'", func() {
			command := exec.Command(WEGO_BIN_PATH, "abcd")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Then I should see wego error message", func() {
			Eventually(session.Err).Should(gbytes.Say("Error: unknown command \"abcd\" for \"wego\""))
			Eventually(session.Err).Should(gbytes.Say("Run 'wego --help' for usage."))
		})
	})

	VerifyControllersInstallation := func() {

		By("Then I should see flux controllers installed", func() {
			Eventually(session).Should(gbytes.Say("deployment.apps/helm-controller created"))
			Eventually(session).Should(gbytes.Say("deployment.apps/kustomize-controller created"))
			Eventually(session).Should(gbytes.Say("deployment.apps/notification-controller created"))
			Eventually(session).Should(gbytes.Say("deployment.apps/source-controller created"))
		})
	}

	VerifyControllersInCluster := func() {

		By("And I should see flux controllers present in the cluster", func() {
			Eventually(session).Should(gbytes.Say("helm-controller"))
			Eventually(session).Should(gbytes.Say("kustomize-controller"))
			Eventually(session).Should(gbytes.Say("notification-controller"))
			Eventually(session).Should(gbytes.Say("source-controller"))
		})
	}

	It("Validate wego can add flux controllers to cluster", func() {

		By("When I run the command 'wego install'", func() {
			command := exec.Command("sh", "-c", fmt.Sprintf("%s install | kubectl apply -f -", WEGO_BIN_PATH))
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		VerifyControllersInstallation()

		By("When I search for the controllers with 'kubectl'", func() {
			command := exec.Command("kubectl", "get", "deployments", "-n", "wego-system")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		VerifyControllersInCluster()

	})

	It("Validate wego can add flux controllers with specified namespace", func() {

		By("When I create a namespace for my controllers", func() {
			command := exec.Command("kubectl", "create", "namespace", "test-namespace")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Then I should see my specified namespace created", func() {
			Eventually(session).Should(gbytes.Say("namespace/test-namespace created"))
		})

		By("When I run 'wego install' command with specified namespace", func() {
			command := exec.Command("sh", "-c", fmt.Sprintf("%s install --namespace test-namespace | kubectl apply -f -", WEGO_BIN_PATH))
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		VerifyControllersInstallation()

		By("When I search for the controllers with 'kubectl'", func() {
			command := exec.Command("kubectl", "get", "deployments", "-n", "test-namespace")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		VerifyControllersInCluster()
	})

})
