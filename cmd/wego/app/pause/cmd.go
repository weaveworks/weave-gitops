package pause

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/wego/version"
	"github.com/weaveworks/weave-gitops/pkg/cliutils"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
)

var params app.PauseParams

var Cmd = &cobra.Command{
	Use:           "pause <app-name>",
	Short:         "Pause an application",
	Args:          cobra.MinimumNArgs(1),
	Example:       "wego app pause podinfo",
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

	osysClient, fluxClient, kubeClient, logger := cliutils.GetBaseClients()
	appService := app.New(logger, nil, nil, nil, fluxClient, kubeClient, osysClient)

	if err := appService.Pause(params); err != nil {
		return errors.Wrapf(err, "failed to pause the app %s", params.Name)
	}

	return nil
}
