package clusters

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

func ClusterCommand(endpoint *string, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Add a new cluster using a CAPI template",
		Example: `
# Add a new cluster using a CAPI template
gitops add cluster <template-name>
		`,
		RunE: getClusterCmdRunE(endpoint, client),
		Args: cobra.ExactArgs(1),
	}

	return cmd
}

func getClusterCmdRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return nil
	}
}
