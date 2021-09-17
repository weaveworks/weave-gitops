package ui

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/ui/run"
)

var Cmd = &cobra.Command{
	Use:   "ui",
	Short: "Manages Gitops UI",
	Example: `
  # Run gitops ui in your machine
  gitops ui run
`,
	Args: cobra.MinimumNArgs(1),
}

func init() {
	Cmd.AddCommand(run.Cmd)
}
