package clusters

import (
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/templates"
)

type clusterCommandFlags struct {
	DryRun          bool
	Template        string
	ParameterValues []string
}

var flags clusterCommandFlags

func ClusterCommand(endpoint *string, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Add a new cluster using a CAPI template",
		Example: `
# Add a new cluster using a CAPI template
gitops add cluster --from-template <template-name>

# View a CAPI template populated with parameter values 
# without creating a pull request for it
gitops add cluster --from-template <template-name> --set key=val --dry-run
		`,
		RunE: getClusterCmdRunE(endpoint, client),
	}

	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "View the populated template without creating a pull request")
	cmd.Flags().StringVar(&flags.Template, "from-template", "", "Specify the CAPI template to create a cluster from")
	cmd.Flags().StringSliceVar(&flags.ParameterValues, "set", []string{}, "Set parameter values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")

	return cmd
}

func getClusterCmdRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r, err := adapters.NewHttpClient(*endpoint, client, os.Stdout)
		if err != nil {
			return err
		}

		vals := make(map[string]string)

		for _, v := range flags.ParameterValues {
			kv := strings.SplitN(v, "=", 2)
			if len(kv) == 2 {
				vals[kv[0]] = kv[1]
			}
		}

		if flags.DryRun {
			creds := templates.Credentials{}
			return templates.RenderTemplate(flags.Template, vals, creds, r, os.Stdout)
		}

		return nil
	}
}
