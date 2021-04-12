package fluxops

import (
	"fmt"
	"os/exec"
	"strings"
)

var fluxHandler = defaultFluxHandler

func SetFluxHandler(f func(string) ([]byte, error)) {
	fluxHandler = f
}

func CallFlux(arglist string) ([]byte, error) {
	return fluxHandler(arglist)
}

func defaultFluxHandler(arglist string) ([]byte, error) {
	return CallCommand(fmt.Sprintf("wego flux %s", arglist))
}

func CallCommand(cmdstr string) ([]byte, error) {
	cmd := exec.Command("sh", "-c", escape(cmdstr))
	return cmd.CombinedOutput()
}

func escape(cmd string) string {
	return strings.ReplaceAll(cmd, "'", "'\"'\"'")
}
