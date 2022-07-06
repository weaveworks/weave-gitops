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
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/templates"
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

func ClusterCommand(opts *config.Options, client *resty.Client) *cobra.Command {
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
		PreRunE:       getClusterCmdPreRunE(&opts.Endpoint),
		RunE:          getClusterCmdRunE(opts, client),
	}

	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "View the populated template without creating a pull request")
	cmd.Flags().StringVar(&flags.RepositoryURL, "url", "", "URL of remote repository to create the pull request")
	cmd.Flags().StringVar(&flags.Credentials, "set-credentials", "", "The CAPI credentials to use")
	cmd.Flags().StringArrayVar(&flags.Profiles, "profile", []string{}, "Set profiles values files on the command line (--profile 'name=foo-profile,version=0.0.1' --profile 'name=bar-profile,values=bar-values.yaml')")
	internal.AddTemplateFlags(cmd, &flags.Template, &flags.ParameterValues)
	internal.AddPRFlags(cmd, &flags.HeadBranch, &flags.BaseBranch, &flags.Description, &flags.CommitMessage, &flags.Title)

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
		r, err := adapters.NewHttpClient(opts.Endpoint, opts.Username, opts.Password, client, os.Stdout)
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

		creds := templates.Credentials{}
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
			return templates.RenderTemplateWithParameters(templates.CAPITemplateKind, flags.Template, vals, creds, r, os.Stdout)
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
			TemplateKind:     templates.CAPITemplateKind.String(),
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

		return templates.CreatePullRequestFromTemplate(params, r, os.Stdout)
	}
}

func parseProfileFlags(profiles []string) ([]templates.ProfileValues, error) {
	var profilesValues []templates.ProfileValues

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

		var profileValues templates.ProfileValues

		err = json.Unmarshal(profileJson, &profileValues)
		if err != nil {
			return nil, err
		}

		profilesValues = append(profilesValues, profileValues)
	}

	return profilesValues, nil
}
