package unpause

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/wego/version"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
)

var params app.UnpauseParams

var Cmd = &cobra.Command{
	Use:           "unpause <app-name>",
	Short:         "unpause an app",
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

	cliRunner := &runner.CLIRunner{}
	fluxClient := flux.New(cliRunner)
	logger := logger.New(os.Stdout)
	kubeClient, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("error initializing kube client: %w", err)
	}

	appService := app.New(logger, nil, fluxClient, kubeClient)

	if err := appService.Unpause(params); err != nil {
		return errors.Wrapf(err, "failed to unpause the app %s", params.Name)
	}

	return nil
}
