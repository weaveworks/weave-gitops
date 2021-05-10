package install

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/fluxops/fluxopsfakes"
	"github.com/weaveworks/weave-gitops/pkg/shims"
)

type localExitHandler struct {
	action func(int)
}

func (h localExitHandler) Handle(code int) {
	h.action(code)
}

func TestRunCmd(t *testing.T) {
	assert := assert.New(t)

	fakeHandler := &fluxopsfakes.FakeFluxHandler{
		HandleStub: func(args string) ([]byte, error) {
			return []byte("manifests"), nil
		},
	}
	fluxops.SetFluxHandler(fakeHandler)

	params = paramSet{
		namespace: "my-namespace",
	}
	runCmd(&cobra.Command{}, []string{})

	args := fakeHandler.HandleArgsForCall(0)
	assert.Equal(args, "install --namespace=my-namespace --export")
}

var _ = Describe("Exit Path Test", func() {
	It("Verify that exit is called with expected code", func() {
		By("Executing a code path that contains checkError", func() {
			exitCode := -1
			shims.WithExitHandler(localExitHandler{action: func(code int) { exitCode = code }},
				func() { checkError("An error message", fmt.Errorf("An error")) })
			Expect(exitCode).To(Equal(1))
		})
	})
})
