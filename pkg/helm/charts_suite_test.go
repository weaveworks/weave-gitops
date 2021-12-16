package helm_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCharts(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Charts Suite")
}
