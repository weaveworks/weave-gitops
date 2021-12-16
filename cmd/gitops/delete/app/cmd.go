package app

// Provides support for removing an application from gitops management.

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"k8s.io/apimachinery/pkg/types"

	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
)

var params app.RemoveParams

var Cmd = &cobra.Command{
	Use:   "app <app name>",
	Short: "Delete an app from a gitops cluster",
	Long: strings.TrimSpace(dedent.Dedent(`
        Deletes an application from a gitops cluster so it will no longer be managed via GitOps
    `)),
	Example: `
  # Delete application from gitops control via immediate commit
  gitops delete app podinfo
`,
	Args:          cobra.MinimumNArgs(1),
	RunE:          runCmd,
	SilenceUsage:  true,
	SilenceErrors: true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},
}

func init() {
	Cmd.Flags().BoolVar(&params.DryRun, "dry-run", false, "If set, 'gitops delete app' will not make any changes to the system; it will just display the actions that would have been taken")
	Cmd.Flags().BoolVar(&params.AutoMerge, "auto-merge", false, "If set, 'gitops delete app' will merge changes automatically to the config repository")
}

func runCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	params.Name = args[0]
	params.Namespace, _ = cmd.Parent().Flags().GetString("namespace")

	log := internal.NewCLILogger(os.Stdout)
	factory := services.NewFactory(flux.New(osys.New(), &runner.CLIRunner{}), log)

	appService, err := factory.GetAppService(ctx)
	if err != nil {
		return fmt.Errorf("failed to create app service: %w", err)
	}

	appContent, err := appService.Get(types.NamespacedName{Name: params.Name, Namespace: params.Namespace})
	if err != nil {
		return fmt.Errorf("unable to get application for %s %w", params.Name, err)
	}

	providerClient := internal.NewGitProviderClient(os.Stdout, os.LookupEnv, auth.NewAuthCLIHandler, log)

	gitClient, gitProvider, err := factory.GetGitClients(ctx, providerClient, services.NewGitConfigParamsFromApp(appContent, params.DryRun))
	if err != nil {
		return fmt.Errorf("failed to get git clients: %w", err)
	}

	if err := appService.Remove(gitClient, gitProvider, params); err != nil {
		return errors.Wrapf(err, "failed to remove the app %s", params.Name)
	}

	return nil
}
