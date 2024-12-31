package fluxexec

import (
	"context"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("bootstrapGitLabCmd", func() {
	It("should be able to generate correct install commands", func() {
		By("generating the default command", func() {
			flux, err := NewFlux(".", "/path/to/flux")
			Expect(err).To(BeNil())

			initCmd := flux.bootstrapGitLabCmd(context.TODO(),
				WithGlobalOptions(
					Namespace("weave-gitops-system"),
				),
				WithBootstrapOptions(
					NetworkPolicy(false),
					SecretName("weave-gitops-system"),
				),
			)
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"bootstrap",
				"gitlab",
				"--namespace",
				"weave-gitops-system",
				"--network-policy", "false",
				"--secret-name", "weave-gitops-system",
			}))
		})
	})
})
