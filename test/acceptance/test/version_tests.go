package acceptance

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Weave GitOps Version Test", func() {

	var session *gexec.Session

	BeforeEach(func() {

		By("Given I have a gitops binary installed on my local machine", func() {
			Expect(FileExists(gitopsBinaryPath)).To(BeTrue())
		})
	})

	It("SmokeTestShort - Verify that command gitops version prints the version information", func() {

		By("When I run the command 'gitops version'", func() {
			session = runCommandAndReturnSessionOutput(gitopsBinaryPath + " version")
		})

		By("Then I should see the gitops version printed with newline character", func() {
			Eventually(session).Should(gbytes.Say("Current Version: \\S+\n"))
		})

		By("And git commit with commit id", func() {
			Eventually(session).Should(gbytes.Say("GitCommit: ([a-f0-9]{7})|([a-f0-9]{8}\n)"))
		})

		By("And build timestamp", func() {
			Eventually(session).Should(gbytes.Say("BuildTime: [0-9-_:]+\n"))
		})

		By("And flux version", func() {
			Eventually(session).Should(gbytes.Say("Flux Version: v[0-9].[0-9][0-9].[0-9]\n"))
		})
	})
})
