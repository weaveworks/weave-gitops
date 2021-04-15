package flux

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops/pkg/version"
)

func TestCheckFluxBinSetup(t *testing.T) {
	version.FluxVersion = "0.11.0"
	checkFluxBinSetup()
	homeDir, err := os.UserHomeDir()
	require.NoError(t, err)

	fluxPath := fmt.Sprintf("%v/.wego/bin", homeDir)
	require.DirExists(t, fluxPath)
	binPath := fmt.Sprintf("%v/flux-%v", fluxPath, version.FluxVersion)
	require.FileExists(t, binPath)

	version.FluxVersion = "0.12.0"
	checkFluxBinSetup()
	require.NoFileExists(t, binPath)
	binPath = fmt.Sprintf("%v/flux-%v", fluxPath, version.FluxVersion)
	require.FileExists(t, binPath)
}
