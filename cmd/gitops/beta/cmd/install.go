/*
Copyright © 2021 Weaveworks <support@weave.works>
This file is part of the Weave GitOps CLI.
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/gitops"
)

type params struct {
	DryRun     bool
	ConfigRepo string
}

var (
	installParams params
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install or upgrade GitOps",
	Long: `The beta install command deploys GitOps in the specified namespace,
adds a cluster entry to the GitOps repo, and persists the GitOps runtime into the
repo.`,
	Example: `  # Install GitOps in the wego-system namespace
  gitops beta install --config-repo ssh://git@github.com/me/mygitopsrepo.git`,
	RunE:          installRunCmd,
	SilenceErrors: true,
	SilenceUsage:  true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},
}

func init() {
	Cmd.AddCommand(installCmd)
	installCmd.Flags().BoolVar(&installParams.DryRun, "dry-run", false, "Outputs all the manifests that would be installed")
	installCmd.Flags().StringVar(&installParams.ConfigRepo, "config-repo", "", "URL of external repository that will hold automation manifests")
	cobra.CheckErr(installCmd.MarkFlagRequired("config-repo"))
}

func installRunCmd(cmd *cobra.Command, args []string) error {
	namespace, _ := cmd.Parent().Flags().GetString("namespace")

	log := internal.NewCLILogger(os.Stdout)
	fluxClient := flux.New(osys.New(), &runner.CLIRunner{})

	k, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("error creating k8s http client: %w", err)
	}

	factory := services.NewFactory(fluxClient, log)
	providerClient := internal.NewGitProviderClient(os.Stdout, os.LookupEnv, auth.NewAuthCLIHandler, log)

	gitopsService := gitops.New(log, fluxClient, k)

	gitOpsParams := gitops.InstallParams{
		Namespace:  namespace,
		DryRun:     installParams.DryRun,
		ConfigRepo: installParams.ConfigRepo,
	}

	manifests, err := gitopsService.Install(gitOpsParams)
	if err != nil {
		return err
	}

	gitClient, gitProvider, err := factory.GetGitClients(context.Background(), k, providerClient, services.GitConfigParams{
		ConfigRepo: installParams.ConfigRepo,
		Namespace:  namespace,
		DryRun:     installParams.DryRun,
	})
	if err != nil {
		return fmt.Errorf("error creating git clients: %w", err)
	}

	manifests, err = gitopsService.StoreManifests(gitClient, gitProvider, gitOpsParams, manifests)
	if err != nil {
		return fmt.Errorf("error storing manifests: %w", err)
	}

	if installParams.DryRun {
		for _, manifest := range manifests {
			fmt.Println(string(manifest))
		}
	}

	return nil
}
