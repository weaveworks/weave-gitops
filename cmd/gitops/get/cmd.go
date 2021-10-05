package get

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get/clusters"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get/templates"
)

func GetCommand(endpoint *string, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Display one or many Weave GitOps resources",
		Example: `
# Get all CAPI templates
gitops get templates`,
	}

	cmd.AddCommand(templates.TemplateCommand(endpoint, client))
	cmd.AddCommand(clusters.ClusterCommand(endpoint, client))

	return cmd
}
