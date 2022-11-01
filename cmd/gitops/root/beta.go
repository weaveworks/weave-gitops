package root

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

func betaCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "beta",
		Short: "This component contains unstable or still-in-development functionality",
	}

	cmd.AddCommand(betaRunCommand(opts))

	return cmd
}
