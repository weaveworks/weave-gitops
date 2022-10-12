package get

import (
	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get/bcrypt"
	configCmd "github.com/weaveworks/weave-gitops/cmd/gitops/get/config"
)

func GetCommand(opts *config.Options) *cobra.Command {
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

	cmd.AddCommand(bcrypt.HashCommand(opts))
	cmd.AddCommand(configCmd.ConfigCommand(opts))

	return cmd
}
