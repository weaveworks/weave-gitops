package remove

// Provides support for removing an application from gitops management.

import (
	"context"
	"fmt"
	"strings"

	"github.com/lithammer/dedent"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/pkg/apputils"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/services/app"
	"github.com/weaveworks/weave-gitops/pkg/utils"
	"k8s.io/apimachinery/pkg/types"
)

var params app.RemoveParams

var Cmd = &cobra.Command{
	Use:   "remove [--private-key <keyfile>] <app name>",
	Short: "Remove an app from a gitops cluster",
	Long: strings.TrimSpace(dedent.Dedent(`
        Removes an application from a gitops cluster so it will no longer be managed via GitOps
    `)),
	Example: `
  # Remove application from gitops control via immediate commit
  gitops app remove podinfo
`,
	Args:          cobra.MinimumNArgs(1),
	RunE:          runCmd,
	SilenceUsage:  true,
	SilenceErrors: true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},
}

func init() {
	Cmd.Flags().StringVar(&params.PrivateKey, "private-key", "", "Private key to access git repository over ssh")
	Cmd.Flags().BoolVar(&params.DryRun, "dry-run", false, "If set, 'gitops app remove' will not make any changes to the system; it will just display the actions that would have been taken")
}

func runCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	params.Name = args[0]
	params.Namespace, _ = cmd.Parent().Flags().GetString("namespace")

	kube, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("failed to create kube client: %w", err)
	}

	appObj, err := kube.GetApplication(ctx, types.NamespacedName{Name: params.Name, Namespace: params.Namespace})
	if err != nil {
		return fmt.Errorf("could not get application: %w", err)
	}

	appService, appError := apputils.GetAppService(ctx, appObj.Spec.URL, appObj.Spec.ConfigURL, appObj.Namespace, appObj.IsHelmRepository())
	if appError != nil {
		return fmt.Errorf("failed to create app service: %w", appError)
	}

	utils.SetCommmitMessage(fmt.Sprintf("gitops app remove %s", params.Name))

	if err := appService.Remove(params); err != nil {
		return errors.Wrapf(err, "failed to remove the app %s", params.Name)
	}

	return nil
}
