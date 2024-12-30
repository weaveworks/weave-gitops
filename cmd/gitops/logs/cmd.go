package logs

import (
	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/logs/terraform"
)

func GetCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Get logs for a resource",
	}

	cmd.AddCommand(terraform.Command(opts))

	return cmd
}
