package update

import (
	"github.com/weaveworks/weave-gitops/cmd/gitops/config"
	"github.com/weaveworks/weave-gitops/cmd/gitops/update/profiles"
	"github.com/weaveworks/weave-gitops/pkg/adapters"

	"github.com/spf13/cobra"
)

func UpdateCommand(opts *config.Options, client *adapters.HTTPClient) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a Weave GitOps resource",
		Example: `
	# Update a profile that is installed on a cluster
	gitops update profile --name=podinfo --cluster=prod --config-repo=ssh://git@github.com/owner/config-repo.git  --version=1.0.0
		`,
	}

	cmd.AddCommand(profiles.UpdateCommand(opts, client))

	return cmd
}
