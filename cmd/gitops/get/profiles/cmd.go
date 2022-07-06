package profiles

import (
	"context"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/services/profiles"
	"k8s.io/cli-runtime/pkg/printers"
)

func ProfilesCommand(opts *config.Options, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "profile",
		Aliases:       []string{"profiles"},
		Short:         "Show information about available profiles",
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: `
	# Get all profiles
	gitops get profiles
	`,
		PreRunE: getProfilesCmdPreRunE(&opts.Endpoint),
		RunE:    getProfilesCmdRunE(opts, client),
	}

	return cmd
}

func getProfilesCmdPreRunE(endpoint *string) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		if *endpoint == "" {
			return cmderrors.ErrNoWGEEndpoint
		}

		return nil
	}
}

func getProfilesCmdRunE(opts *config.Options, client *resty.Client) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		r, err := adapters.NewHttpClient(opts.Endpoint, opts.Username, opts.Password, client, os.Stdout)
		if err != nil {
			return err
		}

		w := printers.GetNewTabWriter(os.Stdout)

		defer w.Flush()

		return profiles.NewService(internal.NewCLILogger(os.Stdout)).Get(context.Background(), r, w)
	}
}
