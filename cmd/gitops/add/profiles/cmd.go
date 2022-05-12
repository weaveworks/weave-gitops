package profiles

import (
	"context"
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/names"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/pkg/services/profiles"
)

var opts profiles.Options

// AddCommand provides support for adding a profile to a cluster.
func AddCommand(endpoint *string, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "profile",
		Short:         "Add a profile to a cluster",
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: `
		# Add a profile to a cluster
		gitops add profile --name=podinfo --cluster=prod --version=1.0.0 --config-repo=ssh://git@github.com/owner/config-repo.git
		`,
		PreRunE: addProfileCmdPreRunE(endpoint, client),
		RunE:    addProfileCmdRunE(endpoint, client),
	}

	cmd.Flags().StringVar(&opts.Name, "name", "", "Name of the profile")
	cmd.Flags().StringVar(&opts.Version, "version", "latest", "Version of the profile specified as semver (e.g.: 0.1.0) or as 'latest'")
	cmd.Flags().StringVar(&opts.ConfigRepo, "config-repo", "", "URL of the external repository that contains the automation manifests")
	cmd.Flags().StringVar(&opts.Cluster, "cluster", "", "Name of the cluster to add the profile to")
	cmd.Flags().BoolVar(&opts.AutoMerge, "auto-merge", false, "If set, 'gitops add profile' will merge automatically into the repository's branch")
	internal.AddPRFlags(cmd, &opts.HeadBranch, &opts.BaseBranch, &opts.Description, &opts.Message, &opts.Title)

	requiredFlags := []string{"name", "config-repo", "cluster"}
	for _, f := range requiredFlags {
		if err := cobra.MarkFlagRequired(cmd.Flags(), f); err != nil {
			panic(fmt.Errorf("unexpected error: %w", err))
		}
	}

	return cmd
}

func addProfileCmdPreRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		if *endpoint == "" {
			return cmderrors.ErrNoWGEEndpoint
		}

		return nil
	}
}

func addProfileCmdRunE(endpoint *string, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		log := internal.NewCLILogger(os.Stdout)
		fluxClient := flux.New(&runner.CLIRunner{})
		factory := services.NewFactory(fluxClient, internal.Logr())
		providerClient := internal.NewGitProviderClient(os.Stdout, os.LookupEnv, log)

		r, err := adapters.NewHttpClient(*endpoint, client, os.Stdout)
		if err != nil {
			return err
		}

		if err := validateOptions(opts); err != nil {
			return err
		}

		if opts.Namespace, err = cmd.Flags().GetString("namespace"); err != nil {
			return err
		}

		kubeClient, err := kube.NewKubeHTTPClient()
		if err != nil {
			return fmt.Errorf("failed to create kube client: %w", err)
		}

		_, gitProvider, err := factory.GetGitClients(context.Background(), kubeClient, providerClient, services.GitConfigParams{
			ConfigRepo:       opts.ConfigRepo,
			Namespace:        opts.Namespace,
			IsHelmRepository: true,
			DryRun:           false,
		})
		if err != nil {
			return fmt.Errorf("failed to get git clients: %w", err)
		}

		return profiles.NewService(log).Add(context.Background(), r, gitProvider, opts)
	}
}

func validateOptions(opts profiles.Options) error {
	if names.ApplicationNameTooLong(opts.Name) {
		return fmt.Errorf("--name value is too long: %s; must be <= %d characters",
			opts.Name, names.MaxKubernetesResourceNameLength)
	}

	if opts.Version != "latest" {
		if _, err := semver.StrictNewVersion(opts.Version); err != nil {
			return fmt.Errorf("error parsing --version=%s: %w", opts.Version, err)
		}
	}

	return nil
}
