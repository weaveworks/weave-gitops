package resume

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/resume/app"
)

func GetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume",
		Short: "Resume your GitOps automations",
		Example: `
# Resume gitops automation
gitops resume app <app-name>`,
	}

	cmd.AddCommand(app.Cmd)

	return cmd
}
