package profiles

import (
	"context"
	"fmt"
	"os"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/logger"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/adapters"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/pkg/services/profiles"
)

var profileOpts profiles.Options

// UpdateCommand provides support for updating a profile that is installed on a cluster.
func UpdateCommand(opts *config.Options, client *adapters.HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "profile",
		Short:         "Update a profile installation",
		SilenceUsage:  true,
		SilenceErrors: true,
		Example: `
	# Update a profile that is installed on a cluster
	gitops update profile --name=podinfo --cluster=prod --config-repo=ssh://git@github.com/owner/config-repo.git  --version=1.0.0
		`,
		PreRunE: updateProfileCmdPreRunE(&opts.Endpoint),
		RunE:    updateProfileCmdRunE(opts, client),
	}

	cmd.Flags().StringVar(&profileOpts.Name, "name", "", "Name of the profile")
	cmd.Flags().StringVar(&profileOpts.Version, "version", "latest", "Version of the profile specified as semver (e.g.: 0.1.0) or as 'latest'")
	cmd.Flags().StringVar(&profileOpts.ConfigRepo, "config-repo", "", "URL of the external repository that contains the automation manifests")
	cmd.Flags().StringVar(&profileOpts.Cluster, "cluster", "", "Name of the cluster where the profile is installed")
	cmd.Flags().BoolVar(&profileOpts.AutoMerge, "auto-merge", false, "If set, 'gitops update profile' will merge automatically into the repository's branch")
	internal.AddPRFlags(cmd, &profileOpts.HeadBranch, &profileOpts.BaseBranch, &profileOpts.Description, &profileOpts.Message, &profileOpts.Title)

	requiredFlags := []string{"name", "config-repo", "cluster", "version"}
	for _, f := range requiredFlags {
		if err := cobra.MarkFlagRequired(cmd.Flags(), f); err != nil {
			panic(fmt.Errorf("unexpected error: %w", err))
		}
	}

	return cmd
}

func updateProfileCmdPreRunE(endpoint *string) func(*cobra.Command, []string) error {
	return func(c *cobra.Command, s []string) error {
		if *endpoint == "" {
			return cmderrors.ErrNoWGEEndpoint
		}

		return nil
	}
}

func updateProfileCmdRunE(opts *config.Options, client *adapters.HTTPClient) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		log := logger.NewCLILogger(os.Stdout)
		fluxClient := flux.New(&runner.CLIRunner{})
		factory := services.NewFactory(fluxClient, logger.Logr())
		providerClient := internal.NewGitProviderClient(os.Stdout, os.LookupEnv, log)

		err := client.ConfigureClientWithOptions(opts, os.Stdout)
		if err != nil {
			return err
		}

		if profileOpts.Version != "latest" {
			if _, err := semver.StrictNewVersion(profileOpts.Version); err != nil {
				return fmt.Errorf("error parsing --version=%s: %w", profileOpts.Version, err)
			}
		}

		if profileOpts.Namespace, err = cmd.Flags().GetString("namespace"); err != nil {
			return err
		}

		kubeClient, err := kube.NewKubeHTTPClient()
		if err != nil {
			return fmt.Errorf("failed to create kube client: %w", err)
		}

		_, gitProvider, err := factory.GetGitClients(context.Background(), kubeClient, providerClient, services.GitConfigParams{
			ConfigRepo:       profileOpts.ConfigRepo,
			Namespace:        profileOpts.Namespace,
			IsHelmRepository: true,
			DryRun:           false,
		})
		if err != nil {
			return fmt.Errorf("failed to get git clients: %w", err)
		}

		return profiles.NewService(log).Update(context.Background(), client, gitProvider, profileOpts)
	}
}
