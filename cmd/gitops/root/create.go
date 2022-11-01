package root

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

var createCommandFlags struct {
	Export  bool
	Timeout time.Duration
}

func createCommand(opts *config.Options) *cobra.Command {
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

	cmd.PersistentFlags().BoolVar(&createCommandFlags.Export, "export", false, "Export in YAML format to stdout.")
	cmd.PersistentFlags().DurationVar(&createCommandFlags.Timeout, "timeout", 3*time.Minute, "The timeout for operations during resource creation.")

	cmd.AddCommand(createDashboardCommand(opts))

	return cmd
}
