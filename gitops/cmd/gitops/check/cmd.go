package check

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/gitops/pkg/services/check"

	"github.com/spf13/cobra"
)

var (
	pre bool
)

var Cmd = &cobra.Command{
	Use:   "check",
	Short: "Validates flux compatibility",
	Example: `
# Validate flux and kubernetes compatibility
gitops check --pre
`,
	RunE: runCmd,
}

func init() {
	Cmd.Flags().BoolVarP(&pre, "pre", "p", true, "perform only the pre-installation checks")
}

func runCmd(_ *cobra.Command, _ []string) error {
	output, err := check.Pre()
	if err != nil {
		return err
	}

	fmt.Println(output)

	return nil
}
