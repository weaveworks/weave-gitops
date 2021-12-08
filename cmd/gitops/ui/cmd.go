package ui

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/ui/run"
)

// Command returns the `ui` command and its subcommands.
func Command() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ui",
		Short: "Manages Gitops UI",
		Example: `
	  # Run gitops ui in your machine
	  gitops ui run
	`,
		Args: cobra.MinimumNArgs(1),
	}

	cmd.AddCommand(run.Command())

	return cmd
}
