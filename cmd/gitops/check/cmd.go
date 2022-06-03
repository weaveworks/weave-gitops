package check

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/services/check"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "check",
	Short: "Validates flux compatibility",
	Example: `
# Validate flux and kubernetes compatibility
gitops check
`,
	RunE: runCmd,
}

func runCmd(_ *cobra.Command, _ []string) error {
	output, err := check.Pre()
	if err != nil {
		return err
	}

	fmt.Println(output)

	return nil
}
