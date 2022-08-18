package fluxexec

import (
	"context"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("installCmd", func() {
	It("should be able to generate correct install commands", func() {
		By("generating the default command", func() {
			flux, err := NewFlux(".", "/mock/path/to/flux")
			Expect(err).To(BeNil())

			initCmd, err := flux.installCmd(context.TODO())
			Expect(err).To(BeNil())
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"install",
				"--watch-all-namespaces",
				"--cluster-domain",
				"cluster.local",
				"--log-level",
				"info",
				"--network-policy",
				"--registry",
				"ghcr.io/fluxcd",
				"--components",
				"source-controller,kustomize-controller,helm-controller,notification-controller",
			}))
		})
		By("generating the command without network policy", func() {
			flux, err := NewFlux(".", "/mock/path/to/flux")
			Expect(err).To(BeNil())

			initCmd, err := flux.installCmd(context.TODO(), NetworkPolicy(false))
			Expect(err).To(BeNil())
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"install",
				"--watch-all-namespaces",
				"--cluster-domain", "cluster.local",
				"--log-level", "info",
				"--registry", "ghcr.io/fluxcd",
				"--components", "source-controller,kustomize-controller,helm-controller,notification-controller",
			}))
		})
		By("generating the command to install only source controller", func() {
			flux, err := NewFlux(".", "/mock/path/to/flux")
			Expect(err).To(BeNil())

			initCmd, err := flux.installCmd(context.TODO(), Components(ComponentSourceController))
			Expect(err).To(BeNil())
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"install",
				"--watch-all-namespaces",
				"--cluster-domain", "cluster.local",
				"--log-level", "info",
				"--network-policy",
				"--registry", "ghcr.io/fluxcd",
				"--components", "source-controller",
			}))
		})

		By("generating the command to install extra controllers", func() {
			flux, err := NewFlux(".", "/mock/path/to/flux")
			Expect(err).To(BeNil())

			initCmd, err := flux.installCmd(context.TODO(), ComponentsExtra(ComponentImageReflectorController, ComponentImageReflectorController))
			Expect(err).To(BeNil())
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"install",
				"--watch-all-namespaces",
				"--cluster-domain",
				"cluster.local",
				"--log-level",
				"info",
				"--network-policy",
				"--registry",
				"ghcr.io/fluxcd",
				"--components",
				"source-controller,kustomize-controller,helm-controller,notification-controller",
				"--components-extra",
				"image-reflector-controller,image-reflector-controller",
			}))
		})

	})
})
