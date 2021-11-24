package profiles

import (
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/capi"
	"k8s.io/cli-runtime/pkg/printers"
)

func TemplateCommand(endpoint *string, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "profile",
		Aliases: []string{"profiles"},
		Short:   "Display one or many CAPI profiles",
		Example: `
# Get all CAPI profiles
gitops get profiles
		`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE:       getProfileCmdPreRunE(endpoint, client),
		RunE:          getProfileCmdRunE(endpoint, client),
		Args:          cobra.MaximumNArgs(1),
	}

	return cmd
}

func getProfileCmdPreRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, args []string) error {
		if *endpoint == "" {
			return cmderrors.ErrNoWGEEndpoint
		}

		return nil
	}
}

func getProfileCmdRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r, err := adapters.NewHttpClient(*endpoint, client, os.Stdout)
		if err != nil {
			return err
		}

		w := printers.GetNewTabWriter(os.Stdout)
		defer w.Flush()

		if len(args) == 0 {
			return capi.GetProfiles(r, w)
		}

		return nil
	}
}
