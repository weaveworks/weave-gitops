package clusters

import (
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/apputils"
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
		RunE: deleteClusterCmdRunE(endpoint, client),
		Args: cobra.MinimumNArgs(1),
	}

	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.RepositoryURL, "pr-repo", "", "The repository to open a pull request against")
	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.BaseBranch, "pr-base", "", "The base branch to open the pull request against")
	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.HeadBranch, "pr-branch", "", "The branch to create the pull request from")
	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.Title, "pr-title", "", "The title of the pull request")
	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.Description, "pr-description", "", "The description of the pull request")
	cmd.PersistentFlags().StringVar(&clustersDeleteCmdFlags.CommitMessage, "pr-commit-message", "", "The commit message to use when deleting the clusters")

	return cmd
}

func deleteClusterCmdRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r, err := adapters.NewHttpClient(*endpoint, client, os.Stdout)
		if err != nil {
			return err
		}

		token, err := apputils.GetTokenForRepositoryURL(clustersDeleteCmdFlags.RepositoryURL)
		if err != nil {
			return fmt.Errorf("failed to get token for git repository %q: %w", clustersDeleteCmdFlags.RepositoryURL, err)
		}

		return clusters.DeleteClusters(clusters.DeleteClustersParams{
			GitProviderToken: token,
			RepositoryURL:    clustersDeleteCmdFlags.RepositoryURL,
			HeadBranch:       clustersDeleteCmdFlags.HeadBranch,
			BaseBranch:       clustersDeleteCmdFlags.BaseBranch,
			Title:            clustersDeleteCmdFlags.Title,
			Description:      clustersDeleteCmdFlags.Description,
			ClustersNames:    args,
			CommitMessage:    clustersDeleteCmdFlags.CommitMessage,
		}, r, os.Stdout)
	}
}
