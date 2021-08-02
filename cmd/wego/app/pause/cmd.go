package pause

import (
	"fmt"
	"os"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/wego/version"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
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

	cliRunner := &runner.CLIRunner{}
	osysClient := osys.New()
	fluxClient := flux.New(osysClient, cliRunner)
	logger := logger.NewCLILogger(os.Stdout)
	kubeClient, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("error initializing kube client: %w", err)
	}

	appService := app.New(logger, nil, fluxClient, kubeClient, osysClient)

	if err := appService.Pause(params); err != nil {
		return errors.Wrapf(err, "failed to pause the app %s", params.Name)
	}

	return nil
}
