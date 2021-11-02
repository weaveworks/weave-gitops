package app

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/apputils"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
)

var Cmd = &cobra.Command{
	Use:           "app",
	Aliases:       []string{"apps"},
	Short:         "Show information about one or all of the applications under gitops control",
	Args:          cobra.MaximumNArgs(1),
	SilenceUsage:  true,
	SilenceErrors: true,
	Example: `
# Get all applications under gitops control
gitops get apps

# Get status of an application under gitops control
gitops get app <app-name>
`,
	RunE: runCmd,
}

func runCmd(cmd *cobra.Command, args []string) error {
	if len(args) == 1 {
		return getApplicationStatus(cmd, args)
	}

	return getApplications(cmd)
}

func getApplicationStatus(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	params := app.StatusParams{}

	params.Name = args[0]
	params.Namespace, _ = cmd.Parent().Flags().GetString("namespace")

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
}

func getApplications(cmd *cobra.Command) error {
	kubeClient, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("error initializing kubernetes client: %w", err)
	}

	ns, err := cmd.Parent().Flags().GetString("namespace")
	if err != nil {
		return err
	}

	apps, err := kubeClient.GetApplications(context.Background(), ns)
	if err != nil {
		return err
	}

	fmt.Println("NAME")

	for _, app := range apps {
		fmt.Println(app.Name)
	}

	return nil
}
