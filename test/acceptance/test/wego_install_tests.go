// +build smoke acceptance

package acceptance

import (
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	log "github.com/sirupsen/logrus"
	"os/exec"
)

var (
	err       error
	namespace string
)

func VerifyControllersInCluster(session *gexec.Session) {

	By("Then I should wait for controllers to get ready", func() {
		command := exec.Command("sh", "-c", fmt.Sprintf("%s wait --for=condition=Ready -n %s --all pod", "kubectl", namespace))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, INSTALL_RESET_TIMEOUT).Should(gexec.Exit())
	})

	By("And I search for the controllers with 'kubectl'", func() {
		command := exec.Command("kubectl", "get", "deploy", "-n", namespace)
		session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit())
	})

	By(" Then I should see flux controllers present in the cluster", func() {
		Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`helm-controller[\s]+1/1`))
		Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`kustomize-controller[\s]+1/1`))
		Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`notification-controller[\s]+1/1`))
		Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`source-controller[\s]+1/1`))
	})
}

// Resetting namespace is an expensive operation, only use this when absolutely necessary
func ResetNamespace(namespace string) {
	log.Infof("Resetting namespace for controllers...")

	By("And there's no previous wego installation", func() {
		command := exec.Command("sh", "-c", fmt.Sprintf("%s install --namespace %s| kubectl --ignore-not-found=true delete -f -", WEGO_BIN_PATH, namespace))
		session, err := gexec.Start(command, GinkgoWriter, GinkgoWriter)
		Expect(err).ShouldNot(HaveOccurred())
		Eventually(session, INSTALL_RESET_TIMEOUT).Should(gexec.Exit())
	})
}

var _ = Describe("WEGO Acceptance Tests", func() {

	var session *gexec.Session

	BeforeEach(func() {
		By("Given I have a wego binary installed on my local machine", func() {
			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
		})
	})

	It("Validate that wego displays help text for 'install' command", func() {

		By("When I run the command 'wego install -'h", func() {
			command := exec.Command(WEGO_BIN_PATH, "install", "-h")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
			Eventually(session).Should(gexec.Exit())
		})

		By("Then I should see wego help text for 'install' command", func() {
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

		namespace = "wego-system"

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

		VerifyControllersInCluster(session)
	})

	It("Verify that wego can add flux controllers to specified namespace", func() {

		namespace = "test-namespace"

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

		VerifyControllersInCluster(session)
	})

})
