package terraform

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/templates"
)

type terraformCommandFlags struct {
	Template              string
	ParameterValues       []string
	RepositoryURL         string
	BaseBranch            string
	HeadBranch            string
	Title                 string
	Description           string
	CommitMessage         string
	InsecureSkipTlsVerify bool
}

var flags terraformCommandFlags

func AddCommand(opts *config.Options, client *adapters.HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "terraform",
		Short: "Add a new Terraform resource using a TF template",
		Example: `
# Add a new Terraform resource using a TF template
gitops add terraform --from-template <template-name> --set key=val
		`,
		SilenceUsage:  true,
		SilenceErrors: true,
		PreRunE:       addTerraformCmdPreRunE(&opts.Endpoint),
		RunE:          addTerraformCmdRunE(opts, client),
	}

	cmd.Flags().StringVar(&flags.RepositoryURL, "url", "", "URL of remote repository to create the pull request")
	internal.AddTemplateFlags(cmd, &flags.Template, &flags.ParameterValues)
	internal.AddPRFlags(cmd, &flags.HeadBranch, &flags.BaseBranch, &flags.Description, &flags.CommitMessage, &flags.Title)

	return cmd
}

func addTerraformCmdPreRunE(endpoint *string) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		if *endpoint == "" {
			return cmderrors.ErrNoWGEEndpoint
		}

		return nil
	}
}

func addTerraformCmdRunE(opts *config.Options, client *adapters.HTTPClient) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		err := client.ConfigureClientWithOptions(opts, os.Stdout)
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

		params := templates.CreatePullRequestFromTemplateParams{
			GitProviderToken: token,
			TemplateName:     flags.Template,
			TemplateKind:     templates.GitOpsTemplateKind.String(),
			ParameterValues:  vals,
			RepositoryURL:    flags.RepositoryURL,
			HeadBranch:       flags.HeadBranch,
			BaseBranch:       flags.BaseBranch,
			Title:            flags.Title,
			Description:      flags.Description,
			CommitMessage:    flags.CommitMessage,
		}

		return templates.CreatePullRequestFromTemplate(params, client, os.Stdout)
	}
}
