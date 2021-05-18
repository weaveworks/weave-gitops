package cmdimpl

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
