package flux

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/version"
)

//go:embed bin/flux
var fluxExe []byte

func SetupFluxBin() {
	exePath, err := GetFluxExePath()
	checkError(err)
	binPath, err := GetFluxBinPath()
	checkError(err)

	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		// Clean bin if file doesnt exist
		checkError(os.RemoveAll(binPath))
		checkError(os.MkdirAll(binPath, 0755))
		checkError(os.WriteFile(exePath, fluxExe, 0755))
	}
}

func GetFluxBinPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%v/.wego/bin", homeDir), nil
}

func GetFluxExePath() (string, error) {
	path, err := GetFluxBinPath()
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v/flux-%v", path, version.FluxVersion), nil
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
