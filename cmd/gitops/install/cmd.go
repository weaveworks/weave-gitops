package install

// Provides support for adding a repository of manifests to a gitops cluster. If the cluster does not have
// gitops installed, the user will be prompted to install gitops and then the repository will be added.

import (
	_ "embed"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
)

var Cmd = &cobra.Command{
	Use:           "install",
	Short:         "Install or upgrade GitOps",
	Long:          `The install command deploys GitOps in the specified namespace (though right now it just tells you how to)`,
	RunE:          installRunCmd,
	SilenceErrors: true,
	SilenceUsage:  true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},
}

func installRunCmd(cmd *cobra.Command, args []string) error {
	println(`To install weave gitops, run:
    flux install # If you haven't already
    flux create source git weave-gitops --url=https://github.com/weaveworks/weave-gitops/ --branch=v2
    flux create helmrelease weave-gitops --interval=10m --source=GitRepository/weave-gitops --chart=./charts/weave-gitops --target-namespace=flux-system`)

	return nil
}
