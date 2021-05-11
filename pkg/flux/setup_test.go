package flux

import (
	"fmt"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops/pkg/shims"
	"github.com/weaveworks/weave-gitops/pkg/version"
)

type localExitHandler struct {
	action func(int)
}

func (h localExitHandler) Handle(code int) {
	h.action(code)
}

func TestSetup(t *testing.T) {
	_, err := GetFluxBinPath()
	require.NoError(t, err)

	_, err = GetFluxExePath()
	require.NoError(t, err)
}

func TestSetupFluxBin(t *testing.T) {
	version.FluxVersion = "0.11.0"
	SetupFluxBin()
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	fluxPath := fmt.Sprintf("%v/.wego/bin", homeDir)
	require.DirExists(t, fluxPath)
	binPath := fmt.Sprintf("%v/flux-%v", fluxPath, version.FluxVersion)
	require.FileExists(t, binPath)

	version.FluxVersion = "0.12.0"
	SetupFluxBin()
	require.NoFileExists(t, binPath)
	binPath = fmt.Sprintf("%v/flux-%v", fluxPath, version.FluxVersion)
	require.FileExists(t, binPath)
}

var _ = Describe("Flux Setup Failure", func() {
	It("Verify that exit is called with expected code", func() {
		By("Executing a code path that contains checkError", func() {
			exitCode := -1
			shims.WithExitHandler(localExitHandler{action: func(code int) { exitCode = code }},
				func() {
					checkError(fmt.Errorf("An error"))
				})
			Expect(exitCode).To(Equal(1))
		})
	})
})
