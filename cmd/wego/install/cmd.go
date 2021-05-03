package install

// Provides support for adding a repository of manifests to a wego cluster. If the cluster does not have
// wego installed, the user will be prompted to install wego and then the repository will be added.

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/fluxops"
)

type paramSet struct {
	namespace string
}

var (
	params paramSet
)

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
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		os.Exit(1)
	}
}

func init() {
	Cmd.Flags().StringVar(&params.namespace, "namespace", "wego-system", "the namespace scope for this operation")
}

func runCmd(cmd *cobra.Command, args []string) {
	_, err := fluxops.Install(params.namespace)
	checkError("failed outputing install manifests", err)
}
