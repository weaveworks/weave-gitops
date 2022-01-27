package status

import (
	"context"
	"fmt"
	"os"

	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
)

var Cmd = &cobra.Command{
	Use:           "status <app-name>",
	Short:         "Get status of a workload under gitops control",
	Args:          cobra.MinimumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	Example:       "gitops app status podinfo",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()

		params := app.StatusParams{}

		params.Name = args[0]
		params.Namespace, _ = cmd.Parent().Parent().Flags().GetString("namespace")

		log := internal.NewCLILogger(os.Stdout)
		fluxClient := flux.New(osys.New(), &runner.CLIRunner{})
		appFactory := services.NewFactory(fluxClient, log)

		kubeClient, _, err := kube.NewKubeHTTPClient()
		if err != nil {
			return fmt.Errorf("failed to create kube client: %w", err)
		}
		appService, err := appFactory.GetAppService(ctx, kubeClient)
		if err != nil {
			return fmt.Errorf("failed to create app service: %w", err)
		}

		fluxOutput, lastSuccessReconciliation, err := appService.Status(params)
		if err != nil {
			return fmt.Errorf("failed getting application status: %w", err)
		}

		log.Printf("Last successful reconciliation: %s\n\n", lastSuccessReconciliation)
		log.Println(fluxOutput)

		return nil
	},
}
