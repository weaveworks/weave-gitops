package root

import (
	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

func getCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Display one or many Weave GitOps resources",
		Example: `
# Get the CLI configuration for Weave GitOps
gitops get config

# Generate a hashed secret
PASSWORD="<your password>"
echo -n $PASSWORD | gitops get bcrypt-hash`,
	}

	cmd.AddCommand(getBcryptHashCommand(opts))
	cmd.AddCommand(getConfigCommand(opts))

	return cmd
}
