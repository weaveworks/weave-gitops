package cmdimpl

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/weaveworks/weave-gitops/pkg/shims"
)

type localExitHandler struct {
	action func(int)
}

func (h localExitHandler) Handle(code int) {
	h.action(code)
}

func TestFluxCmds(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Command Tests")
}

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
