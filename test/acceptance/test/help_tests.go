package acceptance

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("WEGO Help Tests", func() {

	var session *gexec.Session
	var err error

	BeforeEach(func() {

		By("Given I have a wego binary installed on my local machine", func() {
			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
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

	VerifyUsageText := func() {

		By("Then I should see help message printed with the product name", func() {
			Eventually(session).Should(gbytes.Say("Weave GitOps"))
		})

		By("And Usage category", func() {
			Eventually(session).Should(gbytes.Say("Usage:"))
			Eventually(string(session.Wait().Out.Contents())).Should(ContainSubstring("wego [command]"))
		})

		By("And Avalaible-Commands category", func() {
			Eventually(session).Should(gbytes.Say("Available Commands:"))
			Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`flux[\s]+Use flux commands`))
			Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`help[\s]+Help about any command`))
			Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`version[\s]+Display wego version`))
		})

		By("And Flags category", func() {
			Eventually(session).Should(gbytes.Say("Flags:"))
			Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-h, --help[\s]+help for wego`))
			Eventually(string(session.Wait().Out.Contents())).Should(MatchRegexp(`-v, --verbose[\s]+Enable verbose output`))
		})

	}

	It("Verify that wego help flag prints the help text", func() {

		By("When I run the command 'wego --help' ", func() {
			command := exec.Command(WEGO_BIN_PATH, "--help")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		VerifyUsageText()

	})

	It("Verify that wego command prints the help text", func() {

		By("When I run the command 'wego'", func() {
			command := exec.Command(WEGO_BIN_PATH)
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		VerifyUsageText()

	})
})
