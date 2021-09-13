package manifests

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestManifests(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Manifests Suite")
}
