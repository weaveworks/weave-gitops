package beta

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/beta/run"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

func GetCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "beta",
		Short: "This component contains unstable or still-in-development functionality",
	}

	cmd.AddCommand(run.RunCommand(opts))

	return cmd
}
