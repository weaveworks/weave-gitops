package terraform

import (
	"errors"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/printers"

	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/templates"
)

type templateCommandFlags struct {
	ListTemplateParameters bool
}

var flags templateCommandFlags

func TerraformCommand(endpoint, username, password *string, client *resty.Client) *cobra.Command {
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
		PreRunE:       getTerraformTemplateCmdPreRunE(endpoint, client),
		RunE:          getTerraformTemplateCmdRunE(endpoint, username, password, client),
		Args:          cobra.MaximumNArgs(1),
	}

	cmd.Flags().BoolVar(&flags.ListTemplateParameters, "list-parameters", false, "Show parameters of Terraform template")

	return cmd
}

func getTerraformTemplateCmdPreRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, args []string) error {
		if *endpoint == "" {
			return cmderrors.ErrNoWGEEndpoint
		}

		return nil
	}
}

func getTerraformTemplateCmdRunE(endpoint, username, password *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r, err := adapters.NewHttpClient(*endpoint, *username, *password, client, os.Stdout)
		if err != nil {
			return err
		}

		w := printers.GetNewTabWriter(os.Stdout)
		defer w.Flush()

		if flags.ListTemplateParameters {
			if len(args) == 0 {
				return errors.New("terraform template name is required")
			}

			return templates.GetTemplateParameters(templates.GitOpsTemplateKind, args[0], r, w)
		}

		if len(args) == 0 {
			return templates.GetTemplates(templates.GitOpsTemplateKind, r, w)
		}

		return nil
	}
}
