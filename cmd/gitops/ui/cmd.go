package ui

// Makes it easy to access the UI via port-forwarding

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
)

var Cmd = &cobra.Command{
	Use:           "ui",
	Short:         "Access the gitops UI",
	Long:          `The UI command is currently being rewritten but will help access the UI.`, // FIXME
	RunE:          uiCmd,
	SilenceErrors: true,
	SilenceUsage:  true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},
}

func uiCmd(cmd *cobra.Command, args []string) error {
	namespace, _ := cmd.Parent().Flags().GetString("namespace")

	fmt.Printf(`🚧 This command is temporarily out of order 🚧
To access the UI installed by 'gitops install' you will need to use port-forwarding:
  kubectl port-forward -n %[1]s svc/gitops-server 9001:9001

then navigate to: http://localhost:9001/
`, namespace)

	return nil
}
