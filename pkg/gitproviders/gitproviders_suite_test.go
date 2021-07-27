package gitproviders_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGitproviders(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gitproviders Suite")
}
