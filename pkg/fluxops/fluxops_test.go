package fluxops_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/assert"
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
	RunSpecs(t, "Log Tests")
}

var _ = Describe("User Lookup Test", func() {
	It("Verify that looking up a user's credentials via hub works correctly", func() {
		By("setting up a fake HOME directory and pulling hub credentials", func() {
			dir, err := ioutil.TempDir("", "tmp-dir")
			Expect(err).To(BeNil())
			err = os.Setenv("HOME", dir)
			Expect(err).To(BeNil())
			err = os.MkdirAll(filepath.Join(dir, ".config"), 0755)
			Expect(err).To(BeNil())
			err = ioutil.WriteFile(filepath.Join(dir, ".config", "hub"), []byte(hubCreds), 0600)
			Expect(err).To(BeNil())
			creds, err := fluxops.GetUserFromHubCredentials()
			Expect(err).To(BeNil())
			Expect(creds).To(Equal("pandagool"))
		})
	})
})

func TestFluxInstall(t *testing.T) {
	assert := assert.New(t)

	fakeHandler := &fluxopsfakes.FakeFluxHandler{
		HandleStub: func(args string) ([]byte, error) {
			return []byte("foo"), nil
		},
	}
	fluxops.SetFluxHandler(fakeHandler)
	output, err := fluxops.Install("flux-system")
	assert.Equal("foo", string(output))
	assert.NoError(err)

	output, err = fluxops.Install("my-namespace")
	assert.Equal("apiVersion: v1\nkind: Namespace\nmetadata:\n  name: flux-system\n---\nfoo", string(output))
	assert.NoError(err)

	args := fakeHandler.HandleArgsForCall(1)
	assert.Equal("install --namespace=my-namespace --export", args)
}
