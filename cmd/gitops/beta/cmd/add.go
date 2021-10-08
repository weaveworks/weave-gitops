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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
