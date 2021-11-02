package clusters

import (
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/clusters"
)

type clustersDeleteFlags struct {
	RepositoryURL string
	BaseBranch    string
	HeadBranch    string
	Title         string
	Description   string
	ClustersNames string
	CommitMessage string
}

var clustersDeleteCmdFlags clustersDeleteFlags

func ClusterCommand(endpoint *string, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"clusters"},
		Short:   "Delete a cluster given its name",
		Example: `
# Delete a CAPI cluster by its name
gitops delete cluster <cluster-name>
		`,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE:          getClusterCmdRunE(endpoint, client),
		Args:          cobra.MinimumNArgs(1),
	}

	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.RepositoryURL, "url", "", "The repository to open a pull request against")
	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.BaseBranch, "base", "", "The base branch to open the pull request against")
	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.HeadBranch, "branch", "", "The branch to create the pull request from")
	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.Title, "title", "", "The title of the pull request")
	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.Description, "description", "", "The description of the pull request")
	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.CommitMessage, "commit-message", "", "The commit message to use when deleting the clusters")

	return cmd
}

func getClusterCmdRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r, err := adapters.NewHttpClient(*endpoint, client, os.Stdout)
		if err != nil {
			return err
		}

		return clusters.DeleteClusters(clusters.DeleteClustersParams{
			RepositoryURL: clustersDeleteCmdFlags.RepositoryURL,
			HeadBranch:    clustersDeleteCmdFlags.HeadBranch,
			BaseBranch:    clustersDeleteCmdFlags.BaseBranch,
			Title:         clustersDeleteCmdFlags.Title,
			Description:   clustersDeleteCmdFlags.Description,
			ClustersNames: args,
			CommitMessage: clustersDeleteCmdFlags.CommitMessage,
		}, r, os.Stdout)
	}
}
