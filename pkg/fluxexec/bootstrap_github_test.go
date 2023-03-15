package fluxexec

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("bootstrapGitHubCmd", func() {
	It("should be able to generate correct install commands", func() {
		By("generating the default command", func() {
			flux, err := NewFlux(".", "/path/to/flux")
			Expect(err).To(BeNil())

			initCmd := flux.bootstrapGitHubCmd(context.TODO())
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"bootstrap",
				"github",
			}))
		})

		By("generating the command without network-policy, without private, but with token-auth", func() {
			flux, err := NewFlux(".", "/path/to/flux")
			Expect(err).To(BeNil())

			initCmd := flux.bootstrapGitHubCmd(context.TODO(),
				WithBootstrapOptions(
					TokenAuth(true),
					NetworkPolicy(false)),
				Private(false))
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"bootstrap",
				"github",
				"--network-policy", "false",
				"--token-auth",
				"--private", "false",
			}))
		})

		By("generating the command for the different flux namespace", func() {
			flux, err := NewFlux(".", "/path/to/flux")
			Expect(err).To(BeNil())

			initCmd := flux.bootstrapGitHubCmd(context.TODO(),
				WithGlobalOptions(
					Namespace("weave-gitops-system"),
				),
				WithBootstrapOptions(
					NetworkPolicy(false),
					SecretName("weave-gitops-system"),
				),
				Private(false))
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"bootstrap",
				"github",
				"--namespace",
				"weave-gitops-system",
				"--network-policy", "false",
				"--secret-name", "weave-gitops-system",
				"--private", "false",
			}))
		})
	})
})
