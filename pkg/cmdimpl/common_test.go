package cmdimpl

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCmdImplementations(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Command Implementations Tests")
}
