package flux

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/flux"
)

//go:embed bin/flux
var fluxExe []byte

var Cmd = &cobra.Command{
	Use:   "flux",
	Short: "Use flux commands",
	Run:   runCmd,
}

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the last known status of flux namespaces",
	Run:   runStatusCmd,
}

func init() {
	Cmd.AddCommand(StatusCmd)
	checkFluxBinSetup()
}

// Example flux command with flags 'wego flux -- install -h'
func runCmd(cmd *cobra.Command, args []string) {
	exePath, err := flux.GetFluxExePath()
	checkError(err)

	c := exec.Command(exePath, args...)

	// run command
	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	} else {
		fmt.Printf("Output: %s\n", output)
	}
}

func runStatusCmd(cmd *cobra.Command, args []string) {
	status, err := flux.GetLatestStatusAllNamespaces()
	if err != nil {
		checkError(err)
	}
	fmt.Printf("Status: %s\n", status)
}

func checkFluxBinSetup() {
	exePath, err := flux.GetFluxExePath()
	checkError(err)
	binPath, err := flux.GetFluxBinPath()
	checkError(err)

	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		// Clean bin if file doesnt exist
		checkError(os.RemoveAll(binPath))
		checkError(os.MkdirAll(binPath, 0755))
		checkError(os.WriteFile(exePath, fluxExe, 0755))
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
