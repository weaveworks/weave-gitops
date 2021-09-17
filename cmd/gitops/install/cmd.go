package install

// Provides support for adding a repository of manifests to a gitops cluster. If the cluster does not have
// gitops installed, the user will be prompted to install gitops and then the repository will be added.

import (
	"fmt"

	_ "embed"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/pkg/apputils"
	"github.com/weaveworks/weave-gitops/pkg/services/gitops"
)

type params struct {
	DryRun bool
}

var (
	installParams params
)

var Cmd = &cobra.Command{
	Use:   "install",
	Short: "Install or upgrade GitOps",
	Long: `The install command deploys GitOps in the specified namespace.
If a previous version is installed, then an in-place upgrade will be performed.`,
	Example: `  # Install GitOps in the wego-system namespace
  gitops install`,
	RunE:          installRunCmd,
	SilenceErrors: true,
	SilenceUsage:  true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},
}

func init() {
	Cmd.Flags().BoolVar(&installParams.DryRun, "dry-run", false, "outputs all the manifests that would be installed")
}

func installRunCmd(cmd *cobra.Command, args []string) error {
	namespace, _ := cmd.Parent().Flags().GetString("namespace")

	_, fluxClient, kubeClient, logger, clientErr := apputils.GetBaseClients()
	if clientErr != nil {
		return clientErr
	}

	gitopsService := gitops.New(logger, fluxClient, kubeClient)

	manifests, err := gitopsService.Install(gitops.InstallParams{
		Namespace: namespace,
		DryRun:    installParams.DryRun,
	})
	if err != nil {
		return err
	}

	if installParams.DryRun {
		fmt.Println(string(manifests))
	}

	return nil
}
