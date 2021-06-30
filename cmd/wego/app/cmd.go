package app

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/add"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/list"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/status"
)

var ApplicationCmd = &cobra.Command{
	Use:  "app [subcommand]",
	Args: cobra.MinimumNArgs(1),
}

func init() {
	ApplicationCmd.AddCommand(status.Cmd)
	ApplicationCmd.AddCommand(add.Cmd)
	ApplicationCmd.AddCommand(list.Cmd)
}
