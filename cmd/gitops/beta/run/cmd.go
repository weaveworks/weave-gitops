package run

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/fluxcd/flux2/pkg/manifestgen/install"
	runclient "github.com/fluxcd/pkg/runtime/client"
	"github.com/fluxcd/pkg/ssa"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/cmd/internal/config"
	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type runCommandFlags struct{}

const (
	fluxDirectory = "flux-system"
	shortFilename = "gotk-components.yaml"
)

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

		fluxVersion, err := getFluxVersion(u)
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

			err = installFlux(kubeClient, ctx2, filepath.Join(filePath, fluxDirectory), shortFilename)
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

func getFluxVersion(obj unstructured.Unstructured) (string, error) {
	labels := obj.GetLabels()
	if labels == nil {
		return "", fmt.Errorf("error getting labels")
	}

	fluxVersion := labels[flux.VersionLabelKey]
	if fluxVersion == "" {
		return "", fmt.Errorf("no flux version found in labels")
	}

	return fluxVersion, nil
}

func installFlux(kubeClient ctrlclient.Client, ctx context.Context, filePath string, shortFilename string) error {
	opts := install.Options{
		BaseURL:      install.MakeDefaultOptions().BaseURL,
		Version:      "v0.31.2",
		Namespace:    "flux-system",
		Components:   []string{"source-controller", "kustomize-controller", "helm-controller", "notification-controller"},
		ManifestFile: "flux-system.yaml",
		Timeout:      5 * time.Second,
	}

	manifest, err := install.Generate(opts, "")
	if err != nil {
		return fmt.Errorf("couldn't generate manifests: %+v", err)
	}

	content := []byte(manifest.Content)

	err = os.MkdirAll(filePath, os.ModePerm)
	if err != nil {
		return fmt.Errorf("couldn't create file %+v", err)
	}

	err = os.WriteFile(filepath.Join(filePath, shortFilename), content, 0666)
	if err != nil {
		return fmt.Errorf("couldn't write flux manifests to file %+v", err)
	}

	var kubeconfigArgs = genericclioptions.NewConfigFlags(false)

	var kubeclientOptions = new(runclient.Options)

	// *kubeconfigArgs.Namespace = rootArgs.defaults.Namespace
	fromEnv := os.Getenv("FLUX_SYSTEM_NAMESPACE")
	if fromEnv != "" {
		kubeconfigArgs.Namespace = &fromEnv
	}

	kubeconfigArgs.APIServer = nil // prevent AddFlags from configuring --server flag
	kubeconfigArgs.Timeout = nil   // prevent AddFlags from configuring --request-timeout flag, we have --timeout instead
	// kubeconfigArgs.AddFlags(rootCmd.PersistentFlags())

	// Since some subcommands use the `-s` flag as a short version for `--silent`, we manually configure the server flag
	// without the `-s` short version. While we're no longer on par with kubectl's flags, we maintain backwards compatibility
	// on the CLI interface.
	apiServer := ""
	kubeconfigArgs.APIServer = &apiServer
	// rootCmd.PersistentFlags().StringVar(kubeconfigArgs.APIServer, "server", *kubeconfigArgs.APIServer, "The address and port of the Kubernetes API server")

	// kubeclientOptions.BindFlags(rootCmd.PersistentFlags())

	// rootCmd.RegisterFlagCompletionFunc("context", contextsCompletionFunc)
	// rootCmd.RegisterFlagCompletionFunc("namespace", resourceNamesCompletionFunc(corev1.SchemeGroupVersion.WithKind("Namespace")))

	manifestPath := filepath.Join(filePath, fluxDirectory)

	applyOutput, err := apply(kubeClient, ctx, kubeconfigArgs, kubeclientOptions, manifestPath, content)
	if err != nil {
		return fmt.Errorf("install failed: %w", err)
	}

	fmt.Println(applyOutput)

	return nil
}

func newManager(kubeClient ctrlclient.Client, rcg genericclioptions.RESTClientGetter) (*ssa.ResourceManager, error) {
	restMapper, err := rcg.ToRESTMapper()
	if err != nil {
		return nil, err
	}

	kubePoller := polling.NewStatusPoller(kubeClient, restMapper, polling.Options{})

	return ssa.NewResourceManager(kubeClient, kubePoller, ssa.Owner{
		Field: "flux",
		Group: "fluxcd.io",
	}), nil
}

func applySet(kubeClient ctrlclient.Client, ctx context.Context, rcg genericclioptions.RESTClientGetter, objects []*unstructured.Unstructured) (*ssa.ChangeSet, error) {
	man, err := newManager(kubeClient, rcg)
	if err != nil {
		return nil, err
	}

	return man.ApplyAll(ctx, objects, ssa.DefaultApplyOptions())
}

func waitForSet(kubeClient ctrlclient.Client, rcg genericclioptions.RESTClientGetter, changeSet *ssa.ChangeSet) error {
	man, err := newManager(kubeClient, rcg)
	if err != nil {
		return err
	}

	return man.WaitForSet(changeSet.ToObjMetadataSet(), ssa.WaitOptions{Interval: 2 * time.Second, Timeout: time.Minute})
}

func apply(kubeClient ctrlclient.Client, ctx context.Context, rcg genericclioptions.RESTClientGetter, opts *runclient.Options, manifestPath string, manifestContent []byte) (string, error) {
	objs, err := ssa.ReadObjects(bytes.NewReader(manifestContent))

	if err != nil {
		return "", err
	}

	if len(objs) == 0 {
		return "", fmt.Errorf("no Kubernetes objects found at: %s", manifestPath)
	}

	if err := ssa.SetNativeKindsDefaults(objs); err != nil {
		return "", err
	}

	changeSet := ssa.NewChangeSet()

	// contains only CRDs and Namespaces
	var stageOne []*unstructured.Unstructured

	// contains all objects except for CRDs and Namespaces
	var stageTwo []*unstructured.Unstructured

	for _, u := range objs {
		if ssa.IsClusterDefinition(u) {
			stageOne = append(stageOne, u)
		} else {
			stageTwo = append(stageTwo, u)
		}
	}

	if len(stageOne) > 0 {
		cs, err := applySet(kubeClient, ctx, rcg, stageOne)
		if err != nil {
			return "", err
		}

		changeSet.Append(cs.Entries)
	}

	if err := waitForSet(kubeClient, rcg, changeSet); err != nil {
		return "", err
	}

	if len(stageTwo) > 0 {
		cs, err := applySet(kubeClient, ctx, rcg, stageTwo)
		if err != nil {
			return "", err
		}

		changeSet.Append(cs.Entries)
	}

	return changeSet.String(), nil
}
