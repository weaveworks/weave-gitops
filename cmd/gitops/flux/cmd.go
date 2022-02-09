package flux

import (
	"fmt"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

var Cmd = &cobra.Command{
	Use:                "flux [flux commands or flags]",
	Short:              "Use flux commands",
	DisableFlagParsing: true,
	Example:            "gitops flux install -h",
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

// Example flux command with flags 'gitops flux -- install -h'
func runCmd(cmd *cobra.Command, args []string) {
	cliRunner := &runner.CLIRunner{}
	fluxClient := flux.New(osys.New(), cliRunner)

	exePath, err := fluxClient.GetExePath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	c := exec.Command(exePath, args...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	// run command
	_ = c.Run()
}

func runStatusCmd(cmd *cobra.Command, args []string) {
	cliRunner := &runner.CLIRunner{}
	fluxClient := flux.New(osys.New(), cliRunner)

	status, err := fluxClient.GetLatestStatusAllNamespaces()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Status: %s\n", status)
}
