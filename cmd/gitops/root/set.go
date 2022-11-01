package root

import (
	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
)

func setCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Sets one or many Weave GitOps CLI configs or resources",
		Example: `
# Enables analytics in the current user's CLI configuration for Weave GitOps
gitops set config analytics true`,
	}

	cmd.AddCommand(setConfigCommand(opts))

	return cmd
}
