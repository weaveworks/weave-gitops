package install

// Provides support for adding a repository of manifests to a gitops cluster. If the cluster does not have
// gitops installed, the user will be prompted to install gitops and then the repository will be added.

import (
	"context"
	_ "embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/gitops"
)

var (
	gitopsParams gitops.InstallParams
)

var Cmd = &cobra.Command{
	Use:   "install",
	Short: "Install or upgrade GitOps",
	Long: `The install command deploys GitOps in the specified namespace,
adds a cluster entry to the GitOps repo, and persists the GitOps runtime into the
repo. If a previous version is installed, then an in-place upgrade will be performed.`,
	Example: fmt.Sprintf(`  # Install GitOps in the %s namespace
  gitops install --config-repo=ssh://git@github.com/me/mygitopsrepo.git`, wego.DefaultNamespace),
	RunE:          installRunCmd,
	SilenceErrors: true,
	SilenceUsage:  true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},
}

func init() {
	Cmd.Flags().BoolVar(&gitopsParams.DryRun, "dry-run", false, "Outputs all the manifests that would be installed")
	Cmd.Flags().StringVar(&gitopsParams.ConfigRepo, "config-repo", "", "URL of external repository that will hold automation manifests")
	Cmd.Flags().StringVar(&gitopsParams.FluxHTTPSUsername, "flux-https-username", "git", "Optional: only needed if using an https:// repo URL for flux")
	Cmd.Flags().StringVar(&gitopsParams.FluxHTTPSPassword, "flux-https-password", "", "Optional: only needed if using an https:// repo URL for flux")
	cobra.CheckErr(Cmd.MarkFlagRequired("config-repo"))
}

func installRunCmd(cmd *cobra.Command, args []string) error {
	namespace, _ := cmd.Parent().Flags().GetString("namespace")
	osysClient := osys.New()
	log := internal.NewCLILogger(os.Stdout)
	runner := flux.New(osysClient, &runner.CLIRunner{})

	k, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("error creating k8s http client: %w", err)
	}

	gitopsService := gitops.New(log, runner, k)
	gitopsParams.Namespace = namespace
	manifests, err := gitopsService.Install(gitopsParams)
	if err != nil {
		return err
	}

	var gitClient git.Git
	var gitProvider gitproviders.GitProvider

	if !gitopsParams.DryRun {
		factory := services.NewFactory(runner, log)
		providerClient := internal.NewGitProviderClient(osysClient.Stdout(), osysClient.LookupEnv, auth.NewAuthCLIHandler, log)

		gitClient, gitProvider, err = factory.GetGitClients(context.Background(), providerClient, services.GitConfigParams{
			URL:               gitopsParams.ConfigRepo,
			Namespace:         namespace,
			DryRun:            gitopsParams.DryRun,
			FluxHTTPSUsername: gitopsParams.FluxHTTPSUsername,
			FluxHTTPSPassword: gitopsParams.FluxHTTPSPassword,
		})

		if err != nil {
			return fmt.Errorf("error creating git clients: %w", err)
		}
	}

	_, err = gitopsService.StoreManifests(gitClient, gitProvider, gitopsParams, manifests)
	if err != nil {
		return err
	}

	if gitopsParams.DryRun {
		for _, manifest := range manifests {
			fmt.Println(manifest)
		}
	}
	return nil
}
