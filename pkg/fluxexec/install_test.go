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

			initCmd := flux.installCmd(context.TODO())
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"install",
			}))
		})

		By("generating the command without network policy", func() {
			flux, err := NewFlux(".", "/mock/path/to/flux")
			Expect(err).To(BeNil())

			initCmd := flux.installCmd(context.TODO(), NetworkPolicy(false))
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"install",
				"--network-policy", "false",
			}))
		})
		By("generating the command to install only source controller", func() {
			flux, err := NewFlux(".", "/mock/path/to/flux")
			Expect(err).To(BeNil())

			initCmd := flux.installCmd(context.TODO(), Components(ComponentSourceController, ComponentHelmController))
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"install",
				"--components", "source-controller,helm-controller",
			}))
		})

		By("generating the command to install extra controllers", func() {
			flux, err := NewFlux(".", "/mock/path/to/flux")
			Expect(err).To(BeNil())

			initCmd := flux.installCmd(context.TODO(), ComponentsExtra(ComponentImageReflectorController, ComponentImageAutomationController))
			Expect(initCmd.Args[1:]).To(Equal([]string{
				"install",
				"--components-extra",
				"image-reflector-controller,image-automation-controller",
			}))
		})
	})
})
