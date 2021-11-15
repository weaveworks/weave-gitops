package upgrade

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/upgrade"
)

var upgradeCmdFlags upgrade.UpgradeValues

var Cmd = &cobra.Command{
	Use:   "upgrade",
	Short: "Upgrade to Weave GitOps Enterprise",
	Example: fmt.Sprintf(`  # Install GitOps in the %s namespace
  gitops upgrade`, wego.DefaultNamespace),
	RunE:          upgradeCmdRunE(),
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.RepoOrgAndName, "repo", "", "The repository to open a pull request against. E.g: acme/my-config-repo (default: git current working directory)")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.Remote, "remote", "origin", "The remote to push the branch to")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.BaseBranch, "base", "main", "The base branch to open the pull request against")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.HeadBranch, "branch", "tier-upgrade-enterprise", "The branch to create the pull request from")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.CommitMessage, "commit-message", "Upgrade to WGE", "The commit message")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.GitRepository, "git-repository", "", "The namespace and name of the GitRepository object governing the flux repo (default: git current working directory)")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.ConfigMap, "config-map", "", "The name of the ConfigMap which contains values for this profile.")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.Out, "out", "", "Optional location to create the profile installation folder in. This should be relative to the current working directory. (default: current)")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.ProfileBranch, "profile-branch", "main", "The branch to use on the repository in which the profile is.")
	Cmd.PersistentFlags().BoolVar(&upgradeCmdFlags.DryRun, "dry-run", false, "Output the generated profile without creating a pull request")
}

func upgradeCmdRunE() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		namespace, err := cmd.Parent().Flags().GetString("namespace")

		if err != nil {
			return fmt.Errorf("couldn't read namespace flag: %v", err)
		}

		// FIXME: maybe a better way to do this?
		upgradeCmdFlags.Namespace = namespace

		log := internal.NewCLILogger(os.Stdout)
		fluxClient := flux.New(osys.New(), &runner.CLIRunner{})
		factory := services.NewFactory(fluxClient, log)

		providerClient := internal.NewGitProviderClient(os.Stdout, os.LookupEnv, auth.NewAuthCLIHandler, log)

		gitClient, gitProvider, err := factory.GetGitClients(ctx, providerClient, services.GitConfigParams{
			URL:       upgradeCmdFlags.RepoURL,
			Namespace: upgradeCmdFlags.Namespace,
			DryRun:    upgradeCmdFlags.DryRun,
		})
		if err != nil {
			return fmt.Errorf("failed to get git clients: %w", err)
		}

		return upgrade.Upgrade(
			ctx,
			gitClient,
			gitProvider,
			upgradeCmdFlags,
			log,
			os.Stdout,
		)
	}
}
