package fluxinstall

import (
	"context"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Install Flux CLI", func() {
	cacheDir, err := os.UserCacheDir()
	Expect(err).To(BeNil())
	gitopsCacheFluxDir := filepath.Join(cacheDir, ".gitops", "flux", "0.32.0")

	It("should be able to manage lifecycle of a Flux binary", func() {
		By("creating an installer with a version, and call install", func() {
			ctx := context.TODO()

			product := &Product{
				Version: "0.32.0",
				cli:     &MockProductHTTPClient{},
			}

			installer := NewInstaller()
			execPath, err := installer.Install(ctx, product)
			Expect(err).To(BeNil())
			Expect(execPath).To(Equal(filepath.Join(gitopsCacheFluxDir, "flux")))
		})

		By("ensure that there's a version installed", func() {
			ctx := context.TODO()

			product := &Product{
				Version: "0.32.0",
				cli:     &MockProductHTTPClient{},
			}

			installer := NewInstaller()
			execPath, err := installer.Ensure(ctx, product)
			Expect(err).To(BeNil())
			Expect(execPath).To(Equal(filepath.Join(gitopsCacheFluxDir, "flux")))

			err = installer.Remove(context.TODO())
			Expect(err).To(BeNil())
		})
	})
})
