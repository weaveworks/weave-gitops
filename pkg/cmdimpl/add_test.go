package cmdimpl

import (
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/weaveworks/weave-gitops/pkg/shims"
)

var _ = Describe("Exit Path Test", func() {
	It("Verify that exit is called with expected code", func() {
		By("Executing a code path that contains checkError", func() {
			exitCode := -1
			shims.WithExitHandler(localExitHandler{action: func(code int) { exitCode = code }},
				func() {
					checkError("An error message", fmt.Errorf("An error"))
				})
			Expect(exitCode).To(Equal(1))
		})
	})
})
