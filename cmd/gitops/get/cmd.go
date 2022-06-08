package get

import (
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops/cmd/gitops/get/clusters"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get/credentials"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get/profiles"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get/templates"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get/templates/terraform"
)

func GetCommand(endpoint, username, password *string, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Display one or many Weave GitOps resources",
		Example: `
# Get all CAPI templates
gitops get templates

# Get all CAPI credentials
gitops get credentials

# Get all CAPI clusters
gitops get clusters`,
	}

	templateCommand := templates.TemplateCommand(endpoint, username, password, client)
	terraformCommand := terraform.TerraformCommand(endpoint, username, password, client)
	templateCommand.AddCommand(terraformCommand)

	cmd.AddCommand(templateCommand)
	cmd.AddCommand(credentials.CredentialCommand(endpoint, username, password, client))
	cmd.AddCommand(clusters.ClusterCommand(endpoint, username, password, client))
	cmd.AddCommand(profiles.ProfilesCommand(endpoint, username, password, client))

	return cmd
}
