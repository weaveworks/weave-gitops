package clusters

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/capi"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
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
	Profiles        []string
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

# Add a new cluster supplied with profiles versions and values files
gitops add cluster --from-template <template-name> \
--profile 'name=foo-profile,version=0.0.1' --profile 'name=bar-profile,values=bar-values.yaml
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
	cmd.Flags().StringArrayVar(&flags.Profiles, "profile", []string{}, "Set profiles values files on the command line (--profile 'name=foo-profile,version=0.0.1' --profile 'name=bar-profile,values=bar-values.yaml')")

	return cmd
}

func getClusterCmdPreRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		if *endpoint == "" {
			return cmderrors.ErrNoWGEEndpoint
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

		profilesValues, err := parseProfileFlags(flags.Profiles)
		if err != nil {
			return fmt.Errorf("error parsing profiles: %w", err)
		}

		if flags.DryRun {
			return capi.RenderTemplateWithParameters(flags.Template, vals, creds, r, os.Stdout)
		}

		if flags.RepositoryURL == "" {
			return cmderrors.ErrNoURL
		}

		url, err := gitproviders.NewRepoURL(flags.RepositoryURL, true)
		if err != nil {
			return fmt.Errorf("cannot parse url: %w", err)
		}

		token, err := internal.GetToken(url, os.Stdout, os.LookupEnv, auth.NewAuthCLIHandler, internal.NewCLILogger(os.Stdout))
		if err != nil {
			return err
		}

		params := capi.CreatePullRequestFromTemplateParams{
			GitProviderToken: token,
			TemplateName:     flags.Template,
			ParameterValues:  vals,
			RepositoryURL:    flags.RepositoryURL,
			HeadBranch:       flags.HeadBranch,
			BaseBranch:       flags.BaseBranch,
			Title:            flags.Title,
			Description:      flags.Description,
			CommitMessage:    flags.CommitMessage,
			Credentials:      creds,
			ProfileValues:    profilesValues,
		}

		return capi.CreatePullRequestFromTemplate(params, r, os.Stdout)
	}
}

func parseProfileFlags(profiles []string) ([]capi.ProfileValues, error) {
	var profilesValues []capi.ProfileValues

	// Validate values include alphanumeric or - or .
	r := regexp.MustCompile(`^[A-Za-z0-9]([A-Za-z0-9.-]*[A-Za-z0-9])?$`)

	for _, p := range flags.Profiles {
		valuesPairs := strings.Split(p, ",")
		profileMap := make(map[string]string)

		for _, pair := range valuesPairs {
			fmt.Println(pair)
			kv := strings.Split(pair, "=")

			if kv[0] != "name" && kv[0] != "version" && kv[0] != "values" {
				return nil, fmt.Errorf("invalid key: %s", kv[0])
			} else if !r.MatchString(kv[1]) {
				return nil, fmt.Errorf("invalid value for %s: %s", kv[0], kv[1])
			} else {
				profileMap[kv[0]] = kv[1]
			}
		}

		profileJson, err := json.Marshal(profileMap)
		if err != nil {
			return nil, err
		}

		var profileValues capi.ProfileValues

		err = json.Unmarshal(profileJson, &profileValues)
		if err != nil {
			return nil, err
		}

		profilesValues = append(profilesValues, profileValues)
	}

	return profilesValues, nil
}
