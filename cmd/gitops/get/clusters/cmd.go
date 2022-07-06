package clusters

import (
	"errors"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/clusters"
	"k8s.io/cli-runtime/pkg/printers"
)

type clustersGetFlags struct {
	Kubeconfig bool
}

var clustersGetCmdFlags clustersGetFlags

func ClusterCommand(opts *config.Options, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"clusters"},
		Short:   "Display one or many CAPI clusters",
		Example: `
# Get all CAPI clusters
gitops get clusters

# Get a single CAPI cluster
gitops get cluster <cluster-name>

# Get the Kubeconfig of a cluster
gitops get cluster <cluster-name> --kubeconfig`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE:       getClusterCmdPreRunE(&opts.Endpoint),
		RunE:          getClusterCmdRunE(opts, client),
	}

	cmd.PersistentFlags().BoolVar(&clustersGetCmdFlags.Kubeconfig, "print-kubeconfig", false, "Returns the Kubeconfig of the workload cluster")

	return cmd
}

func getClusterCmdPreRunE(endpoint *string) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		if *endpoint == "" {
			return cmderrors.ErrNoWGEEndpoint
		}

		return nil
	}
}

func getClusterCmdRunE(opts *config.Options, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r, err := adapters.NewHttpClient(opts, client, os.Stdout)
		if err != nil {
			return err
		}

		w := printers.GetNewTabWriter(os.Stdout)

		defer w.Flush()

		if clustersGetCmdFlags.Kubeconfig {
			if len(args) == 0 {
				return errors.New("cluster name is required")
			}

			return clusters.GetClusterKubeconfig(args[0], r, os.Stdout)
		}

		if len(args) == 1 {
			return clusters.GetClusterByName(args[0], r, w)
		}

		return clusters.GetClusters(r, w)
	}
}
