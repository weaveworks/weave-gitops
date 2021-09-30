package upgrade

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/upgrade"
)

type upgradeFlags struct {
	RepoOrgAndName string
	Remote         string
	BaseBranch     string
	HeadBranch     string
	CommitMessage  string
	GitRepository  string
	Version        string
}

var upgradeCmdFlags upgradeFlags

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
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.Version, "version", "latest", "The version to install")
}

func upgradeCmdRunE() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return upgrade.Upgrade(upgrade.UpgradeValues{
			RepoOrgAndName: upgradeCmdFlags.RepoOrgAndName,
			Remote:         upgradeCmdFlags.Remote,
			HeadBranch:     upgradeCmdFlags.HeadBranch,
			BaseBranch:     upgradeCmdFlags.BaseBranch,
			CommitMessage:  upgradeCmdFlags.CommitMessage,
			GitRepository:  upgradeCmdFlags.GitRepository,
			Version:        upgradeCmdFlags.Version,
		}, os.Stdout)
	}
}
