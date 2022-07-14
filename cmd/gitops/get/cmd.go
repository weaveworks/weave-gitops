package get

import (
	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get/bcrypt"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get/clusters"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get/credentials"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get/profiles"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get/templates"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get/templates/terraform"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
)

func GetCommand(opts *config.Options, client *adapters.HTTPClient) *cobra.Command {
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

	templateCommand := templates.TemplateCommand(opts, client)
	terraformCommand := terraform.TerraformCommand(opts, client)
	templateCommand.AddCommand(terraformCommand)

	cmd.AddCommand(templateCommand)
	cmd.AddCommand(credentials.CredentialCommand(opts, client))
	cmd.AddCommand(clusters.ClusterCommand(opts, client))
	cmd.AddCommand(profiles.ProfilesCommand(opts, client))
	cmd.AddCommand(bcrypt.HashCommand(opts, client))

	return cmd
}
