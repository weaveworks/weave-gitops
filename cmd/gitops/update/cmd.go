package update

import (
	"github.com/weaveworks/weave-gitops/cmd/gitops/update/profiles"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
)

func UpdateCommand(endpoint *string, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a Weave GitOps resource",
		Example: `
# Update an installed profile to a certain version
gitops update profile --name <profile-name> --version <version> --config-repo <config-repo-url> --cluster <cluster-name> --namespace <ns-name>
`,
	}

	cmd.AddCommand(profiles.UpdateCommand())

	return cmd
}
