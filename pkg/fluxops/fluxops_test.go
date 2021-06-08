package fluxops_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/weaveworks/weave-gitops/pkg/fluxops"
	"github.com/weaveworks/weave-gitops/pkg/fluxops/fluxopsfakes"
)

const hubCreds = `github.com:
- user: pandagool
  oauth_token: 36e4f5f7b4f7b626069d3503a5b6a22a54fcd127
  protocol: https
`

func TestFluxOps(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Fluxops Tests")
}

var _ = Describe("User Lookup Test", func() {
	It("Verify that looking up a user's credentials via hub works correctly", func() {
		By("setting up a fake HOME directory and pulling hub credentials", func() {
			os.Unsetenv("GITHUB_ORG")
			dir, err := ioutil.TempDir("", "tmp-dir")
			Expect(err).To(BeNil())
			err = os.Setenv("HOME", dir)
			Expect(err).To(BeNil())
			err = os.MkdirAll(filepath.Join(dir, ".config"), 0755)
			Expect(err).To(BeNil())
			err = ioutil.WriteFile(filepath.Join(dir, ".config", "hub"), []byte(hubCreds), 0600)
			Expect(err).To(BeNil())
			creds, err := fluxops.GetOwnerFromEnv()
			Expect(err).To(BeNil())
			Expect(creds).To(Equal("pandagool"))
		})
	})
})

var _ = Describe("Flux Install Test", func() {
	It("Check all the install paths", func() {
		By("Using a mock to mimic an install", func() {
			fakeHandler := &fluxopsfakes.FakeFluxHandler{
				HandleStub: func(args string) ([]byte, error) {
					return []byte("foo"), nil
				},
			}
			fluxops.SetFluxHandler(fakeHandler)
			output, err := fluxops.Install("flux-system", false)
			Expect(err).To(BeNil())
			Expect(string(output)).To(Equal("foo"))

			_, err = fluxops.Install("my-namespace", false)
			Expect(err).To(BeNil())
			args := fakeHandler.HandleArgsForCall(1)
			Expect(args).To(Equal("install --namespace=my-namespace --components-extra=image-reflector-controller,image-automation-controller"))
		})

		By("Using a mock to fail verbose manifest generation", func() {
			fakeHandler := &fluxopsfakes.FakeFluxHandler{
				HandleStub: func(args string) ([]byte, error) {
					return nil, fmt.Errorf("failed")
				},
			}
			fluxops.SetFluxHandler(fakeHandler)
			_, err := fluxops.Install("flux-system", false)
			Expect(err.Error()).To(Equal("failed"))
		})

		By("Using a mock to fail quiet manifest generation", func() {
			fakeHandler := &fluxopsfakes.FakeFluxHandler{
				HandleStub: func(args string) ([]byte, error) {
					return nil, fmt.Errorf("failed")
				},
			}
			fluxops.SetFluxHandler(fakeHandler)
			_, err := fluxops.QuietInstall("flux-system")
			Expect(err.Error()).To(Equal("failed"))
		})
	})
})
