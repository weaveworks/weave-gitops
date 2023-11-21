package check

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/cmd/gitops/check/oidcconfig"
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/pkg/services/check"

	"github.com/spf13/cobra"
)

func GetCommand(opts *config.Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "Validates flux compatibility",
		Example: `
# Validate flux and kubernetes compatibility
gitops check
`,
		RunE: runCmd,
	}

	cmd.AddCommand(oidcconfig.OIDCConfigCommand(opts))

	return cmd
}

func runCmd(_ *cobra.Command, _ []string) error {
	output, err := check.Pre()
	if err != nil {
		return err
	}

	fmt.Println(output)

	return nil
}
