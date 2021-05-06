package fluxops_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/fluxops/fluxopsfakes"
)

func TestFluxInstall(t *testing.T) {
	assert := assert.New(t)

	fakeHandler := &fluxopsfakes.FakeFluxHandler{
		HandleStub: func(args string) ([]byte, error) {
			return []byte("foo"), nil
		},
	}
	fluxops.SetFluxHandler(fakeHandler)
	output, err := fluxops.Install("flux-system")
	assert.Equal("foo", string(output))
	assert.NoError(err)

	output, err = fluxops.Install("my-namespace")
	assert.Equal("apiVersion: v1\nkind: Namespace\nmetadata:\n  name: flux-system\n---\nfoo", string(output))
	assert.NoError(err)

	args := fakeHandler.HandleArgsForCall(1)
	assert.Equal("install --namespace=my-namespace --export", args)
}
