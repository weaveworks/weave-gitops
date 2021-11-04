package commits

import (
	"context"
	"fmt"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/apputils"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/utils"

	"k8s.io/apimachinery/pkg/types"
)

var Cmd = &cobra.Command{
	Use:   "commits",
	Short: "Get most recent commits for an application",
	Example: `
# Get last 10 commits for an application
gitops get commits <app-name>`,
	SilenceUsage:  true,
	SilenceErrors: true,
	Args:          cobra.ExactArgs(1),
	RunE:          runCmd,
}

func runCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	params := app.CommitParams{}
	params.Name = args[0]
	params.Namespace, _ = cmd.Parent().Flags().GetString("namespace")
	// Hardcode PageSize and PageToken until there is a plan around pagination for cli
	params.PageSize = 10
	params.PageToken = 0

	appService, appError := apputils.GetAppService(ctx, params.Name, params.Namespace)
	if appError != nil {
		return fmt.Errorf("failed to create app service: %w", appError)
	}

	appContent, err := appService.Get(types.NamespacedName{Name: params.Name, Namespace: params.Namespace})
	if err != nil {
		return fmt.Errorf("unable to get application for %s %w", params.Name, err)
	}

	logger := apputils.GetLogger()

	commits, err := appService.GetCommits(params, appContent)
	if err != nil {
		return errors.Wrapf(err, "failed to get commits for app %s", params.Name)
	}

	printCommitTable(logger, commits)

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
