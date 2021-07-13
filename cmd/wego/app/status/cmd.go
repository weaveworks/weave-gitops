package status

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
)

var Cmd = &cobra.Command{
	Use:     "status <app-name>",
	Short:   "Get status of an app",
	Args:    cobra.MinimumNArgs(1),
	Example: "wego app status podinfo",
	RunE: func(cmd *cobra.Command, args []string) error {
		params.Namespace, _ = cmd.Parent().Parent().Flags().GetString("namespace")
		params.AppName = args[0]
		cliRunner := &runner.CLIRunner{}
		fluxClient := flux.New(cliRunner)
		kubeClient := kube.New(cliRunner)

		appService := app.New(logger.New(os.Stdout), nil, fluxClient, kubeClient, nil)
		return appService.Status(params)
	},
}

var params app.StatusParams
