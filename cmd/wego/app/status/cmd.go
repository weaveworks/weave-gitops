package status

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
)

var Cmd = &cobra.Command{
	Use:           "status <app-name>",
	Short:         "Get status of a workload under wego control",
	Args:          cobra.MinimumNArgs(1),
  SilenceUsage:  true,
  SilenceErros:  true,
	Example:       "wego app status podinfo",
	RunE: func(cmd *cobra.Command, args []string) error {
		params := app.StatusParams{}

		params.Name = args[0]
		params.Namespace, _ = cmd.Parent().Parent().Flags().GetString("namespace")

		cliRunner := &runner.CLIRunner{}
		fluxClient := flux.New(cliRunner)
		kubeClient, err := kube.NewKubeHTTPClient()
		if err != nil {
			return fmt.Errorf("error initializing kube client: %w", err)
		}

		gitClient := git.New(nil)
		gitProviders := gitproviders.New()
		logger := logger.New(os.Stdout)

		appService := app.New(logger, gitClient, fluxClient, kubeClient, gitProviders)

		fluxOutput, lastSuccessReconciliation, err := appService.Status(params)
		if err != nil {
			return fmt.Errorf("failed getting application status: %w", err)
		}

		logger.Printf("Last successful reconciliation: %s\n\n", lastSuccessReconciliation)
		logger.Println(fluxOutput)

		return nil
	},
}
