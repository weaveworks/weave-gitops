package create

import (
	"fmt"
	"regexp"
	"time"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/create/dashboard"
)

type CreateCommandFlags struct {
	Export  string
	Timeout time.Duration
}

var flags CreateCommandFlags

func GetCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "create",
		Short:   "Creates a resource",
		PreRunE: createCommandPreRunE(&opts.Endpoint),
		Example: `
# Create a HelmRepository and HelmRelease to deploy Weave GitOps
gitops create dashboard ww-gitops \
  --password=$PASSWORD \
  --export > ./clusters/my-cluster/weave-gitops-dashboard.yaml
		`,
	}

	cmd.PersistentFlags().StringVar(&flags.Export, "export", "", "The path to export manifests to.")
	cmd.PersistentFlags().DurationVar(&flags.Timeout, "timeout", 30*time.Second, "The timeout for operations during resource creation.")

	cmd.AddCommand(dashboard.DashboardCommand(opts))

	return cmd
}

func createCommandPreRunE(endpoint *string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		numArgs := len(args)

		if numArgs == 0 {
			return cmderrors.ErrNoName
		}

		if numArgs > 1 {
			return cmderrors.ErrMultipleNames
		}

		name := args[0]
		if !validateObjectName(name) {
			return fmt.Errorf("name '%s' is invalid, it should adhere to standard defined in RFC 1123, the name can only contain alphanumeric characters or '-'", name)
		}

		return nil
	}
}

func validateObjectName(name string) bool {
	r := regexp.MustCompile(`^[a-z0-9]([a-z0-9\\-]){0,61}[a-z0-9]$`)
	return r.MatchString(name)
}
