package clusters

import (
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/clusters"
	"k8s.io/cli-runtime/pkg/printers"
)

type clustersGetFlags struct {
	Kubeconfig bool
}

var clustersGetCmdFlags clustersGetFlags

func ClusterCommand(endpoint *string, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"clusters"},
		Short:   "Display one or many CAPI clusters",
		Example: `
# Get all CAPI clusters
gitops get clusters
		`,
		RunE: getClustersCmdRunE(endpoint, client),
	}

	cmd.PersistentFlags().BoolVar(&clustersGetCmdFlags.Kubeconfig, "kubeconfig", false, "Returns the Kubeconfig of the workload cluster")

	return cmd
}

func getClustersCmdRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r, err := adapters.NewHttpClient(*endpoint, client, os.Stdout)
		if err != nil {
			return err
		}

		w := printers.GetNewTabWriter(os.Stdout)

		defer w.Flush()

		if len(args) == 1 {
			if clustersGetCmdFlags.Kubeconfig {
				return clusters.GetClusterKubeconfig(args[0], r, os.Stdout)
			}

			return clusters.GetClusterByName(args[0], r, w)
		}

		return clusters.GetClusters(r, w)
	}
}
