package profiles_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestProfiles(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Profiles Suite")
}
