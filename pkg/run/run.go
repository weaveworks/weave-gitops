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
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	FluxDirectory         = "flux-system"
	ShortManifestFilename = "gotk-components.yaml"
)

func GetFluxVersion(obj unstructured.Unstructured) (string, error) {
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

func InstallFlux(kubeClient ctrlclient.Client, ctx context.Context, filePath string, shortFilename string) error {
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

	manifestPath := filepath.Join(filePath, FluxDirectory)

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
