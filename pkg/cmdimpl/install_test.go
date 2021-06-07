package cmdimpl

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/fluxops/fluxopsfakes"
	"github.com/weaveworks/weave-gitops/pkg/override"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

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
					_, err := Install(InstallParamSet{Namespace: "my-namespace"})
					Expect(err).To(BeNil())

					args := fakeHandler.HandleArgsForCall(0)
					Expect(args).To(Equal("install --namespace=my-namespace"))

					return override.Result{}
				},
				utils.OverrideIgnore(utils.CallCommandForEffectWithInputPipeOp))
		})
	})
})
