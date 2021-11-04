/*
Copyright © 2021 Weaveworks <support@weave.works>
This file is part of the Weave GitOps CLI.
*/
package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/gitops"
)

type params struct {
	DryRun       bool
	AppConfigURL string
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
  gitops beta install --app-config-url ssh://git@github.com/me/mygitopsrepo.git`,
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
	installCmd.Flags().StringVar(&installParams.AppConfigURL, "app-config-url", "", "URL of external repository that will hold automation manifests")
	cobra.CheckErr(installCmd.MarkFlagRequired("app-config-url"))
}

func installRunCmd(cmd *cobra.Command, args []string) error {
	namespace, _ := cmd.Parent().Flags().GetString("namespace")

	log := logger.NewCLILogger(os.Stdout)
	fluxClient := flux.New(osys.New(), &runner.CLIRunner{})

	k, rawClient, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("error creating k8s http client: %w", err)
	}

	normalizedURL, err := gitproviders.NewRepoURL(installParams.AppConfigURL)
	if err != nil {
		return fmt.Errorf("failed to normalize URL %q: %w", installParams.AppConfigURL, err)
	}

	providerClient := internal.NewGitProviderClient(os.Stdout, os.LookupEnv, auth.NewAuthCLIHandler, log)

	var gitProvider gitproviders.GitProvider
	if installParams.DryRun {
		if gitProvider, err = gitproviders.NewDryRun(); err != nil {
			return fmt.Errorf("error creating git provider client: %w", err)
		}
	} else {
		if gitProvider, err = providerClient.GetProvider(normalizedURL, gitproviders.GetAccountType); err != nil {
			return fmt.Errorf("error obtaining git provider token: %w", err)
		}
	}

	authService, err := auth.NewAuthService(fluxClient, rawClient, gitProvider, log)
	if err != nil {
		return fmt.Errorf("error creating auth service: %w", err)
	}

	clusterName, err := k.GetClusterName(context.Background())
	if err != nil {
		log.Warningf("Cluster name not found, using default : %v", err)

		clusterName = "default"
	}

	gitClient, err := authService.CreateGitClient(context.Background(), normalizedURL, clusterName, namespace)
	if err != nil {
		return fmt.Errorf("failed to create git client for repo %q : %w", installParams.AppConfigURL, err)
	}

	gitopsService := gitops.New(log, fluxClient, k)

	manifests, err := gitopsService.Install(gitClient, gitProvider, gitops.InstallParams{
		Namespace:    namespace,
		DryRun:       installParams.DryRun,
		AppConfigURL: installParams.AppConfigURL,
	})
	if err != nil {
		return err
	}

	if installParams.DryRun {
		fmt.Println(string(manifests))
	}

	return nil
}
