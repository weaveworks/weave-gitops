package replan

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/replan/terraform"
)

func Command(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replan",
		Short: "Replan a resource",
		Example: `
# Replan the Terraform plan of a Terraform object from the "flux-system" namespace
gitops replan terraform --namespace flux-system my-resource
`,
	}

	cmd.AddCommand(terraform.Command(opts))

	return cmd
}
