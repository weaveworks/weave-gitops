package app

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/app/status"
)

var ApplicationCmd = &cobra.Command{
	Use:   "app",
	Short: "Manages your applications",
	Example: `
  # Status an application under gitops control
  gitops app status <app-name>`,
}

func init() {
	ApplicationCmd.AddCommand(status.Cmd)
}
