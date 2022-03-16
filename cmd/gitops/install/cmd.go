package install

// Provides support for adding a repository of manifests to a gitops cluster. If the cluster does not have
// gitops installed, the user will be prompted to install gitops and then the repository will be added.

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/spf13/cobra"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

type params struct {
	ConfigRepo string
}

var (
	installParams params
)

var Cmd = &cobra.Command{
	Use:   "install",
	Short: "Install or upgrade GitOps",
	Long:  `The install command is currently being rewritten.`,
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

	kubeClient, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("error creating k8s http client: %w", err)
	}

	clusterName, err := kubeClient.GetClusterName(ctx)
	if err != nil {
		return err
	}

	fmt.Printf(`🚧 This command is temporarily out of order 🚧
To install gitops using an already initialised repository, for now use:
  flux install -n %[1]s
  flux create source git -n %[1]s weave-gitops --url=%[2]s --branch=main
  flux create kustomization weave-gitops -n %[1]s --interval=10m --source=GitRepository/weave-gitops --path=./..weave-gitops/clusters/%[3]s
`, namespace, configURL, clusterName)

	return nil
}
