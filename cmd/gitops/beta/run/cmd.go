package run

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/internal/config"
	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/run"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type runCommandFlags struct{}

// TODO: Add flags when adding the actual run command.
var flags runCommandFlags //nolint

func RunCommand(opts *config.Options, client *resty.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Set up an interactive sync between your cluster and your local file system",
		Long:  "This will set up a sync between the cluster in your kubeconfig and the path that you specify on your local filesystem.  If you do not have Flux installed on the cluster then this will add it to the cluster automatically.  This is a requirement so we can sync the files successfully from your local system onto the cluster.  Flux will take care of producing the objects for you.",
		Example: `
# Run the sync on the current working directory
gitops beta run . [flags]

# Run the sync against the dev overlay path
gitops beta run ./deploy/overlays/dev [flags]`,
		SilenceUsage:      true,
		SilenceErrors:     true,
		PreRunE:           betaRunCommandPreRunE(&opts.Endpoint),
		RunE:              betaRunCommandRunE(opts, client),
		DisableAutoGenTag: true,
	}

	return cmd
}

func betaRunCommandPreRunE(endpoint *string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		numArgs := len(args)

		if numArgs == 0 {
			return cmderrors.ErrNoFilePath
		}

		if numArgs > 1 {
			return cmderrors.ErrMultipleFilePaths
		}

		return nil
	}
}

func betaRunCommandRunE(opts *config.Options, client *resty.Client) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		// If there is no cluster in the kube config, return an error.
		cfg, ctx, err := kube.RestConfig()
		if err != nil {
			return cmderrors.ErrNoCluster
		}

		// If there is a valid connection to a cluster when the command is run, connect to the currently selected cluster.
		kubeClient, err := kube.NewKubeHTTPClientWithConfig(cfg, ctx)
		if err != nil {
			return cmderrors.ErrGetKubeClient
		}

		// Check if Flux is installed on the cluster.

		listResult := unstructured.UnstructuredList{}

		listResult.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "Namespace",
		})

		ctx2 := context.Background()

		listOptions := ctrlclient.MatchingLabels{
			coretypes.PartOfLabel: "flux",
		}

		u := unstructured.Unstructured{}

		err = kubeClient.List(ctx2, &listResult, listOptions)
		if err != nil {
			fmt.Println("error getting list:", err)
		} else {
			for _, item := range listResult.Items {
				if item.GetLabels()[flux.VersionLabelKey] != "" {
					u = item
					break
				}
			}
		}

		fluxVersion, err := run.GetFluxVersion(u)
		if err != nil {
			fmt.Println("error getting flux version", err)

			fluxVersion = ""
		}

		fmt.Println("fluxVersion:", fluxVersion)

		// If Flux is not installed on the cluster then the prerequisites will be installed to initiate the reconciliation process.
		// This includes all default controllers to set up a reconciliation loop such as the notification-controller, helm-controller, kustomization-controller, and source-controller.
		// This will also add all relevant CRDs from the controllers above such as Kustomizations, Helm Releases, Git Repository, Helm Repository, Bucket, Alerts, Providers, and Receivers.

		if fluxVersion == "" {
			filePath := args[0]

			err = run.InstallFlux(kubeClient, ctx2, filepath.Join(filePath, run.FluxDirectory), run.ShortManifestFilename)
			if err != nil {
				return fmt.Errorf("flux installation failed: %w", err)
			}
		}

		// If Flux is installed on the cluster then we do not need to install flux.
		// ^^^ this should be easy! :-)

		// Note:
		// This should be able to work with local (kind and k3d) and remote clusters.

		// Out of scope:
		// This just includes installing Flux onto the cluster but does not involve creating the prerequisite reconciliation.
		// This does not include image auto CRDs and controller from Flux. These are not normally installed by default and will be dealt with in a future story.

		return nil
	}
}
