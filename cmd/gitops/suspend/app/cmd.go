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

var params app.PauseParams

var Cmd = &cobra.Command{
	Use:           "app <app-name>",
	Short:         "Suspend an application",
	Args:          cobra.MinimumNArgs(1),
	Example:       "gitops suspend app podinfo",
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

	fluxClient := flux.New(osys.New(), &runner.CLIRunner{})
	factory := services.NewFactory(fluxClient, internal.NewCLILogger(os.Stdout))

	kubeClient, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("failed to create kube client: %w", err)
	}

	appService, err := factory.GetAppService(ctx, kubeClient)
	if err != nil {
		return fmt.Errorf("failed to create app service: %w", err)
	}

	if err := appService.Pause(params); err != nil {
		return errors.Wrapf(err, "failed to pause the app %s", params.Name)
	}

	return nil
}
