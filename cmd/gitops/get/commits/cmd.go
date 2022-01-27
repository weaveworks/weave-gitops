package commits

import (
	"context"
	"fmt"
	"os"

	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
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

	log := internal.NewCLILogger(os.Stdout)
	fluxClient := flux.New(osys.New(), &runner.CLIRunner{})
	factory := services.NewFactory(fluxClient, log)

	kubeClient, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("failed to create kube client: %w", err)
	}

	appService, err := factory.GetAppService(ctx, kubeClient)
	if err != nil {
		return fmt.Errorf("failed to create app service: %w", err)
	}

	appContent, err := appService.Get(types.NamespacedName{Name: params.Name, Namespace: params.Namespace})
	if err != nil {
		return fmt.Errorf("unable to get application for %s %w", params.Name, err)
	}

	providerClient := internal.NewGitProviderClient(os.Stdout, os.LookupEnv, auth.NewAuthCLIHandler, log)

	_, gitProvider, err := factory.GetGitClients(ctx, kubeClient, providerClient, services.NewGitConfigParamsFromApp(appContent, false))
	if err != nil {
		return fmt.Errorf("failed to get git clients: %w", err)
	}

	commits, err := appService.GetCommits(gitProvider, params, appContent)
	if err != nil {
		return errors.Wrapf(err, "failed to get commits for app %s", params.Name)
	}

	printCommitTable(log, commits)

	return nil
}

func printCommitTable(log logger.Logger, commits []gitprovider.Commit) {
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

	utils.PrintTable(log, header, rows)
}
