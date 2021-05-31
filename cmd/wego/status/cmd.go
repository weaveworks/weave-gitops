package status

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/cmdimpl"
)

var Cmd = &cobra.Command{
	Use:     "status [subcommands]",
	Short:   "status of a resource",
	Args:    cobra.MinimumNArgs(1),
	Example: "wego status application podinfo",
}

var ApplicationCmd = &cobra.Command{
	Use:   "application [name]",
	Short: "status of an application resource",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		params.Namespace, _ = cmd.Parent().Parent().Flags().GetString("namespace")
		params.Name = args[0]
		return cmdimpl.Status(params)
	},
}

var params cmdimpl.AddParamSet

func init() {
	Cmd.AddCommand(ApplicationCmd)
}
