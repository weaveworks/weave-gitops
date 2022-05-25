package clusters

import (
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/clusters"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
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

var flags clustersDeleteFlags

func ClusterCommand(endpoint, username, password *string, client *resty.Client) *cobra.Command {
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
		PreRunE:       getClusterCmdPreRunE(endpoint, client),
		RunE:          getClusterCmdRunE(endpoint, username, password, client),
		Args:          cobra.MinimumNArgs(1),
	}

	cmd.Flags().StringVar(&flags.RepositoryURL, "url", "", "The repository to open a pull request against")
	cmd.Flags().StringVar(&flags.BaseBranch, "base", "", "The base branch to open the pull request against")
	cmd.Flags().StringVar(&flags.HeadBranch, "branch", "", "The branch to create the pull request from")
	cmd.Flags().StringVar(&flags.Title, "title", "", "The title of the pull request")
	cmd.Flags().StringVar(&flags.Description, "description", "", "The description of the pull request")
	cmd.Flags().StringVar(&flags.CommitMessage, "commit-message", "", "The commit message to use when deleting the clusters")

	return cmd
}

func getClusterCmdPreRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if *endpoint == "" {
			return cmderrors.ErrNoWGEEndpoint
		}

		return nil
	}
}

func getClusterCmdRunE(endpoint, username, password *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		r, err := adapters.NewHttpClient(*endpoint, *username, *password, client, os.Stdout)
		if err != nil {
			return err
		}

		if flags.RepositoryURL == "" {
			return cmderrors.ErrNoURL
		}

		url, err := gitproviders.NewRepoURL(flags.RepositoryURL)
		if err != nil {
			return fmt.Errorf("cannot parse url: %w", err)
		}

		token, err := internal.GetToken(url, os.LookupEnv)
		if err != nil {
			return err
		}

		return clusters.DeleteClusters(clusters.DeleteClustersParams{
			GitProviderToken: token,
			RepositoryURL:    flags.RepositoryURL,
			HeadBranch:       flags.HeadBranch,
			BaseBranch:       flags.BaseBranch,
			Title:            flags.Title,
			Description:      flags.Description,
			ClustersNames:    args,
			CommitMessage:    flags.CommitMessage,
		}, r, os.Stdout)
	}
}
