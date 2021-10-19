/*
Copyright © 2021 Weaveworks <support@weave.works>
This file is part of the Weave GitOps CLI.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/pkg/apputils"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
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
	installCmd.Flags().BoolVar(&installParams.DryRun, "dry-run", false, "outputs all the manifests that would be installed")
	installCmd.Flags().StringVar(&installParams.AppConfigURL, "app-config-url", "", "URL of external repository that will hold automation manifests")
	cobra.CheckErr(installCmd.MarkFlagRequired("app-config-url"))
}

func installRunCmd(cmd *cobra.Command, args []string) error {
	namespace, _ := cmd.Parent().Flags().GetString("namespace")

	clients, err := apputils.GetBaseClients()
	if err != nil {
		return err
	}

	normalizedURL, err := gitproviders.NewRepoURL(installParams.AppConfigURL)
	if err != nil {
		return fmt.Errorf("failed to normalize URL %q: %w", installParams.AppConfigURL, err)
	}

	authHandler, err := auth.NewAuthCLIHandler(normalizedURL.Provider())
	if err != nil {
		return fmt.Errorf("error initializing cli auth handler: %w", err)
	}

	gitProvider, err := auth.InitGitProvider(normalizedURL, clients.Osys, clients.Logger, authHandler, gitproviders.GetAccountType)
	if err != nil {
		return fmt.Errorf("error obtaining git provider token: %w", err)
	}

	gitopsService := gitops.New(clients.Logger, clients.Flux, clients.Kube, gitProvider, nil)

	manifests, err := gitopsService.Install(gitops.InstallParams{
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
