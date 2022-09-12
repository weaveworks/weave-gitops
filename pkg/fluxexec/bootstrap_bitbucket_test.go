package fluxexec

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"os"
	"path/filepath"
)

var _ = Describe("bootstrapBitBucketServerCmd", func() {
	It("should be able to generate correct bootstrap command", func() {
		By("generating the default command", func() {
			flux, err := NewFlux(".", "/path/to/flux")
			Expect(err).To(BeNil())

			homedir, err := os.UserHomeDir()
			Expect(err).To(BeNil())

			initCmd, err := flux.bootstrapBitbucketServerCmd(context.TODO())
			Expect(err).To(BeNil())
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"bootstrap",
				"bitbucket-server",
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
				"--interval",
				"1m0s",
				"--private",
				"--username",
				"git",
			}))
		})

		By("generating the command with bit-bucket group", func() {
			flux, err := NewFlux(".", "/path/to/flux")
			Expect(err).To(BeNil())

			homedir, err := os.UserHomeDir()
			Expect(err).To(BeNil())

			initCmd, err := flux.bootstrapBitbucketServerCmd(context.TODO(), Group("group1", "group2"))
			Expect(err).To(BeNil())
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"bootstrap",
				"bitbucket-server",
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
				"--group",
				"group1,group2",
				"--interval",
				"1m0s",
				"--private",
				"--username",
				"git",
			}))
		})

	})

})
