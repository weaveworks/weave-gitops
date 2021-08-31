package status

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/cliutils"
	"github.com/weaveworks/weave-gitops/pkg/kube"
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
		params := app.StatusParams{}

		params.Name = args[0]
		params.Namespace, _ = cmd.Parent().Parent().Flags().GetString("namespace")

		osysClient, fluxClient, _, logger := cliutils.GetBaseClients()

		kubeClient, _, kubeErr := kube.NewKubeHTTPClient()
		if kubeErr != nil {
			return fmt.Errorf("error initializing kube client: %w", kubeErr)
		}

		appService := app.New(logger, nil, nil, nil, fluxClient, kubeClient, osysClient)

		fluxOutput, lastSuccessReconciliation, err := appService.Status(params)
		if err != nil {
			return fmt.Errorf("failed getting application status: %w", err)
		}

		logger.Printf("Last successful reconciliation: %s\n\n", lastSuccessReconciliation)
		logger.Println(fluxOutput)

		return nil
	},
}
