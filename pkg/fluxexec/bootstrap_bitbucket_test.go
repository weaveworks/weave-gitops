package fluxexec

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("bootstrapBitBucketServerCmd", func() {
	It("should be able to generate correct bootstrap command", func() {
		By("generating the default command", func() {
			flux, err := NewFlux(".", "/path/to/flux")
			Expect(err).To(BeNil())

			initCmd, err := flux.bootstrapBitbucketServerCmd(context.TODO())
			Expect(err).To(BeNil())
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"bootstrap",
				"bitbucket-server",
			}))
		})

		By("generating the command with bit-bucket group", func() {
			flux, err := NewFlux(".", "/path/to/flux")
			Expect(err).To(BeNil())

			initCmd, err := flux.bootstrapBitbucketServerCmd(context.TODO(),
				WithGlobalOptions(
					Namespace("weave-gitops-system"),
				),
				WithBootstrapOptions(
					NetworkPolicy(false),
					SecretName("weave-gitops-system"),
				),
				Group("group1", "group2"),
			)
			Expect(err).To(BeNil())
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"bootstrap",
				"bitbucket-server",
				"--namespace",
				"weave-gitops-system",
				"--network-policy", "false",
				"--secret-name", "weave-gitops-system",
				"--group", "group1,group2",
			}))
		})

	})

})
