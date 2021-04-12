package fluxops

import (
	"fmt"
	"os"
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
	homedir := os.Getenv("HOME")
	return CallCommand(fmt.Sprintf("%s/.wego/bin/flux %s", homedir, arglist))
}

func CallCommand(cmdstr string) ([]byte, error) {
	cmd := exec.Command(fmt.Sprintf("sh -c '%s'", escape(cmdstr)))
	return cmd.CombinedOutput()
}

func escape(cmd string) string {
	return "'" + strings.ReplaceAll(cmd, "'", "'\"'\"'") + "'"
}
