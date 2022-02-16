package flux

import (
	"embed"
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/version"
)

//go:embed bin/*
var binFS embed.FS

//SetupFluxBin creates flux binary from embedded file if it doesnt already exist
func (f *FluxClient) SetupBin() {
	exePath, err := f.getExePath()
	f.checkError(err)
	binPath, err := f.getBinPath()
	f.checkError(err)

	var fluxBinary []byte

	fluxBinaryOverride := f.osys.Getenv(fluxBinaryPathEnvVar)
	if fluxBinaryOverride == "" {
		// Try and read embedded binary, this won't work if weave-gitops
		// is being used as a go module dependency by another module
		// and the fluxBinaryPathEnvVar env var must be set
		fluxBinary, err = binFS.ReadFile("bin/flux")
		if err != nil {
			f.checkError(fmt.Errorf(`error reading embedded flux binary: %v `, err))
		}
	} else {
		bin, err := os.ReadFile(fluxBinaryOverride)
		f.checkError(err)
		fluxBinary = bin
	}

	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		// Clean bin if file doesnt exist
		f.checkError(os.RemoveAll(binPath))
		f.checkError(os.MkdirAll(binPath, 0755))
		f.checkError(os.WriteFile(exePath, fluxBinary, 0755))
	}
}

//GetFluxBinPath -
func (f *FluxClient) getBinPath() (string, error) {
	homeDir, err := f.osys.UserHomeDir()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v/.wego/bin", homeDir), nil
}

//GetFluxExePath -
func (f *FluxClient) getExePath() (string, error) {
	path, err := f.getBinPath()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v/flux-%v", path, version.FluxVersion), nil
}

func (f *FluxClient) checkError(err error) {
	if err != nil {
		fmt.Fprintf(f.osys.Stderr(), "Error: %v\n", err)
		f.osys.Exit(1)
	}
}
