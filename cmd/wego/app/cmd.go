package app

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/add"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/list"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/status"
)

var ApplicationCmd = &cobra.Command{
	Use:   "app [subcommand]",
	Short: "Manages your applications",
	Long:  `
  Add:    Add an application to wego control,
  Status: Get status of application under wego control`,
	Example:`
  # Add an application to wego from local git repository
  wego app add ./myapp --name myapp

  # Status an application under wego control
  wego app status myapp`,
	Args: cobra.MinimumNArgs(1),
}

func init() {
	ApplicationCmd.AddCommand(status.Cmd)
	ApplicationCmd.AddCommand(add.Cmd)
	ApplicationCmd.AddCommand(list.Cmd)
}
