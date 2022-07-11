package clusters

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
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

func ClusterCommand(opts *config.Options, client *adapters.HTTPClient) *cobra.Command {
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
		PreRunE:       getClusterCmdPreRunE(&opts.Endpoint),
		RunE:          getClusterCmdRunE(opts, client),
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

func getClusterCmdPreRunE(endpoint *string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if *endpoint == "" {
			return cmderrors.ErrNoWGEEndpoint
		}

		return nil
	}
}

func getClusterCmdRunE(opts *config.Options, client *adapters.HTTPClient) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := client.ConfigureClientWithOptions(opts, os.Stdout)
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
		}, client, os.Stdout)
	}
}
