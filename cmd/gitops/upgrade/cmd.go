package upgrade

import (
	"os"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/pkg/upgrade"
)

func upgradeCmd(client *resty.Client) *cobra.Command {
	var cmd = &cobra.Command{
		Use:           "upgrade",
		Short:         "Upgrade to WGE",
		Example:       `gitops upgrade`,
		RunE:          upgradeCmdRunE(client),
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.RepoOrgAndName, "pr-repo", "", "The repository to open a pull request against. E.g: acme/my-config-repo (default: git current working directory)")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.Remote, "pr-remote", "origin", "The remote to push the branch to")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.BaseBranch, "pr-base", "main", "The base branch to open the pull request against")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.HeadBranch, "pr-branch", "tier-upgrade-enterprise", "The branch to create the pull request from")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.CommitMessage, "pr-commit-message", "Upgrade to WGE", "The commit message")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.GitRepository, "git-repository", "", "The namespace and name of the GitRepository object governing the flux repo (default: git current working directory)")
	cmd.PersistentFlags().StringVar(&upgradeCmdFlags.Version, "version", "latest", "The version to install")

	return cmd
}

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

func upgradeCmdRunE(client *resty.Client) func(*cobra.Command, []string) error {
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
