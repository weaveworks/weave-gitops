package get

import (
	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/get/bcrypt"
)

func GetCommand(opts *config.Options) *cobra.Command {
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

	cmd.AddCommand(bcrypt.HashCommand(opts))

	return cmd
}
