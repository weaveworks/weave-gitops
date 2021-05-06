package fluxops_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/fluxops/fluxopsfakes"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

func TestFluxInstall(t *testing.T) {
	assert := assert.New(t)

	fakeHandler := &fluxopsfakes.FakeFluxHandler{
		HandleStub: func(args string) ([]byte, error) {
			return []byte("foo"), nil
		},
	}
	fluxops.SetFluxHandler(fakeHandler)
	utils.SetCommandForEffectWithInputPipeResponse("kubectl apply -f -", nil)
	output, err := fluxops.Install("my-namespace")
	assert.Equal("foo", string(output))
	assert.NoError(err)

	args := fakeHandler.HandleArgsForCall(0)
	assert.Equal("install --namespace=my-namespace --export", args)
}
