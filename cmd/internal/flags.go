package internal

import "github.com/spf13/cobra"

func AddPRFlags(cmd *cobra.Command, headBranch, baseBranch, description, message, title *string) {
	cmd.Flags().StringVar(headBranch, "branch", "", "The branch to create the pull request from")
	cmd.Flags().StringVar(message, "commit-message", "", "The commit message to use")
	cmd.Flags().StringVar(title, "title", "", "The title of the pull request")
	cmd.Flags().StringVar(baseBranch, "base", "", "The base branch of the remote repository")
	cmd.Flags().StringVar(description, "description", "", "The description of the pull request")
}

func AddTemplateFlags(cmd *cobra.Command, template *string, parameterValues *[]string) {
	cmd.Flags().StringVar(template, "from-template", "", "Specify the template to create the resource from")
	cmd.Flags().StringSliceVar(parameterValues, "set", []string{}, "Set parameter values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
}
