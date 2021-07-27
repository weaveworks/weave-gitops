package flux

import (
	"fmt"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

var Cmd = &cobra.Command{
	Use:                "flux [flux commands or flags]",
	Short:              "Use flux commands",
	DisableFlagParsing: true,
	Example:            "wego flux install -h",
	Run:                runCmd,
}

var StatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check the last known status of flux namespaces",
	Run:   runStatusCmd,
}

var fluxClient *flux.FluxClient

var osysClient *osys.OsysClient

func init() {
	Cmd.AddCommand(StatusCmd)
	cliRunner := &runner.CLIRunner{}
	osysClient = osys.New()
	fluxClient = flux.New(osysClient, cliRunner)
}

// Example flux command with flags 'wego flux -- install -h'
func runCmd(cmd *cobra.Command, args []string) {
	exePath, err := fluxClient.GetExePath()
	if err != nil {
		fmt.Fprintf(osysClient.Stderr(), "Error: %v\n", err)
		osysClient.Exit(1)
	}

	c := exec.Command(exePath, args...)

	// run command
	output, _ := c.CombinedOutput()
	fmt.Print(string(output))
}

func runStatusCmd(cmd *cobra.Command, args []string) {
	status, err := fluxClient.GetLatestStatusAllNamespaces()
	if err != nil {
		fmt.Fprintf(osysClient.Stderr(), "Error: %v\n", err)
		osysClient.Exit(1)
	}
	fmt.Printf("Status: %s\n", status)
}
