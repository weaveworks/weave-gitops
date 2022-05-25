package templates

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/capi"
	"k8s.io/cli-runtime/pkg/printers"
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

func TemplateCommand(endpoint, username, password *string, client *resty.Client) *cobra.Command {
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
		PreRunE:       getTemplateCmdPreRunE(endpoint, client),
		RunE:          getTemplateCmdRunE(endpoint, username, password, client),
		Args:          cobra.MaximumNArgs(1),
	}

	cmd.Flags().BoolVar(&flags.ListTemplateParameters, "list-parameters", false, "Show parameters of CAPI template")
	cmd.Flags().BoolVar(&flags.ListTemplateProfiles, "list-profiles", false, "Show profiles of CAPI template")
	cmd.Flags().StringVar(&flags.Provider, "provider", "", fmt.Sprintf("Filter templates by provider. Supported providers: %s", strings.Join(providers, " ")))
	return cmd
}

func getTemplateCmdPreRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
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

func getTemplateCmdRunE(endpoint, username, password *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r, err := adapters.NewHttpClient(*endpoint, *username, *password, client, os.Stdout)
		if err != nil {
			return err
		}
		w := printers.GetNewTabWriter(os.Stdout)
		defer w.Flush()

		if flags.ListTemplateParameters {
			if len(args) == 0 {
				return errors.New("template name is required")
			}

			return capi.GetTemplateParameters(args[0], r, w)
		}

		if flags.ListTemplateProfiles {
			if len(args) == 0 {
				return errors.New("template name is required")
			}

			return capi.GetTemplateProfiles(args[0], r, w)
		}

		if len(args) == 0 {
			if flags.Provider != "" {
				return capi.GetTemplatesByProvider(flags.Provider, r, w)
			}

			return capi.GetTemplates(r, w)
		}

		return nil
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
