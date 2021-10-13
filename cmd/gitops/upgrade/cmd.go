package upgrade

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/upgrade"
)

var upgradeCmdFlags upgrade.UpgradeValues

var Cmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade to Weave GitOps Enterprise",
	Example: `  # Install GitOps in the wego-system namespace
  gitops upgrade`,
	RunE:          upgradeCmdRunE(),
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.RepoOrgAndName, "pr-repo", "", "The repository to open a pull request against. E.g: acme/my-config-repo (default: git current working directory)")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.Remote, "pr-remote", "origin", "The remote to push the branch to")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.BaseBranch, "pr-base", "main", "The base branch to open the pull request against")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.HeadBranch, "pr-branch", "tier-upgrade-enterprise", "The branch to create the pull request from")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.CommitMessage, "pr-commit-message", "Upgrade to WGE", "The commit message")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.GitRepository, "git-repository", "", "The namespace and name of the GitRepository object governing the flux repo (default: git current working directory)")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.ConfigMap, "config-map", "", "The name of the ConfigMap which contains values for this profile.")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.Out, "out", "", " Optional location to create the profile installation folder in. This should be relative to the current working directory. (default: current)")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.ProfileBranch, "profile-branch", "main", "The branch to use on the repository in which the profile is.")
	Cmd.PersistentFlags().BoolVar(&upgradeCmdFlags.DryRun, "dry-run", false, "Output the generated profile without creating a pull request")
}

func upgradeCmdRunE() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		namespace, err := cmd.Parent().Flags().GetString("namespace")
		if err != nil {
			return fmt.Errorf("couldn't read namespace flag: %v", err)
		}

		// FIXME: maybe a better way to do this?
		upgradeCmdFlags.Namespace = namespace

		return upgrade.Upgrade(
			context.Background(),
			upgradeCmdFlags,
			os.Stdout,
		)
	}
}
