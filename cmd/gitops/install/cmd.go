package install

// Provides support for adding a repository of manifests to a gitops cluster. If the cluster does not have
// gitops installed, the user will be prompted to install gitops and then the repository will be added.

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper"

	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/services/gitopswriter"

	corev1 "k8s.io/api/core/v1"

	"github.com/weaveworks/weave-gitops/cmd/internal"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"

	"github.com/weaveworks/weave-gitops/pkg/services/install"

	"github.com/weaveworks/weave-gitops/pkg/kube"

	"github.com/spf13/cobra"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
)

type params struct {
	DryRun     bool
	AutoMerge  bool
	ConfigRepo string
}

var (
	installParams params
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
	Cmd.Flags().BoolVar(&installParams.DryRun, "dry-run", false, "Outputs all the manifests that would be installed")
	Cmd.Flags().BoolVar(&installParams.AutoMerge, "auto-merge", false, "If set, 'gitops install' will automatically update the default branch for the configuration repository")
	Cmd.Flags().StringVar(&installParams.ConfigRepo, "config-repo", "", "URL of external repository that will hold automation manifests")
	cobra.CheckErr(Cmd.MarkFlagRequired("config-repo"))
}

func installRunCmd(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	namespace, _ := cmd.Parent().Flags().GetString("namespace")

	configURL, err := gitproviders.NewRepoURL(installParams.ConfigRepo)
	if err != nil {
		return err
	}

	osysClient := osys.New()
	fluxClient := flux.New(osysClient, &runner.CLIRunner{})

	kubeClient, rawK8sClient, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("error creating k8s http client: %w", err)
	}

	clusterName, err := kubeClient.GetClusterName(ctx)
	if err != nil {
		return err
	}

	log := internal.NewCLILogger(os.Stdout)

	token, err := internal.GetToken(configURL, osysClient.Stdout(), osysClient.LookupEnv, auth.NewAuthCLIHandler, log)
	if err != nil {
		return err
	}

	gitProvider, err := gitproviders.New(gitproviders.Config{
		Provider: configURL.Provider(),
		Token:    token,
		Hostname: configURL.URL().Host,
	}, configURL.Owner(), gitproviders.GetAccountType)
	if err != nil {
		return fmt.Errorf("error creating git provider client: %w", err)
	}

	if installParams.DryRun {
		manifests, err := models.BootstrapManifests(ctx, fluxClient, gitProvider, clusterName, namespace, configURL)
		if err != nil {
			return fmt.Errorf("failed getting gitops manifests: %w", err)
		}

		for _, manifest := range manifests {
			fmt.Println(string(manifest.Content))
		}

		return nil
	}

	//providerClient := internal.NewGitProviderClient(osysClient.Stdout(), osysClient.LookupEnv, auth.NewAuthCLIHandler, log)
	//factory := services.NewFactory(fluxClient, log)

	// We temporarily need this here otherwise GetGitClients is going to fail
	// as it needs the namespace created to apply the secret
	namespaceObj := &corev1.Namespace{}
	namespaceObj.Name = namespace

	if err := rawK8sClient.Create(ctx, namespaceObj); err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed creating namespace %s: %w", namespace, err)
		}
	}

	// This is creating the secret, uploads it and applies it to the cluster
	// This is going to be broken up to reduce complexity
	// and then generates the source yaml of the config repo when using dry-run option
	//gitClient, gitProvider, err := factory.GetGitClients(context.Background(), providerClient, services.GitConfigParams{
	//	URL:       installParams.ConfigRepo,
	//	Namespace: namespace,
	//	DryRun:    installParams.DryRun,
	//})
	//if err != nil {
	//	return fmt.Errorf("failed getting git clients: %w", err)
	//}

	// TODO: remove git provider parameter. It was used to create the deploy key but the deploy key is created in a different place now
	authService, err := auth.NewAuthService(fluxClient, rawK8sClient, gitProvider, log)
	if err != nil {
		return err
	}

	deployKey, err := authService.SetupDeployKey2(ctx, namespace, clusterName, configURL)
	if err != nil {
		return err
	}

	gitClient := git.New(deployKey, wrapper.NewGoGit())

	repoWriter := gitopswriter.NewRepoWriter(log, gitClient, gitProvider)
	installer := install.NewInstaller(fluxClient, kubeClient, gitClient, gitProvider, log, repoWriter)

	if err = installer.Install(namespace, configURL, installParams.AutoMerge); err != nil {
		return fmt.Errorf("failed installing: %w", err)
	}

	return nil
}
