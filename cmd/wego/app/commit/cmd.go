package commit

import (
	"fmt"
	"os"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/wego/version"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

var params app.CommitParams

var Cmd = &cobra.Command{
	Use:           "get-commits <app-name>",
	Short:         "Get the last 10 commits for an applications git repo",
	Args:          cobra.MinimumNArgs(1),
	Example:       "wego app get-commits podinfo",
	RunE:          runCmd,
	SilenceUsage:  true,
	SilenceErrors: true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},
}

func runCmd(cmd *cobra.Command, args []string) error {
	params.Namespace, _ = cmd.Parent().Flags().GetString("namespace")
	params.Name = args[0]

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

	commits, err := appService.GetCommits(params)
	if err != nil {
		return errors.Wrapf(err, "failed to get commits for app %s", params.Name)
	}

	printCommitTable(logger, commits)

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
