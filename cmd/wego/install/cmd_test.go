package install

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/fluxops/fluxopsfakes"
)

func TestRunCmdError(t *testing.T) {
	assert := assert.New(t)

	fakeHandler := &fluxopsfakes.FakeFluxHandler{
		HandleStub: func(args string) ([]byte, error) {
			return []byte(""), fmt.Errorf("foo")
		},
	}
	fluxops.SetFluxHandler(fakeHandler)

	_, err := fluxops.Install("my-namespace")
	assert.EqualError(err, "foo")
}
