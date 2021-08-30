package unpause

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/wego/version"
	"github.com/weaveworks/weave-gitops/pkg/cliutils"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
)

var params app.UnpauseParams

var Cmd = &cobra.Command{
	Use:           "unpause <app-name>",
	Short:         "Unpause an application",
	Args:          cobra.MinimumNArgs(1),
	Example:       "wego app unpause podinfo",
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

	if err := appService.Unpause(params); err != nil {
		return errors.Wrapf(err, "failed to unpause the app %s", params.Name)
	}

	return nil
}
