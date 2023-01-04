package fluxinstall

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFluxInstall(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flux Install Suite")
}
