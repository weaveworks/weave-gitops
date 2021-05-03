package fluxops

import (
	"fmt"
	"os"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/utils"
	"github.com/weaveworks/weave-gitops/pkg/version"
)

var (
	fluxHandler = defaultFluxHandler
	fluxBinary  string
)

func FluxPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("%v/.wego/bin", homeDir)
	return fmt.Sprintf("%v/flux-%v", path, version.FluxVersion), nil
}

func FluxBinaryPath() string {
	return fluxBinary
}

func SetFluxHandler(f func(string) ([]byte, error)) {
	fluxHandler = f
}

func CallFlux(arglist []string) ([]byte, error) {
	return fluxHandler(strings.Join(arglist, " "))
}

func Install(namespace string) ([]byte, error) {
	args := []string{
		"install",
		fmt.Sprintf("--namespace=%s", namespace),
		"--export",
	}

	return CallFlux(args)
}

func defaultFluxHandler(arglist string) ([]byte, error) {
	initFluxBinary()
	return utils.CallCommand(fmt.Sprintf("%s %s", fluxBinary, arglist))
}

func initFluxBinary() {
	if fluxBinary == "" {
		fluxPath, err := FluxPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to retrieve wego executable path: %v", err)
			os.Exit(1)
		}
		fluxBinary = fluxPath
	}
}
