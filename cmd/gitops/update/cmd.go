package update

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/update/profiles"
)

func UpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update one or many Weave GitOps resources",
		Example: `
# Update a profile
gitops update profile
`,
	}

	cmd.AddCommand(profiles.Cmd)

	return cmd
}
