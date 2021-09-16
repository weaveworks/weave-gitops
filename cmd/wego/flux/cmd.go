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

func init() {
	Cmd.AddCommand(StatusCmd)
}

// Example flux command with flags 'wego flux -- install -h'
func runCmd(cmd *cobra.Command, args []string) {
	cliRunner := &runner.CLIRunner{}
	osysClient := osys.New()
	fluxClient := flux.New(osysClient, cliRunner)

	exePath, err := fluxClient.GetExePath()
	if err != nil {
		fmt.Fprintf(osysClient.Stderr(), "Error: %v\n", err)
		osysClient.Exit(1)
	}

	c := exec.Command(exePath, args...)
	c.Stdin = osysClient.Stdin()
	c.Stdout = osysClient.Stdout()
	c.Stderr = osysClient.Stderr()
	// run command
	_ = c.Run()
}

func runStatusCmd(cmd *cobra.Command, args []string) {
	cliRunner := &runner.CLIRunner{}
	osysClient := osys.New()
	fluxClient := flux.New(osysClient, cliRunner)

	status, err := fluxClient.GetLatestStatusAllNamespaces()
	if err != nil {
		fmt.Fprintf(osysClient.Stderr(), "Error: %v\n", err)
		osysClient.Exit(1)
	}

	fmt.Printf("Status: %s\n", status)
}
