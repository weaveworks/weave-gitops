package clusters

import (
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/capi"
	"github.com/weaveworks/weave-gitops/pkg/wegoerrors"
)

type clusterCommandFlags struct {
	DryRun          bool
	Template        string
	ParameterValues []string
	RepositoryURL   string
	BaseBranch      string
	HeadBranch      string
	Title           string
	Description     string
	CommitMessage   string
	Credentials     string
}

var flags clusterCommandFlags

func ClusterCommand(endpoint *string, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cluster",
		Short: "Add a new cluster using a CAPI template",
		Example: `
# Add a new cluster using a CAPI template
gitops add cluster --from-template <template-name> --set key=val

# View a CAPI template populated with parameter values 
# without creating a pull request for it
gitops add cluster --from-template <template-name> --set key=val --dry-run
		`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE:       getClusterCmdPreRunE(endpoint, client),
		RunE:          getClusterCmdRunE(endpoint, client),
	}

	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "View the populated template without creating a pull request")
	cmd.Flags().StringVar(&flags.Template, "from-template", "", "Specify the CAPI template to create a cluster from")
	cmd.Flags().StringSliceVar(&flags.ParameterValues, "set", []string{}, "Set parameter values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	cmd.Flags().StringVar(&flags.RepositoryURL, "url", "", "URL of remote repository to create the pull request")
	cmd.Flags().StringVar(&flags.BaseBranch, "base", "", "The base branch of the remote repository")
	cmd.Flags().StringVar(&flags.HeadBranch, "branch", "", "The branch to create the pull request from")
	cmd.Flags().StringVar(&flags.Title, "title", "", "The title of the pull request")
	cmd.Flags().StringVar(&flags.Description, "description", "", "The description of the pull request")
	cmd.Flags().StringVar(&flags.CommitMessage, "commit-message", "", "The commit message to use when adding the CAPI template")
	cmd.Flags().StringVar(&flags.Credentials, "set-credentials", "", "The CAPI credentials to use")

	return cmd
}

func getClusterCmdPreRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		if *endpoint == "" {
			return wegoerrors.ErrWGEHTTPApiEndpointNotSet
		}

		return nil
	}
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

		creds := capi.Credentials{}
		if flags.Credentials != "" {
			creds, err = r.RetrieveCredentialsByName(flags.Credentials)
			if err != nil {
				return err
			}
		}

		if flags.DryRun {
			return capi.RenderTemplateWithParameters(flags.Template, vals, creds, r, os.Stdout)
		}

		params := capi.CreatePullRequestFromTemplateParams{
			TemplateName:    flags.Template,
			ParameterValues: vals,
			RepositoryURL:   flags.RepositoryURL,
			HeadBranch:      flags.HeadBranch,
			BaseBranch:      flags.BaseBranch,
			Title:           flags.Title,
			Description:     flags.Description,
			CommitMessage:   flags.CommitMessage,
			Credentials:     creds,
		}

		return capi.CreatePullRequestFromTemplate(params, r, os.Stdout)
	}
}
