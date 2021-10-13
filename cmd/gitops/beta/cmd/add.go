/*
Copyright © 2021 Weaveworks <support@weave.works>
This file is part of the Weave GitOps CLI.
*/
package cmd

import (
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/beta/cmd/add"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "This is for adding content to the GitOps repo",
}

func init() {
	addCmd.AddCommand(add.AppCmd)
	Cmd.AddCommand(addCmd)
}
