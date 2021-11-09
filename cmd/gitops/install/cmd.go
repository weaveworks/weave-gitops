package install

// Provides support for adding a repository of manifests to a gitops cluster. If the cluster does not have
// gitops installed, the user will be prompted to install gitops and then the repository will be added.

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/cmd/gitops/version"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/osys"
	"github.com/weaveworks/weave-gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/services/gitops"
)

type params struct {
	DryRun bool
}

var (
	installParams params
)

var Cmd = &cobra.Command{
	Use:   "install",
	Short: "Install or upgrade GitOps",
	Long: `The install command deploys GitOps in the specified namespace.
If a previous version is installed, then an in-place upgrade will be performed.`,
	Example: fmt.Sprintf(`  # Install GitOps in the %s namespace
  gitops install`, wego.DefaultNamespace),
	RunE:          installRunCmd,
	SilenceErrors: true,
	SilenceUsage:  true,
	PostRun: func(cmd *cobra.Command, args []string) {
		version.CheckVersion(version.CheckpointParamsWithFlags(version.CheckpointParams(), cmd))
	},
}

func init() {
	Cmd.Flags().BoolVar(&installParams.DryRun, "dry-run", false, "Outputs all the manifests that would be installed")
}

func installRunCmd(cmd *cobra.Command, args []string) error {
	namespace, _ := cmd.Parent().Flags().GetString("namespace")

	log := logger.NewCLILogger(os.Stdout)
	flux := flux.New(osys.New(), &runner.CLIRunner{})

	k, _, err := kube.NewKubeHTTPClient()
	if err != nil {
		return fmt.Errorf("error creating k8s http client: %w", err)
	}

	gitopsService := gitops.New(log, flux, k)

	gitopsParams := gitops.InstallParams{
		Namespace: namespace,
		DryRun:    installParams.DryRun,
	}

	manifests, err := gitopsService.Install(gitopsParams)
	if err != nil {
		return err
	}

	_, err = gitopsService.StoreManifests(nil, nil, gitopsParams, manifests)
	if err != nil {
		return err
	}

	if installParams.DryRun {
		for _, manifest := range manifests {
			fmt.Println(string(manifest))
		}
	}

	return nil
}
