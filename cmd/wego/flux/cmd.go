package flux

import (
	_ "embed"
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/version"
)

//go:embed bin/flux
var fluxExe []byte

var Cmd = &cobra.Command{
	Use:   "flux",
	Short: "Use flux commands",
	Run:   runCmd,
}
var exePath string

func init() {
	homeDir, err := os.UserHomeDir()
	checkError(err)

	path := fmt.Sprintf("%v/.wego/bin", homeDir)
	exePath = fmt.Sprintf("%v/flux-%v", path, version.FluxVersion)
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		// Clean bin if file doesnt exist
		checkError(os.RemoveAll(path))
		checkError(os.MkdirAll(path, 0755))
		checkError(os.WriteFile(exePath, fluxExe, 0755))
	}
}

// Example flux command with flags 'wego flux -- install -h'
func runCmd(cmd *cobra.Command, args []string) {
	c := exec.Command(exePath, args...)

	// run command
	if output, err := c.CombinedOutput(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	} else {
		fmt.Printf("Output: %s\n", output)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
