package upgrade

import (
	"context"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/cmd/gitops/logger"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/pkg/upgrade"
)

var upgradeCmdFlags upgrade.UpgradeValues

var example = fmt.Sprintf(`  # Upgrade Weave GitOps in the %s namespace
  gitops upgrade --version 0.0.17 --config-repo https://github.com/my-org/my-management-cluster.git

  # Upgrade Weave GitOps and set the natsURL
  gitops upgrade --version 0.0.17 --set "agentTemplate.natsURL=my-cluster.acme.org:4222" \
    --config-repo https://github.com/my-org/my-management-cluster.git`,
	wego.DefaultNamespace)

var Cmd = &cobra.Command{
	Use:           "upgrade",
	Short:         "Upgrade to Weave GitOps Enterprise",
	Example:       example,
	RunE:          upgradeCmdRunE(),
	SilenceErrors: true,
	SilenceUsage:  true,
}

func init() {
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.ConfigRepo, "config-repo", "", "URL of external repository that will hold automation manifests")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.Version, "version", "", "Version of Weave GitOps Enterprise to be installed")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.BaseBranch, "base", "", "The base branch to open the pull request against")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.HeadBranch, "branch", "tier-upgrade-enterprise", "The branch to create the pull request from")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.ClusterPath, "path", "", "The path within the Git repository containing files for the current cluster")
	Cmd.PersistentFlags().StringVar(&upgradeCmdFlags.CommitMessage, "commit-message", "Upgrade to WGE", "The commit message")
	Cmd.PersistentFlags().StringArrayVar(&upgradeCmdFlags.Values, "set", []string{}, "set profile values on the command line (can specify multiple or separate values with commas: key1=val1,key2=val2)")
	Cmd.PersistentFlags().BoolVar(&upgradeCmdFlags.DryRun, "dry-run", false, "Output the generated profile without creating a pull request")

	cobra.CheckErr(Cmd.MarkPersistentFlagRequired("version"))
}

func upgradeCmdRunE() func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		namespace, err := cmd.Parent().Flags().GetString("namespace")

		if err != nil {
			return fmt.Errorf("couldn't read namespace flag: %v", err)
		}

		kubeClient, err := kube.NewKubeHTTPClient()
		if err != nil {
			return fmt.Errorf("failed to create kube client: %w", err)
		}

		// FIXME: maybe a better way to do this?
		upgradeCmdFlags.Namespace = namespace

		log := logger.NewCLILogger(os.Stdout)
		fluxClient := flux.New(&runner.CLIRunner{})
		factory := services.NewFactory(fluxClient, logger.Logr())

		providerClient := internal.NewGitProviderClient(os.Stdout, os.LookupEnv, log)

		gitClient, gitProvider, err := factory.GetGitClients(ctx, kubeClient, providerClient, services.GitConfigParams{
			ConfigRepo: upgradeCmdFlags.ConfigRepo,
			Namespace:  upgradeCmdFlags.Namespace,
			DryRun:     upgradeCmdFlags.DryRun,
		})
		if err != nil {
			return fmt.Errorf("failed to get git clients: %w", err)
		}

		return upgrade.Upgrade(
			ctx,
			kubeClient,
			gitClient,
			gitProvider,
			upgradeCmdFlags,
			log,
			os.Stdout,
		)
	}
}
