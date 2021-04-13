package fluxops

import (
	"bytes"
	"fmt"
	"os"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/utils"
	"github.com/weaveworks/weave-gitops/pkg/version"
)

var (
	fluxHandler FluxHandler = defaultFluxHandler{}
	fluxBinary  string
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . FluxHandler
type FluxHandler interface {
	Handle(args string) ([]byte, error)
}

type defaultFluxHandler struct{}

func (h defaultFluxHandler) Handle(arglist string) ([]byte, error) {
	initFluxBinary()
	return utils.CallCommand(fmt.Sprintf("%s %s", fluxBinary, arglist))
}

func FluxPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("%v/.wego/bin", homeDir)
	return fmt.Sprintf("%v/flux-%v", path, version.FluxVersion), nil
}

func SetFluxHandler(h FluxHandler) {
	fluxHandler = h
}

func CallFlux(arglist []string) ([]byte, error) {
	return fluxHandler.Handle(strings.Join(arglist, " "))
}

func Install(namespace string) ([]byte, error) {
	args := []string{
		"install",
		fmt.Sprintf("--namespace=%s", namespace),
		"--export",
	}

	return CallFlux(args)
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
