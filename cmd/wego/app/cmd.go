package app

import (
	"fmt"
	"os"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/add"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/list"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/pause"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/remove"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/status"
	"github.com/weaveworks/weave-gitops/cmd/wego/app/unpause"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/utils"
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
	params := app.CommitParams{}
	params.Name = args[0]
	params.Namespace, _ = cmd.Parent().Flags().GetString("namespace")

	command := args[1]
	object := args[2]

	cliRunner := &runner.CLIRunner{}
	osysClient := osys.New()
	fluxClient := flux.New(osysClient, cliRunner)
	logger := logger.NewCLILogger(os.Stdout)
	kubeClient, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("error initializing kube client: %w", err)
	}

	token, err := osysClient.GetGitProviderToken()
	if err != nil {
		return err
	}

	params.GitProviderToken = token

	appService := app.New(logger, nil, fluxClient, kubeClient, osysClient)

	if command != "get" {
		_ = cmd.Help()
		return fmt.Errorf("invalid command %s", command)
	}

	switch object {
	case "commits":
		commits, err := appService.GetCommits(params)
		if err != nil {
			_ = cmd.Help()
			return errors.Wrapf(err, "failed to get commits for app %s", params.Name)
		}
		printCommitTable(logger, commits)
	default:
		return fmt.Errorf("unkown resource type \"%s\"", object)
	}

	return nil
}

func printCommitTable(logger logger.Logger, commits []gitprovider.Commit) {
	header := []string{"Commit Hash", "Author", "Message", "Created At"}
	rows := [][]string{}
	for _, commit := range commits {
		c := commit.Get()
		rows = append(rows, []string{c.Sha, c.Author, c.Message, c.CreatedAt.String()})
	}

	utils.PrintTable(logger, header, rows)
}
