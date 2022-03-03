package profiles_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProfile(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Profiles Suite")
}
