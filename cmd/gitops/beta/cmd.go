package beta

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/beta/run"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

func GetCommand(opts *config.Options, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "beta",
		Short: "This component contains unstable or still-in-development functionality",
	}

	cmd.AddCommand(run.RunCommand(opts, client))

	return cmd
}
