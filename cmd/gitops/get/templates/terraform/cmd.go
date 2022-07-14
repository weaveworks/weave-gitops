package terraform

import (
	"errors"
	"os"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/printers"

	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/templates"
)

type templateCommandFlags struct {
	ListTemplateParameters bool
}

var flags templateCommandFlags

func TerraformCommand(opts *config.Options, client *adapters.HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "terraform",
		Aliases: []string{"terraform"},
		Short:   "Display one or many Terraform templates",
		Example: `
# Get all terraform templates
gitops get template terraform

# Show the parameters of a Terraform template
gitops get template terraform <template-name> --list-parameters
		`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE:       getTerraformTemplateCmdPreRunE(&opts.Endpoint),
		RunE:          getTerraformTemplateCmdRunE(opts, client),
		Args:          cobra.MaximumNArgs(1),
	}

	cmd.Flags().BoolVar(&flags.ListTemplateParameters, "list-parameters", false, "Show parameters of Terraform template")

	return cmd
}

func getTerraformTemplateCmdPreRunE(endpoint *string) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, args []string) error {
		if *endpoint == "" {
			return cmderrors.ErrNoWGEEndpoint
		}

		return nil
	}
}

func getTerraformTemplateCmdRunE(opts *config.Options, client *adapters.HTTPClient) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := client.ConfigureClientWithOptions(opts, os.Stdout)
		if err != nil {
			return err
		}

		w := printers.GetNewTabWriter(os.Stdout)
		defer w.Flush()

		if flags.ListTemplateParameters {
			if len(args) == 0 {
				return errors.New("terraform template name is required")
			}

			return templates.GetTemplateParameters(templates.GitOpsTemplateKind, args[0], client, w)
		}

		if len(args) == 0 {
			return templates.GetTemplates(templates.GitOpsTemplateKind, client, w)
		}

		return templates.GetTemplate(args[0], templates.GitOpsTemplateKind, client, w)
	}
}
