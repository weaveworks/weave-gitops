package fluxexec

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("bootstrapGitHubCmd", func() {
	It("should be able to generate correct install commands", func() {
		By("generating the default command", func() {
			flux, err := NewFlux(".", "/path/to/flux")
			Expect(err).To(BeNil())

			homedir, err := os.UserHomeDir()
			Expect(err).To(BeNil())

			initCmd, err := flux.bootstrapGitHubCmd(context.TODO())
			Expect(err).To(BeNil())
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"bootstrap",
				"github",
				"--cache-dir",
				filepath.Join(homedir, ".kube", "cache"),
				"--kube-api-burst",
				"100",
				"--kube-api-qps",
				"50",
				"--namespace",
				"flux-system",
				"--timeout",
				"5m0s",
				"--author-name",
				"Flux",
				"--branch",
				"main",
				"--cluster-domain",
				"cluster.local",
				"--components",
				"source-controller,kustomize-controller,helm-controller,notification-controller",
				"--log-level",
				"info",
				"--network-policy",
				"--registry",
				"ghcr.io/fluxcd",
				"--secret-name",
				"flux-system",
				"--ssh-ecdsa-curve",
				"p384",
				"--ssh-key-algorithm",
				"ecdsa",
				"--ssh-rsa-bits",
				"2048",
				"--watch-all-namespaces",
				"--hostname",
				"github.com",
				"--interval",
				"1m0s",
				"--private",
			}))
		})

		By("generating the command without network-policy, without private, but with token-auth", func() {
			flux, err := NewFlux(".", "/path/to/flux")
			Expect(err).To(BeNil())

			homedir, err := os.UserHomeDir()
			Expect(err).To(BeNil())

			initCmd, err := flux.bootstrapGitHubCmd(context.TODO(),
				WithBootstrapOptions(
					TokenAuth(true),
					NetworkPolicy(false)),
				Private(false))
			Expect(err).To(BeNil())
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"bootstrap",
				"github",
				"--cache-dir",
				filepath.Join(homedir, ".kube", "cache"),
				"--kube-api-burst",
				"100",
				"--kube-api-qps",
				"50",
				"--namespace",
				"flux-system",
				"--timeout",
				"5m0s",
				"--author-name",
				"Flux",
				"--branch",
				"main",
				"--cluster-domain",
				"cluster.local",
				"--components",
				"source-controller,kustomize-controller,helm-controller,notification-controller",
				"--log-level",
				"info",
				"--registry",
				"ghcr.io/fluxcd",
				"--secret-name",
				"flux-system",
				"--ssh-ecdsa-curve",
				"p384",
				"--ssh-key-algorithm",
				"ecdsa",
				"--ssh-rsa-bits",
				"2048",
				"--token-auth",
				"--watch-all-namespaces",
				"--hostname",
				"github.com",
				"--interval",
				"1m0s",
			}))
		})

		By("generating the command for the different flux namespace", func() {
			flux, err := NewFlux(".", "/path/to/flux")
			Expect(err).To(BeNil())

			homedir, err := os.UserHomeDir()
			Expect(err).To(BeNil())

			initCmd, err := flux.bootstrapGitHubCmd(context.TODO(),
				WithGlobalOptions(
					Namespace("weave-gitops-system"),
				),
				WithBootstrapOptions(
					NetworkPolicy(false),
					SecretName("weave-gitops-system"),
				),
				Private(false))
			Expect(err).To(BeNil())
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"bootstrap",
				"github",
				"--cache-dir",
				filepath.Join(homedir, ".kube", "cache"),
				"--kube-api-burst",
				"100",
				"--kube-api-qps",
				"50",
				"--namespace",
				"weave-gitops-system",
				"--timeout",
				"5m0s",
				"--author-name",
				"Flux",
				"--branch",
				"main",
				"--cluster-domain",
				"cluster.local",
				"--components",
				"source-controller,kustomize-controller,helm-controller,notification-controller",
				"--log-level",
				"info",
				"--registry",
				"ghcr.io/fluxcd",
				"--secret-name",
				"weave-gitops-system",
				"--ssh-ecdsa-curve",
				"p384",
				"--ssh-key-algorithm",
				"ecdsa",
				"--ssh-rsa-bits",
				"2048",
				"--watch-all-namespaces",
				"--hostname",
				"github.com",
				"--interval",
				"1m0s",
			}))
		})
	})
})
