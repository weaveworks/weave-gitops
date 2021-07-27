package flux

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/shims"
	"github.com/weaveworks/weave-gitops/pkg/version"
)

//go:embed bin/flux
var fluxExe []byte

//SetupFluxBin creates flux binary from embedded file if it doesnt already exist
func (f *FluxClient) SetupBin() {
	exePath, err := f.GetExePath()
	checkError(err)
	binPath, err := f.GetBinPath()
	checkError(err)

	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		// Clean bin if file doesnt exist
		checkError(os.RemoveAll(binPath))
		checkError(os.MkdirAll(binPath, 0755))
		checkError(os.WriteFile(exePath, fluxExe, 0755))
	}
}

//GetFluxBinPath -
func (f *FluxClient) GetBinPath() (string, error) {
	homeDir, err := f.osys.UserHomeDir()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v/.wego/bin", homeDir), nil
}

//GetFluxExePath -
func (f *FluxClient) GetExePath() (string, error) {
	path, err := f.GetBinPath()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v/flux-%v", path, version.FluxVersion), nil
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		shims.Exit(1)
	}
}
