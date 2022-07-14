package templates

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/printers"

	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/templates"
)

type templateCommandFlags struct {
	ListTemplateParameters bool
	ListTemplateProfiles   bool
	Provider               string
}

var flags templateCommandFlags

var providers = []string{
	"aws",
	"azure",
	"digitalocean",
	"docker",
	"openstack",
	"packet",
	"vsphere",
}

func TemplateCommand(opts *config.Options, client *adapters.HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "template",
		Aliases: []string{"templates"},
		Short:   "Display one or many CAPI templates",
		Example: `
# Get all CAPI templates
gitops get templates

# Get all AWS CAPI templates
gitops get templates --provider aws

# Show the parameters of a CAPI template
gitops get template <template-name> --list-parameters
		`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE:       getTemplateCmdPreRunE(&opts.Endpoint),
		RunE:          getTemplateCmdRunE(opts, client),
		Args:          cobra.MaximumNArgs(1),
	}

	cmd.Flags().BoolVar(&flags.ListTemplateParameters, "list-parameters", false, "Show parameters of CAPI template")
	cmd.Flags().BoolVar(&flags.ListTemplateProfiles, "list-profiles", false, "Show profiles of CAPI template")
	cmd.Flags().StringVar(&flags.Provider, "provider", "", fmt.Sprintf("Filter templates by provider. Supported providers: %s", strings.Join(providers, " ")))

	return cmd
}

func getTemplateCmdPreRunE(endpoint *string) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, args []string) error {
		if c.Flag("provider").Changed && !contains(providers, c.Flag("provider").Value.String()) {
			return fmt.Errorf("provider %q is not valid", c.Flag("provider").Value.String())
		}

		if *endpoint == "" {
			return cmderrors.ErrNoWGEEndpoint
		}

		return nil
	}
}

func getTemplateCmdRunE(opts *config.Options, client *adapters.HTTPClient) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := client.ConfigureClientWithOptions(opts, os.Stdout)
		if err != nil {
			return err
		}

		w := printers.GetNewTabWriter(os.Stdout)
		defer w.Flush()

		if flags.ListTemplateParameters {
			if len(args) == 0 {
				return errors.New("template name is required")
			}

			return templates.GetTemplateParameters(templates.CAPITemplateKind, args[0], client, w)
		}

		if flags.ListTemplateProfiles {
			if len(args) == 0 {
				return errors.New("template name is required")
			}

			return templates.GetTemplateProfiles(args[0], client, w)
		}

		if len(args) == 0 {
			if flags.Provider != "" {
				return templates.GetTemplatesByProvider(templates.CAPITemplateKind, flags.Provider, client, w)
			}

			return templates.GetTemplates(templates.CAPITemplateKind, client, w)
		}

		return templates.GetTemplate(args[0], templates.CAPITemplateKind, client, w)
	}
}

func contains(ss []string, str string) bool {
	for _, s := range ss {
		if s == str {
			return true
		}
	}

	return false
}
