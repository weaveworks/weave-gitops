/**
* All tests related to 'wego install' will go into this file
 */

package acceptance

import (
	"fmt"
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Weave GitOps Install Tests", func() {

	var session *gexec.Session
	var err error
	var namespace string

	BeforeEach(func() {
		By("Given I have a wego binary installed on my local machine", func() {
			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
		})
	})

	It("Validate that wego displays help text for 'install' command", func() {

		By("When I run the command 'wego install -h'", func() {
			command := exec.Command(WEGO_BIN_PATH, "install", "-h")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		By("Then I should see wego help text displayed for 'install' command", func() {
			Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`The install command deploys Wego in the specified namespace.`))
			Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`If a previous version is installed, then an in-place upgrade will be performed.`))
			Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`Usage:`))
			Eventually(string(session.Wait().Out.Contents())).Should(ContainSubstring("wego install [flags]"))
			Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`Examples:`))
			Eventually(string(session.Wait().Out.Contents())).Should(ContainSubstring("# Install wego in the wego-system namespace"))
			Eventually(string(session.Wait().Out.Contents())).Should(ContainSubstring("wego install"))
			Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`Flags:`))
			Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-h, --help[\s]+help for install`))
			Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`Global Flags`))
			Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`--namespace string[\s]+gitops runtime namespace \(default "wego-system"\)`))
			Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-v, --verbose[\s]+Enable verbose output`))
		})

	})

	It("Verify that wego can install required controllers under default namespace `wego-system`", func() {

		By("And I have a brand new cluster", func() {
			_, err := ResetOrCreateCluster(WEGO_DEFAULT_NAMESPACE)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("When I run 'wego install' command with default namespace", func() {
			command := exec.Command("sh", "-c", fmt.Sprintf("%s install", WEGO_BIN_PATH))
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		verifyControllersInCluster(WEGO_DEFAULT_NAMESPACE)
	})

	It("Verify that wego can add flux controllers to a user-specified namespace", func() {

		namespace = "test-namespace"

		By("And I have a brand new cluster", func() {
			_, err := ResetOrCreateCluster(namespace)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("And I create a namespace for my controllers", func() {
			createNamespace(namespace)
		})

		By("When I run 'wego install' command with specified namespace", func() {
			command := exec.Command("sh", "-c", fmt.Sprintf("%s install --namespace %s", WEGO_BIN_PATH, namespace))
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		verifyControllersInCluster(namespace)

		By("Clean up the namespace", func() {
			_, err := ResetOrCreateCluster(namespace)
			Expect(err).ShouldNot(HaveOccurred())
		})
	})

	It("Verify that wego installation quits if flux-system namespace is present", func() {

		namespace = "flux-system"

		By("And I have a brand new cluster", func() {
			_, err := ResetOrCreateCluster(namespace)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("And I create a namespace for my controllers", func() {
			createNamespace(namespace)
		})

		By("When I run wego install", func() {
			command := exec.Command("sh", "-c", fmt.Sprintf("%s install", WEGO_BIN_PATH))
			session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		By("Then I should see wego install quit message", func() {
			Eventually(session.Err).Should(gbytes.Say(`Weave GitOps does not yet support installation onto a cluster that is using Flux.\nPlease uninstall flux before proceeding:\n\s*. flux uninstall`))
		})
	})

})
