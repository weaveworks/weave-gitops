package suspend

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/suspend/app"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "suspend",
		Short: "Suspend your GitOps automations",
		Example: `
# Suspend gitops automation
gitops suspend app <app-name>`,
	}

	cmd.AddCommand(app.Cmd)

	return cmd
}
