package ui

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/wego/ui/run"
)

var Cmd = &cobra.Command{
	Use:   "ui",
	Short: "Manages Wego UI",
	Example: `
  # Run wego ui in your machine
  wego ui run
`,
	Args: cobra.MinimumNArgs(1),
}

func init() {
	Cmd.AddCommand(run.Cmd)
}
