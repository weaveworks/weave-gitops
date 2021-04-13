package fluxops

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/version"
)

var fluxHandler = defaultFluxHandler

func SetFluxHandler(f func(string) ([]byte, error)) {
	fluxHandler = f
}

func CallFlux(arglist string) ([]byte, error) {
	return fluxHandler(arglist)
}

func defaultFluxHandler(arglist string) ([]byte, error) {
	return CallCommand(fmt.Sprintf("%s %s", fluxBin(), arglist))
}

func CallCommand(cmdstr string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", escape(cmdstr))
	return cmd.CombinedOutput()
}

func escape(cmd string) string {
	return strings.ReplaceAll(cmd, "'", "'\"'\"'")
}

func fluxBin() string {
	homeDir, err := os.UserHomeDir()
	checkError("failed to get user directory", err)

	return fmt.Sprintf("%s/.wego/bin/flux-%s", homeDir, version.FluxVersion)
}

// checkError will print a message to stderr and exit
func checkError(msg string, err interface{}) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		os.Exit(1)
	}
}
