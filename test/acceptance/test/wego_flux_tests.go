// +build smoke acceptance

package test

import (
	"os/exec"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

var _ = Describe("WEGO Acceptance Tests", func() {

	var session *gexec.Session
	var err error

	BeforeEach(func() {

		By("Given I have a wego binary installed on my local machine", func() {
			Expect(utils.Exists(WEGO_BIN_PATH)).To(BeTrue())
		})
	})

	It("Verify that wego-flux displays error message when provided with the wrong flag", func() {

		By("When I run the command 'wego flux foo'", func() {
			command := exec.Command(WEGO_BIN_PATH, "flux", "foo")
			session, err = gexec.Start(command, GinkgoWriter, GinkgoWriter)
			Expect(err).ShouldNot(HaveOccurred())

		})

		By("Then I should see wego error message", func() {
			Eventually(session.Wait().Out.Contents()).Should(ContainSubstring("✗ unknown command \"foo\" for \"flux\""))
		})
	})
})
