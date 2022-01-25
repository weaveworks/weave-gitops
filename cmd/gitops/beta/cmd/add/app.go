/*
Copyright © 2021 Weaveworks <support@weave.works>
This file is part of the Weave GitOps CLI.
*/
package add

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

var (
	params app.AddParams
)

// appCmd represents the app command
var AppCmd = &cobra.Command{
	Use:   "app",
	Short: "Adds an application workload to the GitOps repository",
	Long: `This command mirrors the original add app command in
	that it adds the definition for the application to the repository
	and sets up syncing into a cluster. It uses the new directory
	structure.`,
	RunE: runCmd,
	Args: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errors.New("add app requires a name argument")
		}
		params.Name = args[0]
		return nil
	},
}

func init() {
	AppCmd.Flags().StringVar(&params.Url, "url", "", "Url of remote repository")
	AppCmd.Flags().StringVar(&params.ConfigRepo, "config-repo", "", "Url of external repository (if any) which will hold automation manifests; NONE to store only in the cluster")
	AppCmd.Flags().BoolVar(&params.DryRun, "dry-run", false, "If set, 'gitops add app' will not make any changes to the system; it will just display the actions that would have been taken")
	cobra.CheckErr(AppCmd.MarkFlagRequired("config-repo"))

	// TODO expose support for PRs
	params.AutoMerge = true
}

func runCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	params.Namespace, _ = cmd.Parent().Flags().GetString("namespace")

	if params.Url == "" && len(args) < 2 {
		return fmt.Errorf("you should choose either --url or the app directory")
	}

	log := internal.NewCLILogger(os.Stdout)
	fluxClient := flux.New(osys.New(), &runner.CLIRunner{})
	factory := services.NewFactory(fluxClient, log)

	providerClient := internal.NewGitProviderClient(os.Stdout, os.LookupEnv, auth.NewAuthCLIHandler, log)

	kubeClient, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("failed to create kube client: %w", err)
	}

	appService, err := factory.GetAppService(ctx, kubeClient)
	if err != nil {
		return fmt.Errorf("failed to create app service: %w", err)
	}

	gitClient, gitProvider, err := factory.GetGitClients(ctx, kubeClient, providerClient, services.GitConfigParams{
		URL:              params.Url,
		ConfigRepo:       params.ConfigRepo,
		Namespace:        params.Namespace,
		IsHelmRepository: params.IsHelmRepository(),
		DryRun:           params.DryRun,
	})
	if err != nil {
		return fmt.Errorf("failed to get git clients: %w", err)
	}

	params.MigrateToNewDirStructure = utils.MigrateToNewDirStructure
	if err := appService.Add(gitClient, gitProvider, params); err != nil {
		return fmt.Errorf("failed to add the app %q: %w", params.Name, err)
	}

	return nil
}
