package create

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/create/dashboard"
)

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

	cmd.AddCommand(dashboard.DashboardCommand(opts))

	return cmd
}
