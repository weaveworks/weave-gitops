package status

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/cmdimpl"
)

var Cmd = &cobra.Command{
	Use:     "status <app-name>",
	Short:   "Get status of a workload under wego control",
	Args:    cobra.MinimumNArgs(1),
	Example: "wego app status podinfo",
	RunE: func(cmd *cobra.Command, args []string) error {
		params.Namespace, _ = cmd.Parent().Parent().Flags().GetString("namespace")
		params.Name = args[0]
		return cmdimpl.Status(params)
	},
}

var params cmdimpl.StatusParams
