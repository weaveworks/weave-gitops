package install

// Provides support for adding a repository of manifests to a gitops cluster. If the cluster does not have
// gitops installed, the user will be prompted to install gitops and then the repository will be added.

import (
	"context"
	_ "embed"
	"errors"
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
	"github.com/weaveworks/weave-gitops/pkg/models"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/pkg/services/applier"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/automation"
	"github.com/weaveworks/weave-gitops/pkg/services/gitopswriter"
	"github.com/weaveworks/weave-gitops/pkg/services/gitrepo"
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

const LabelPartOf = "app.kubernetes.io/part-of"

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
	log := internal.NewCLILogger(os.Stdout)
	flux := flux.New(osysClient, &runner.CLIRunner{})

	k, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("error creating k8s http client: %w", err)
	}

	status := k.GetClusterStatus(ctx)

	clusterName, err := k.GetClusterName(ctx)
	if err != nil {
		return err
	}

	switch status {
	case kube.FluxInstalled:
		return errors.New("Weave GitOps does not yet support installation onto a cluster that is using Flux.\nPlease uninstall flux before proceeding:\n  $ flux uninstall")
	case kube.Unknown:
		return fmt.Errorf("Weave GitOps cannot talk to the cluster %s", clusterName)
	}

	clusterApplier := applier.NewClusterApplier(k)

	var gitClient git.Git

	var gitProvider gitproviders.GitProvider

	factory := services.NewFactory(flux, log)

	if err != nil {
		return fmt.Errorf("failed getting kube service: %w", err)
	}

	if installParams.DryRun {
		gitProvider, err = gitproviders.NewDryRun()
		if err != nil {
			return fmt.Errorf("error creating git provider for dry run: %w", err)
		}
	} else {
		_, err = flux.Install(namespace, false)
		if err != nil {
			return err
		}

		providerClient := internal.NewGitProviderClient(osysClient.Stdout(), osysClient.LookupEnv, auth.NewAuthCLIHandler, log)

		gitClient, gitProvider, err = factory.GetGitClients(context.Background(), providerClient, services.GitConfigParams{
			URL:       installParams.ConfigRepo,
			Namespace: namespace,
			DryRun:    installParams.DryRun,
		})

		if err != nil {
			return fmt.Errorf("error creating git clients: %w", err)
		}
	}

	cluster := models.Cluster{Name: clusterName}
	repoWriter := gitrepo.NewRepoWriter(configURL, gitProvider, gitClient, log)
	automationGen := automation.NewAutomationGenerator(gitProvider, flux, log)
	gitOpsDirWriter := gitopswriter.NewGitOpsDirectoryWriter(automationGen, repoWriter, osysClient, log)

	clusterAutomation, err := automationGen.GenerateClusterAutomation(ctx, cluster, configURL, namespace)
	if err != nil {
		return err
	}

	wegoConfigManifest, err := clusterAutomation.GenerateWegoConfigManifest(clusterName, namespace, namespace)
	if err != nil {
		return fmt.Errorf("failed generating wego config manifest: %w", err)
	}

	manifests := append(clusterAutomation.Manifests(), wegoConfigManifest)

	if installParams.DryRun {
		for _, manifest := range manifests {
			log.Println(string(manifest.Content))
		}

		return nil
	}

	err = clusterApplier.ApplyManifests(ctx, cluster, namespace, append(clusterAutomation.BootstrapManifests(), wegoConfigManifest))
	if err != nil {
		return fmt.Errorf("failed applying manifest: %w", err)
	}

	err = gitOpsDirWriter.AssociateCluster(ctx, cluster, configURL, namespace, namespace, installParams.AutoMerge)
	if err != nil {
		return fmt.Errorf("failed associating cluster: %w", err)
	}

	return nil
}
