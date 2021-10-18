package uninstall

// Provides support for adding a repository of manifests to a wego cluster. If the cluster does not have
// wego installed, the user will be prompted to install wego and then the repository will be added.

import (
	"fmt"

	"github.com/spf13/cobra"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/pkg/apputils"
	"github.com/weaveworks/weave-gitops/pkg/services/gitops"
)

type params struct {
	DryRun bool
}

var (
	uninstallParams params
)

var Cmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall GitOps",
	Long:  `The uninstall command removes GitOps components from the cluster.`,
	Example: fmt.Sprintf(`  # Uninstall GitOps from the %s namespace
  gitops uninstall`, wego.DefaultNamespace),
	RunE:          uninstallRunCmd,
	SilenceErrors: true,
	SilenceUsage:  true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},
}

func init() {
	Cmd.Flags().BoolVar(&uninstallParams.DryRun, "dry-run", false, "outputs all the manifests that would be uninstalled")
}

func uninstallRunCmd(cmd *cobra.Command, args []string) error {
	namespace, _ := cmd.Parent().Flags().GetString("namespace")

	clients, err := apputils.GetBaseClients()
	if err != nil {
		return err
	}

	gitopsService := gitops.New(clients.Logger, clients.Flux, clients.Kube)

	return gitopsService.Uninstall(gitops.UninstallParams{
		Namespace: namespace,
		DryRun:    uninstallParams.DryRun,
	})
}
