package install

// Provides support for adding a repository of manifests to a wego cluster. If the cluster does not have
// wego installed, the user will be prompted to install wego and then the repository will be added.

import (
	"fmt"

	_ "embed"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/cmdimpl"
	"github.com/weaveworks/weave-gitops/pkg/shims"
)

var params cmdimpl.InstallParamSet

var Cmd = &cobra.Command{
	Use:   "install",
	Short: "Install or upgrade Wego",
	Long: `The install command deploys Wego in the specified namespace.
If a previous version is installed, then an in-place upgrade will be performed.`,
	Example: `  # Install wego in the wego-system namespace
  wego install`,
	Run: runCmd,
}

// checkError will print a message to stderr and exit
func checkError(msg string, err error) {
	if err != nil {
		fmt.Fprintf(shims.Stderr(), "%s: %v\n", msg, err)
		exit(1)
	}
}

func exit(code int) {
	shims.Exit(code)
}

func init() {
	Cmd.Flags().StringVarP(&params.Namespace, "namespace", "n", "wego-system", "the namespace scope for this operation")
	Cmd.Flags().BoolVar(&params.DryRun, "dry-run", false, "outputs all the manifests that would be installed")
}

func runCmd(cmd *cobra.Command, args []string) {
	err := cmdimpl.Install(params)
	checkError("failed outputing install manifests", err)
}
