package remove

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/remove/run"
)

func GetCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove various components of Weave GitOps",
	}

	cmd.AddCommand(run.RunCommand(opts))

	return cmd
}
