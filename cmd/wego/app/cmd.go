package app

import (
	"context"
	"fmt"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/add"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/list"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/pause"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/remove"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/status"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/unpause"
	"github.com/weaveworks/weave-gitops/pkg/apputils"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
)

var ApplicationCmd = &cobra.Command{
	Use:   "app",
	Short: "Manages your applications",
	Example: `
  # Get last 10 commits for an application
  wego app <app-name> get commits

  # Add an application to wego from local git repository
  wego app add . --name <app-name>

  # Remove an application from wego
  wego app remove <app-name>

  # Status an application under wego control
  wego app status <app-name>

  # List applications under wego control
  wego app list

  # Pause gitops automation
  wego app pause <app-name>

  # Unpause gitops automation
  wego app unpause <app-name>`,
	Args: cobra.MinimumNArgs(3),
	RunE: runCmd,
}

func init() {
	ApplicationCmd.AddCommand(add.Cmd)
	ApplicationCmd.AddCommand(remove.Cmd)
	ApplicationCmd.AddCommand(list.Cmd)
	ApplicationCmd.AddCommand(status.Cmd)
	ApplicationCmd.AddCommand(pause.Cmd)
	ApplicationCmd.AddCommand(unpause.Cmd)
}

func runCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	params := app.CommitParams{}
	params.Name = args[0]
	params.Namespace, _ = cmd.Parent().Flags().GetString("namespace")
	// Hardcode PageSize and PageToken until there is a plan around pagination for cli
	params.PageSize = 10
	params.PageToken = 0

	command := args[1]
	object := args[2]

	appService, appError := apputils.GetAppService(ctx, params.Name, params.Namespace)
	if appError != nil {
		return fmt.Errorf("failed to create app service: %w", appError)
	}

	appContent, err := appService.Get(types.NamespacedName{Name: params.Name, Namespace: params.Namespace})
	if err != nil {
		return fmt.Errorf("unable to get application for %s %w", params.Name, err)
	}

	if appContent.Spec.SourceType == wego.SourceTypeHelm {
		return fmt.Errorf("unable to get commits for a helm chart")
	}

	if command != "get" {
		_ = cmd.Help()
		return fmt.Errorf("invalid command %s", command)
	}

	logger := apputils.GetLogger()

	switch object {
	case "commits":
		commits, err := appService.GetCommits(params, appContent)
		if err != nil {
			return errors.Wrapf(err, "failed to get commits for app %s", params.Name)
		}
		printCommitTable(logger, commits)
	default:
		_ = cmd.Help()
		return fmt.Errorf("unkown resource type \"%s\"", object)
	}

	return nil
}

func printCommitTable(logger logger.Logger, commits []gitprovider.Commit) {
	header := []string{"Commit Hash", "Created At", "Author", "Message", "URL"}
	rows := [][]string{}
	for _, commit := range commits {
		c := commit.Get()
		rows = append(rows, []string{
			utils.ConvertCommitHashToShort(c.Sha),
			utils.CleanCommitCreatedAt(c.CreatedAt),
			c.Author,
			utils.CleanCommitMessage(c.Message),
			utils.ConvertCommitURLToShort(c.URL),
		})
	}

	utils.PrintTable(logger, header, rows)
}
