/**
* All tests related to 'wego install' will go into this file
 */

package acceptance

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Weave GitOps Install Tests", func() {

	BeforeEach(func() {
		By("Given I have a wego binary installed on my local machine", func() {
			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
		})
	})

	It("Validate that wego displays help text for 'install' command", func() {

		var session *gexec.Session
		var err error

		By("When I run the command 'wego gitops install -h'", func() {
			command := exec.Command(WEGO_BIN_PATH, "gitops", "install", "-h")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		By("Then I should see wego help text displayed for 'install' command", func() {
			Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(
				`The install command deploys Wego in the specified namespace.\nIf a previous version is installed, then an in-place upgrade will be performed.\n*Usage:\n\s*wego gitops install \[flags]\n*Examples:\n\s*# Install wego in the wego-system namespace\n\s*wego gitops install\n*Flags:\n\s*-h, --help\s*help for install\n*Global Flags:\n\s*--dry-run\s*outputs all the manifests that would be installed\n\s*-n, --namespace string\s*the namespace scope for this operation \(default "wego-system"\)\n\s*-v, --verbose\s*Enable verbose output`))
		})
	})

	It("Verify that wego can install & uninstall wego components under default namespace `wego-system`", func() {

		By("And I have a brand new cluster", func() {
			_, err := ResetOrCreateCluster(WEGO_DEFAULT_NAMESPACE, true)
			Expect(err).ShouldNot(HaveOccurred())
		})

		installAndVerifyWego(WEGO_DEFAULT_NAMESPACE)

		By("When I run 'wego gitops uninstall' command", func() {
			command := exec.Command("sh", "-c", fmt.Sprintf("%s gitops uninstall --namespace %s", WEGO_BIN_PATH, WEGO_DEFAULT_NAMESPACE))
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		waitForNamespaceToTerminate(WEGO_DEFAULT_NAMESPACE, TIMEOUT_TWO_MINUTES)

		By("Then I should not see any wego components", func() {
			_, errOutput := runCommandAndReturnOutput("kubectl get ns " + WEGO_DEFAULT_NAMESPACE)
			Eventually(errOutput).Should(ContainSubstring(`Error from server (NotFound): namespaces "` + WEGO_DEFAULT_NAMESPACE + `" not found`))
		})
	})

	It("Verify that wego can install & uninstall wego components under a user-specified namespace", func() {

		namespace := "test-namespace"

		By("And I have a brand new cluster", func() {
			_, err := ResetOrCreateCluster(namespace, true)
			Expect(err).ShouldNot(HaveOccurred())
		})

		installAndVerifyWego(namespace)

		By("When I run 'wego gitops uninstall' command", func() {
			command := exec.Command("sh", "-c", fmt.Sprintf("%s gitops uninstall --namespace %s", WEGO_BIN_PATH, namespace))
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		waitForNamespaceToTerminate(namespace, TIMEOUT_TWO_MINUTES)

		By("Then I should not see any wego components", func() {
			_, errOutput := runCommandAndReturnOutput("kubectl get ns " + namespace)
			Eventually(errOutput).Should(ContainSubstring(`Error from server (NotFound): namespaces "` + namespace + `" not found`))
		})
	})
})
