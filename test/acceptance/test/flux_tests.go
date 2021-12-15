/**
* All tests related to 'gitops flux' will go into this file
 */

package acceptance

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("Weave GitOps Flux Tests", func() {

	var sessionOutput *gexec.Session

	BeforeEach(func() {

		By("Given I have a gitops binary installed on my local machine", func() {
			Expect(FileExists(gitopsBinaryPath)).To(BeTrue())
		})
	})

	It("Verify that gitops-flux displays error message when provided with the wrong flag", func() {

		By("When I run the command 'gitops flux foo'", func() {
			sessionOutput = runCommandAndReturnSessionOutput(gitopsBinaryPath + " flux foo")
		})

		By("Then I should see gitops error message", func() {
			Eventually(sessionOutput.Wait().Err.Contents()).Should(ContainSubstring("✗ unknown command \"foo\" for \"flux\""))
		})
	})

	It("Verify that gitops-flux can print out the version of flux", func() {

		By("When I run the command 'gitops flux -v'", func() {
			sessionOutput = runCommandAndReturnSessionOutput(gitopsBinaryPath + " flux -v")
		})

		By("Then I should see flux version", func() {
			Eventually(sessionOutput.Wait().Out.Contents()).Should(MatchRegexp(`flux version 0.[0-9][0-9].[0-9]\d*`))
		})
	})
})
