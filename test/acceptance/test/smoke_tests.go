// +build smoke acceptance

/**
* All smoke tests go to this file, keep them light weight and fast.
* However these should still be end to end user facing scenarios.
* Smoke tests would run as part of full suite too hence the acceptance_tests flag.
 */
package acceptance

import (
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

		By("Given I have a wego binary installed on my local machine", func() {
			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
		})
	})

	AfterEach(func() {

	})

	It("Verify that command wego version prints the version information", func() {

		By("When I run the command 'wego version'", func() {
			command := exec.Command(WEGO_BIN_PATH, "version")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Then I should see the wego version printed in format vm.n.n with newline character", func() {
			Eventually(session).Should(gbytes.Say("Version v[0-3].[0-9].[0-9]\n"))
		})

		By("And git commit with commit id", func() {
			Eventually(session).Should(gbytes.Say("GitCommit: [a-f0-9]{7}\n"))
		})

		By("And build timestamp", func() {
			Eventually(session).Should(gbytes.Say("BuildTime: [0-9-_:]+\n"))
		})

		By("And branch name", func() {
			Eventually(session).Should(gbytes.Say("Branch: main|HEAD\n"))
		})
	})

	VerifyUsageText := func() {

		By("Then I should see help message printed with the product name", func() {
			Eventually(session).Should(gbytes.Say("Weave GitOps"))
		})

		By("And Usage category", func() {
			Eventually(session).Should(gbytes.Say("Usage:"))
			Eventually(session.Wait().Out.Contents()).Should(ContainSubstring("wego [command]"))
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

	It("Verify that wego flux can print out version information", func() {

		By("When I run 'wego flux -v", func() {
			command := exec.Command(WEGO_BIN_PATH, "flux", "-v")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())
		})

		By("Then I should see the wego flux version printed in format m.n.n with newline character", func() {
			Eventually(session).Should(gbytes.Say("Output: flux version [0-3].[0-3][0-9].[0-9]\\d*\n"))
		})
	})
})
