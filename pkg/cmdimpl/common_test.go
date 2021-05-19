package cmdimpl

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestFluxCmds(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Command Tests")
}
