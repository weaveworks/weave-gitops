package gitops

// Provides support for adding a repository of manifests to a wego cluster. If the cluster does not have
// wego installed, the user will be prompted to install wego and then the repository will be added.

import (
	"fmt"
	"os"

	_ "embed"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/wego/version"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
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
  wego gitops install`,
	RunE:          installRunCmd,
	SilenceErrors: true,
	SilenceUsage:  true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},
}

var uinstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Wego",
	Long:  `The uninstall command removes Wego components from the cluster.`,
	Example: `  # Uninstall wego in the wego-system namespace
  wego uninstall`,
	RunE:          uninstallRunCmd,
	SilenceErrors: true,
	SilenceUsage:  true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},
}

func init() {
	Cmd.PersistentFlags().StringVarP(&gitopsParams.Namespace, "namespace", "n", "wego-system", "the namespace scope for this operation")
	Cmd.PersistentFlags().BoolVar(&gitopsParams.DryRun, "dry-run", false, "outputs all the manifests that would be installed")

	Cmd.AddCommand(installCmd)
	Cmd.AddCommand(uinstallCmd)
}

func installRunCmd(cmd *cobra.Command, args []string) error {
	cliRunner := &runner.CLIRunner{}
	osysClient := osys.New()
	fluxClient := flux.New(osysClient, cliRunner)
	kubeClient, err := kube.NewKubeHTTPClient()
	if err != nil {
		return err
	}

	gitopsService := gitops.New(logger.NewCLILogger(os.Stdout), fluxClient, kubeClient)

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

func uninstallRunCmd(cmd *cobra.Command, args []string) error {
	cliRunner := &runner.CLIRunner{}
	osysClient := osys.New()
	fluxClient := flux.New(osysClient, cliRunner)
	kubeClient, err := kube.NewKubeHTTPClient()
	if err != nil {
		return err
	}

	gitopsService := gitops.New(logger.NewCLILogger(os.Stdout), fluxClient, kubeClient)

	return gitopsService.Uninstall(gitops.UinstallParams{
		Namespace: gitopsParams.Namespace,
		DryRun:    gitopsParams.DryRun,
	})
}
