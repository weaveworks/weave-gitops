package credentials

import (
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/gitops/pkg/capi"
	"k8s.io/cli-runtime/pkg/printers"
)

func CredentialCommand(endpoint *string, client *resty.Client) *cobra.Command {
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
		PreRunE:       getCredentialCmdPreRunE(endpoint, client),
		RunE:          getCredentialCmdRunE(endpoint, client),
	}

	return cmd
}

func getCredentialCmdPreRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		if *endpoint == "" {
			return cmderrors.ErrNoWGEEndpoint
		}

		return nil
	}
}

func getCredentialCmdRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r, err := adapters.NewHttpClient(*endpoint, client, os.Stdout)
		if err != nil {
			return err
		}

		w := printers.GetNewTabWriter(os.Stdout)
		defer w.Flush()

		return capi.GetCredentials(r, w)
	}
}
