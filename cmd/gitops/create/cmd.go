package create

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/create/dashboard"
	"github.com/weaveworks/weave-gitops/cmd/gitops/create/terraform"
)

type CreateCommandFlags struct {
	Export  bool
	Timeout time.Duration
}

var flags CreateCommandFlags

func GetCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Creates a resource",
		Example: `
# Create a HelmRepository and HelmRelease to deploy Weave GitOps
gitops create dashboard ww-gitops \
  --password=$PASSWORD \
  --export > ./clusters/my-cluster/weave-gitops-dashboard.yaml

# Create a Terraform object
gitops create terraform my-resource \
  -n my-namespace \
  --source GitRepository/my-project \
  --path ./terraform \
  --interval 1m \
  --export > ./clusters/my-cluster/infra/terraform-my-resource.yaml
		`,
	}

	cmd.PersistentFlags().BoolVar(&flags.Export, "export", false, "Export in YAML format to stdout.")
	cmd.PersistentFlags().DurationVar(&flags.Timeout, "timeout", 3*time.Minute, "The timeout for operations during resource creation.")

	cmd.AddCommand(dashboard.DashboardCommand(opts))
	cmd.AddCommand(terraform.Command(opts))

	return cmd
}
