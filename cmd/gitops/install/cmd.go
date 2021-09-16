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
	Namespace string
	DryRun    bool
}

var (
	gitopsParams params
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
	Cmd.PersistentFlags().StringVarP(&gitopsParams.Namespace, "namespace", "n", "wego-system", "the namespace scope for this operation")
	Cmd.PersistentFlags().BoolVar(&gitopsParams.DryRun, "dry-run", false, "outputs all the manifests that would be installed")
}

func installRunCmd(cmd *cobra.Command, args []string) error {
	_, fluxClient, kubeClient, logger, clientErr := apputils.GetBaseClients()
	if clientErr != nil {
		return clientErr
	}

	gitopsService := gitops.New(logger, fluxClient, kubeClient)

	manifests, err := gitopsService.Install(gitops.InstallParams{
		Namespace: gitopsParams.Namespace,
		DryRun:    gitopsParams.DryRun,
	})
	if err != nil {
		return err
	}

	if gitopsParams.DryRun {
		fmt.Println(string(manifests))
	}

	return nil
}
