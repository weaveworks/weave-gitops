package credentials

import (
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/templates"
	"k8s.io/cli-runtime/pkg/printers"
)

func CredentialCommand(opts *config.Options, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "credential",
		Aliases: []string{"credentials"},
		Short:   "Get CAPI credentials",
		Example: `
# Get all CAPI credentials
gitops get credentials
		`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE:       getCredentialCmdPreRunE(&opts.Endpoint),
		RunE:          getCredentialCmdRunE(opts, client),
	}

	return cmd
}

func getCredentialCmdPreRunE(endpoint *string) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		if *endpoint == "" {
			return cmderrors.ErrNoWGEEndpoint
		}

		return nil
	}
}

func getCredentialCmdRunE(opts *config.Options, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r, err := adapters.NewHttpClient(opts.Endpoint, opts.Username, opts.Password, client, os.Stdout)
		if err != nil {
			return err
		}

		w := printers.GetNewTabWriter(os.Stdout)

		defer w.Flush()

		return templates.GetCredentials(r, w)
	}
}
