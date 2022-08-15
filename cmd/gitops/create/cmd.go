package create

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/create/dashboard"
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
		`,
	}

	cmd.PersistentFlags().BoolVar(&flags.Export, "export", false, "Export in YAML format to stdout.")
	cmd.PersistentFlags().DurationVar(&flags.Timeout, "timeout", 30*time.Second, "The timeout for operations during resource creation.")

	cmd.AddCommand(dashboard.DashboardCommand(opts))

	return cmd
}
