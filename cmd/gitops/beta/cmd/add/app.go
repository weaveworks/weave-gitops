/*
Copyright © 2021 Weaveworks <support@weave.works>
This file is part of the Weave GitOps CLI.
*/
package add

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/apputils"
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
	Long: `This command mirrors the original app add command in 
	that it add the definition for the application to the repository 
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
	AppCmd.Flags().StringVar(&params.AppConfigUrl, "app-config-url", "", "Url of external repository (if any) which will hold automation manifests; NONE to store only in the cluster")
	AppCmd.Flags().BoolVar(&params.DryRun, "dry-run", false, "If set, 'gitops app add' will not make any changes to the system; it will just display the actions that would have been taken")
	cobra.CheckErr(AppCmd.MarkFlagRequired("app-config-url"))

	// TODO expose support for PRs
	params.AutoMerge = true
}

func runCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	params.Namespace, _ = cmd.Parent().Flags().GetString("namespace")

	if params.Url == "" && len(args) < 2 {
		return fmt.Errorf("you should choose either --url or the app directory")
	}

	// if len(args) > 0 {
	// 	path, err := filepath.Abs(args[0])
	// 	if err != nil {
	// 		return errors.Wrap(err, "failed to get absolute path for the repo directory")
	// 	}

	// 	params.Dir = path
	// }

	// if urlErr := ensureUrlIsValid(); urlErr != nil {
	// 	return urlErr
	// }

	// if readyErr := apputils.IsClusterReady(); readyErr != nil {
	// 	return readyErr
	// }

	// isHelmRepository := params.Chart != ""

	//TODO handl helm charts on add
	appService, appError := apputils.GetAppServiceForAdd(ctx, params.Url, params.AppConfigUrl, params.Namespace, false, params.DryRun)
	if appError != nil {
		return fmt.Errorf("failed to create app service: %w", appError)
	}

	params.MigrateToNewDirStructure = utils.MigrateToNewDirStructure
	if err := appService.Add(params); err != nil {
		return fmt.Errorf("failed to add the app %q: %w", params.Name, err)
	}

	return nil
}
