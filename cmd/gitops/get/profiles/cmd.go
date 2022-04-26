package profiles

import (
	"context"
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/services/profiles"
	"k8s.io/cli-runtime/pkg/printers"
)

var (
	port string
)

func ProfilesCommand(endpoint *string, client *resty.Client) *cobra.Command {
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
		PreRunE: getProfilesCmdPreRunE(endpoint, client),
		RunE:    getProfilesCmdRunE(endpoint, client),
	}

	cmd.Flags().StringVar(&port, "port", server.DefaultPort, "Port the profiles API is running on")

	return cmd
}

func getProfilesCmdPreRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		if *endpoint == "" {
			return cmderrors.ErrNoWGEEndpoint
		}

		return nil
	}
}

func getProfilesCmdRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		endpointWithPort := fmt.Sprintf("%s:%s", *endpoint, port)
		r, err := adapters.NewHttpClient(endpointWithPort, client, os.Stdout)
		if err != nil {
			return err
		}

		w := printers.GetNewTabWriter(os.Stdout)

		defer w.Flush()

		return profiles.NewService(internal.NewCLILogger(os.Stdout)).Get(context.Background(), r, w)
	}
}
