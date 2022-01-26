package app

import (
	"context"
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
)

var params app.UnpauseParams

var Cmd = &cobra.Command{
	Use:           "app <app-name>",
	Short:         "Resume an application",
	Args:          cobra.MinimumNArgs(1),
	Example:       "gitops resume app podinfo",
	RunE:          runCmd,
	SilenceUsage:  true,
	SilenceErrors: true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},
}

func runCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	params.Namespace, _ = cmd.Parent().Flags().GetString("namespace")
	params.Name = args[0]

	appFactory := services.NewFactory(flux.New(osys.New(), &runner.CLIRunner{}), internal.NewCLILogger(os.Stdout))

	kubeClient, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("failed to create kube client: %w", err)
	}

	appService, err := appFactory.GetAppService(ctx, kubeClient)
	if err != nil {
		return fmt.Errorf("failed to create app service: %w", err)
	}

	if err := appService.Unpause(params); err != nil {
		return errors.Wrapf(err, "failed to unpause the app %s", params.Name)
	}

	return nil
}
