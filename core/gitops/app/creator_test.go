package app

import (
	"testing"

	. "github.com/onsi/gomega"
)

const (
	testNamespace = "test-system"
)

type creatorFixture struct {
	*GomegaWithT
}

func setUpCreatorTest(t *testing.T) creatorFixture {
	return creatorFixture{
		GomegaWithT: NewGomegaWithT(t),
	}
}

