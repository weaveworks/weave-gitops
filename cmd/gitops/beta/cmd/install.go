/*
Copyright © 2021 Weaveworks <support@weave.works>
This file is part of CLI application wego.
*/
package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/pkg/apputils"
	"github.com/weaveworks/weave-gitops/pkg/services/gitops"
)

// installCmd represents the install command
var installCmd = &cobra.Command{
	Use:   "install",
	Short: "Install or upgrade GitOps",
	Long: `The install command deploys GitOps in the specified namespace.
If a previous version is installed, then an in-place upgrade will be performed.`,
	Example: `  # Install GitOps in the wego-system namespace
  gitops install`,
	RunE:          installRunCmd,
	SilenceErrors: true,
	SilenceUsage:  true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},

	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("install called")
	},
}

func init() {
	Cmd.AddCommand(installCmd)
	Cmd.Flags().BoolVar(&installParams.DryRun, "dry-run", false, "outputs all the manifests that would be installed")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// installCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// installCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

type params struct {
	DryRun bool
}

var (
	installParams params
)

func installRunCmd(cmd *cobra.Command, args []string) error {
	namespace, _ := cmd.Parent().Flags().GetString("namespace")

	_, fluxClient, kubeClient, logger, clientErr := apputils.GetBaseClients()
	if clientErr != nil {
		return clientErr
	}

	gitopsService := gitops.New(logger, fluxClient, kubeClient)

	manifests, err := gitopsService.Install(gitops.InstallParams{
		Namespace: namespace,
		DryRun:    installParams.DryRun,
	})
	if err != nil {
		return err
	}

	if installParams.DryRun {
		fmt.Println(string(manifests))
	}

	return nil
}
