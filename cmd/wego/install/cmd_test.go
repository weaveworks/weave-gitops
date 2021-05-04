package install

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/fluxops/fluxopsfakes"
)

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
