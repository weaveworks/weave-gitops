package uninstall

// Provides support for adding a repository of manifests to a wego cluster. If the cluster does not have
// wego installed, the user will be prompted to install wego and then the repository will be added.

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
)

var Cmd = &cobra.Command{
	Use:           "uninstall",
	Short:         "Uninstall GitOps",
	Long:          `The uninstall command tells you how to remove GitOps components from the cluster.`,
	RunE:          uninstallRunCmd,
	SilenceErrors: true,
	SilenceUsage:  true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},
}

func init() {
}

func uninstallRunCmd(cmd *cobra.Command, args []string) error {
	println(`To uninstall weave gitops, run:
    flux delete helmrelease weave-gitops
    flux delete source git weave-gitops`)

	return nil
}
