package root

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

func removeCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove various components of Weave GitOps",
	}

	cmd.AddCommand(removeRunCommand(opts))

	return cmd
}
