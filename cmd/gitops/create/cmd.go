package create

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/create/tenants"
)

func CreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "create new resources",
		Example: `
# Create a new tenant
gitops create tenants --from-file tenants.yaml`,
	}

	cmd.AddCommand(tenants.TenantsCommand)

	return cmd
}
