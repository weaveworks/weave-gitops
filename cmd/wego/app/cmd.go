package app

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/add"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/commit"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/list"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/pause"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/remove"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/status"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/unpause"
)

var ApplicationCmd = &cobra.Command{
	Use:   "app",
	Short: "Manages your applications",
	Example: `
  # Add an application to wego from local git repository
  wego app add . --name <app-name>

  # Remove an application from wego
  wego app remove <app-name>

  # Status an application under wego control
  wego app status <app-name>

  # List applications under wego control
  wego app list

  # Pause gitops automation
  wego app pause <app-name>

  # Unpause gitops automation
  wego app unpause <app-name>
  
  # Get last 10 commits for application repo
  wego app get-commits <app-name>`,
	Args: cobra.MinimumNArgs(1),
}

func init() {
	ApplicationCmd.AddCommand(add.Cmd)
	ApplicationCmd.AddCommand(remove.Cmd)
	ApplicationCmd.AddCommand(list.Cmd)
	ApplicationCmd.AddCommand(status.Cmd)
	ApplicationCmd.AddCommand(pause.Cmd)
	ApplicationCmd.AddCommand(unpause.Cmd)
	ApplicationCmd.AddCommand(commit.Cmd)
}
