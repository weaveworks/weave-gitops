package fluxexec

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("bootstrapGitCmd", func() {
	It("should be able to generate correct install commands", func() {
		By("generating the default command", func() {
			flux, err := NewFlux(".", "/path/to/flux")
			Expect(err).To(BeNil())

			initCmd, err := flux.bootstrapGitCmd(context.TODO())
			Expect(err).To(BeNil())
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"bootstrap",
				"git",
			}))
		})

		By("generating the command with all options", func() {
			flux, err := NewFlux(".", "/path/to/flux")
			Expect(err).To(BeNil())

			initCmd, err := flux.bootstrapGitCmd(context.TODO(),
				WithGlobalOptions(
					Namespace("weave-gitops-system"),
				),
				WithBootstrapOptions(
					NetworkPolicy(false),
					SecretName("weave-gitops-system"),
				),
				AllowInsecureHTTP(true),
				Interval("2m"),
				Password("password"),
				Path("./"),
				Silent(true),
				URL("git@git.example.com"),
				Username("username"))
			Expect(err).To(BeNil())
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"bootstrap",
				"git",
				"--namespace",
				"weave-gitops-system",
				"--network-policy", "false",
				"--secret-name", "weave-gitops-system",
				"--allow-insecure-http",
				"--interval", "2m",
				"--password", "password",
				"--path", "./",
				"--silent",
				"--url", "git@git.example.com",
				"--username", "username",
			}))
		})
	})
})
