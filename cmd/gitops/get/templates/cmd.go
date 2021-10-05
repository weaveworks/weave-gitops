package templates

import (
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/templates"
	"k8s.io/cli-runtime/pkg/printers"
)

type templateCommandFlags struct {
	ListTemplateParameters bool
}

var flags templateCommandFlags

func TemplateCommand(endpoint *string, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "template",
		Aliases: []string{"templates"},
		Short:   "Display one or many CAPI templates",
		Example: `
# Get all CAPI templates
gitops get templates

# Show the parameters of a CAPI template
gitops get template <template-name> --list-parameters
		`,
		RunE: getTemplateCmdRunE(endpoint, client),
		Args: cobra.MaximumNArgs(1),
	}

	cmd.Flags().BoolVar(&flags.ListTemplateParameters, "list-parameters", false, "Show parameters of CAPI template")

	return cmd
}

func getTemplateCmdRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r, err := adapters.NewHttpClient(*endpoint, client, os.Stdout)
		if err != nil {
			return err
		}

		w := printers.GetNewTabWriter(os.Stdout)
		defer w.Flush()

		if flags.ListTemplateParameters {
			return templates.GetTemplateParameters(args[0], r, w)
		}

		if len(args) == 0 {
			return templates.GetTemplates(r, w)
		}

		return nil
	}
}
