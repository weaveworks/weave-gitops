package gitops

// Provides support for adding a repository of manifests to a wego cluster. If the cluster does not have
// wego installed, the user will be prompted to install wego and then the repository will be added.

import (
	"fmt"

	_ "embed"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services/gitops"
)

var params gitops.InstallParams

var Cmd = &cobra.Command{
	Use:   "gitops",
	Short: "Manages your wego installation",
	Long:  `The gitops sub-commands helps you manage your cluster's gitops setup`,
}

var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install or upgrade Wego",
	Long: `The install command deploys Wego in the specified namespace.
If a previous version is installed, then an in-place upgrade will be performed.`,
	Example: `  # Install wego in the wego-system namespace
  wego install`,
	RunE:          runCmd,
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	installCmd.Flags().StringVarP(&params.Namespace, "namespace", "n", "wego-system", "the namespace scope for this operation")
	installCmd.Flags().BoolVar(&params.DryRun, "dry-run", false, "outputs all the manifests that would be installed")

	Cmd.AddCommand(installCmd)
}

func runCmd(cmd *cobra.Command, args []string) error {
	cliRunner := &runner.CLIRunner{}
	fluxClient := flux.New(cliRunner)
	kubeClient := kube.New(cliRunner)

	gitopsService := gitops.New(fluxClient, kubeClient)

	manifests, err := gitopsService.Install(params)
	if err != nil {
		return err
	}

	if params.DryRun {
		fmt.Println(string(manifests))
	}

	return nil
}
