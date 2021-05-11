package gitproviders

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestGitlab(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Gitlab Tests")
}

var _ = Describe("Gitlab Tests", func() {
	It("Verify that we can create a provider for gitlab", func() {
		By("Invoking the creation function", func() {
			client, err := GetGitlabProvider()
			Expect(err).To(BeNil())
			Expect(client).To(Not(BeNil()))
		})
	})
})
