package add

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/add/clusters"
	"github.com/weaveworks/weave-gitops/cmd/gitops/add/profiles"
	"github.com/weaveworks/weave-gitops/cmd/gitops/add/terraform"
)

func GetCommand(endpoint, username, password *string, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new Weave GitOps resource",
		Example: `
# Add a new cluster using a CAPI template
gitops add cluster`,
	}

	cmd.AddCommand(clusters.ClusterCommand(endpoint, username, password, client))
	cmd.AddCommand(profiles.AddCommand(endpoint, username, password, client))
	cmd.AddCommand(terraform.AddCommand(endpoint, username, password, client))

	return cmd
}
