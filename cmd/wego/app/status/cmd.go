package status

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/apputils"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
)

var Cmd = &cobra.Command{
	Use:           "status <app-name>",
	Short:         "Get status of a workload under wego control",
	Args:          cobra.MinimumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	Example:       "wego app status podinfo",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		params := app.StatusParams{}

		params.Name = args[0]
		params.Namespace, _ = cmd.Parent().Parent().Flags().GetString("namespace")

		appService, appError := apputils.GetAppService(ctx, params.Name, params.Namespace)
		if appError != nil {
			return fmt.Errorf("failed to create app service: %w", appError)
		}

		fluxOutput, lastSuccessReconciliation, err := appService.Status(params)
		if err != nil {
			return fmt.Errorf("failed getting application status: %w", err)
		}

		logger := apputils.GetLogger()
		logger.Printf("Last successful reconciliation: %s\n\n", lastSuccessReconciliation)
		logger.Println(fluxOutput)

		return nil
	},
}
