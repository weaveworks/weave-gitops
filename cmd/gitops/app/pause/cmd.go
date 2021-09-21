package pause

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/pkg/apputils"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
)

var params app.PauseParams

var Cmd = &cobra.Command{
	Use:           "pause <app-name>",
	Short:         "Pause an application",
	Args:          cobra.MinimumNArgs(1),
	Example:       "gitops app pause podinfo",
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

	appService, appError := apputils.GetAppService(ctx, params.Name, params.Namespace)
	if appError != nil {
		return fmt.Errorf("failed to create app service: %w", appError)
	}

	if err := appService.Pause(params); err != nil {
		return errors.Wrapf(err, "failed to pause the app %s", params.Name)
	}

	return nil
}
