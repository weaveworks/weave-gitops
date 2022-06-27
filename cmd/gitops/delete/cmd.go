package delete

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/delete/clusters"
)

func DeleteCommand(opts *config.Options, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete one or many Weave GitOps resources",
		Example: `
# Delete a CAPI cluster given its name
gitops delete cluster <cluster-name>`,
	}

	cmd.AddCommand(clusters.ClusterCommand(opts, client))

	return cmd
}
