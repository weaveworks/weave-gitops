package root

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/services/check"

	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Validates flux compatibility",
	Example: `
# Validate flux and kubernetes compatibility
gitops check
`,
	RunE: checkCmdRunE,
}

func checkCmdRunE(_ *cobra.Command, _ []string) error {
	output, err := check.Pre()
	if err != nil {
		return err
	}

	fmt.Println(output)

	return nil
}
