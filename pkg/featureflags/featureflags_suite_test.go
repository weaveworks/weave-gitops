package featureflags_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestFeatureflags(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Featureflags Suite")
}
