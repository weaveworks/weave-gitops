package profile

import (
	"context"
	"fmt"
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/profile"
)

var params profile.AddParams

// Provides support for adding a profile to gitops management.
func ProfileCommand(client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Add a profile to the cluster",
		RunE:  addProfileCmdRunE(client),
	}

	cmd.Flags().StringVar(&params.Name, "name", "", "Name of the profile")
	cmd.Flags().StringVar(&params.Version, "version", "", "Version of the profile")
	cmd.Flags().StringVar(&params.ConfigRepo, "config-repo", "", "URL of external repository (if any) which will hold automation manifests")
	cmd.Flags().StringVar(&params.Cluster, "cluster", "", "Name of the cluster to add the profile to")
	cmd.Flags().BoolVar(&params.AutoMerge, "auto-merge", false, "If set, 'gitops add profile' will merge automatically into the set --branch")

	return cmd
}

func addProfileCmdRunE(client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		log := internal.NewCLILogger(os.Stdout)
		fluxClient := flux.New(osys.New(), &runner.CLIRunner{})
		factory := services.NewFactory(fluxClient, log)
		providerClient := internal.NewGitProviderClient(os.Stdout, os.LookupEnv, auth.NewAuthCLIHandler, log)

		ctx := context.Background()
		_, gitProvider, err := factory.GetGitClients(ctx, providerClient, services.GitConfigParams{
			ConfigRepo:       params.ConfigRepo,
			Namespace:        "weave-system",
			IsHelmRepository: true,
			DryRun:           false,
		})
		if err != nil {
			return fmt.Errorf("failed to get git clients: %w", err)
		}

		profileService := profile.NewService(ctx, log, osys.New())
		if err := profileService.Add(gitProvider, params); err != nil {
			return errors.Wrapf(err, "failed to add the profile %s", params.Name)
		}

		return nil
	}
}
