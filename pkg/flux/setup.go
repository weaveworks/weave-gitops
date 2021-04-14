package flux

import (
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/version"
)

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
