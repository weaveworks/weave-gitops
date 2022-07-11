package beta

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/beta/run"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
)

func GetCommand(opts *config.Options, client *adapters.HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "beta",
		Short: "This component contains unstable or still-in-development functionality",
	}

	cmd.AddCommand(run.RunCommand(opts, client))

	return cmd
}
