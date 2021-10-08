/*
Copyright © 2021 Weaveworks <support@weave.works>
This file is part of the Weave GitOps CLI.
*/
package add

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/apputils"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

type addParams struct {
	DryRun       bool
	AppConfigUrl string
	Namespace    string
	Url          string
	Name         string
	Path         string
}

var (
	params app.AddParams
)

// appCmd represents the app command
var AppCmd = &cobra.Command{
	Use:   "app",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	RunE: runCmd,
}

func init() {
	AppCmd.Flags().StringVar(&params.Name, "name", "", "Name of application")
	AppCmd.Flags().StringVar(&params.Url, "url", "", "Url of remote repository")
	AppCmd.Flags().StringVar(&params.AppConfigUrl, "app-config-url", "", "Url of external repository (if any) which will hold automation manifests; NONE to store only in the cluster")
	AppCmd.Flags().BoolVar(&params.DryRun, "dry-run", false, "If set, 'gitops app add' will not make any changes to the system; it will just display the actions that would have been taken")
	// TODO expose support for PRs
	params.AutoMerge = true
	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// appCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// appCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func runCmd(cmd *cobra.Command, args []string) error {
	params.Name = args[0]

	ctx := context.Background()
	params.Namespace, _ = cmd.Parent().Flags().GetString("namespace")

	if params.Url != "" && len(args) > 1 {
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
		return fmt.Errorf("failed to add the app %s: %w", params.Name, err)
	}

	return nil
}
