package flux_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/version"
)

func TestFlux(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Flux Suite")
}

var _ = BeforeSuite(func() {
	version.FluxVersion = "0.12.0"
})
