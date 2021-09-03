/**
* All tests related to 'wego flux' will go into this file
 */

package acceptance

//import (
//	. "github.com/onsi/ginkgo"
//	. "github.com/onsi/gomega"
//	"github.com/onsi/gomega/gexec"
//)
//
//var _ = Describe("WEGO Flux Tests", func() {
//
//	var sessionOutput *gexec.Session
//
//	BeforeEach(func() {
//
//		By("Given I have a wego binary installed on my local machine", func() {
//			Expect(FileExists(WEGO_BIN_PATH)).To(BeTrue())
//		})
//	})
//
//	It("Verify that wego-flux displays error message when provided with the wrong flag", func() {
//
//		By("When I run the command 'wego flux foo'", func() {
//			sessionOutput = runCommandAndReturnSessionOutput(WEGO_BIN_PATH + " flux foo")
//		})
//
//		By("Then I should see wego error message", func() {
//			Eventually(sessionOutput.Wait().Err.Contents()).Should(ContainSubstring("✗ unknown command \"foo\" for \"flux\""))
//		})
//	})
//
//	It("Verify that wego-flux can print out the version of flux", func() {
//
//		By("When I run the command 'wego flux -v'", func() {
//			sessionOutput = runCommandAndReturnSessionOutput(WEGO_BIN_PATH + " flux -v")
//		})
//
//		By("Then I should see flux version", func() {
//			Eventually(sessionOutput.Wait().Out.Contents()).Should(MatchRegexp(`flux version 0.[0-9][0-9].[0-9]\d*`))
//		})
//	})
//})
