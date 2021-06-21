package gitops

import (
	"fmt"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops/pkg/cmdimpl"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/fluxops/fluxopsfakes"
	"github.com/weaveworks/weave-gitops/pkg/override"
	"github.com/weaveworks/weave-gitops/pkg/shims"
	"github.com/weaveworks/weave-gitops/pkg/utils"
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

var _ = Describe("Run Command Test", func() {
	It("Verify path through flux commands", func() {
		By("Mocking the result", func() {
			fakeHandler := &fluxopsfakes.FakeFluxHandler{
				HandleStub: func(args string) ([]byte, error) {
					return []byte("manifests"), nil
				},
			}
			fluxops.SetFluxHandler(fakeHandler)

			_ = override.WithOverrides(
				func() override.Result {
					params = cmdimpl.InstallParamSet{
						Namespace: "my-namespace",
					}
					runCmd(&cobra.Command{}, []string{})

					args := fakeHandler.HandleArgsForCall(0)
					Expect(args).To(Equal("install --namespace=my-namespace --components-extra=image-reflector-controller,image-automation-controller"))

					return override.Result{}
				},
				utils.OverrideIgnore(utils.CallCommandForEffectWithInputPipeOp),
				utils.OverrideBehavior(utils.CallCommandSilentlyOp,
					func(args ...interface{}) ([]byte, []byte, error) {
						return []byte("not found"), []byte("not found"), fmt.Errorf("exit 1")
					},
				),
			)
		})
	})
})

var _ = Describe("Exit Path Test", func() {
	It("Verify that exit is called with expected code", func() {
		By("Executing a code path that contains checkError", func() {
			exitCode := -1
			_ = override.WithOverrides(
				func() override.Result {
					checkError("An error message", fmt.Errorf("An error"))
					return override.Result{}
				},
				shims.OverrideExit(localExitHandler{action: func(code int) { exitCode = code }}))
			Expect(exitCode).To(Equal(1))
		})
	})
})
