package gitproviders

import (
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Github Tests", func() {
	It("Verify that we can create a provider for github", func() {
		By("Invoking the creation function", func() {
			client, err := GithubProvider()
			Expect(err).To(BeNil())
			Expect(client).To(Not(BeNil()))
		})
	})
	It("Verify that we fail to create a provider for github if token not set", func() {
		By("Invoking the creation function", func() {
			tokenval := os.Getenv("GITHUB_TOKEN")
			err := os.Unsetenv("GITHUB_TOKEN")
			Expect(err).To(BeNil())
			client, err := GithubProvider()
			Expect(err).To(Not(BeNil()))
			Expect(client).To(BeNil())
			err = os.Setenv("GITHUB_TOKEN", tokenval)
			Expect(err).To(BeNil())
		})
	})
})
