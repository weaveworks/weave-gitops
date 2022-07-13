package run

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"time"

	runclient "github.com/fluxcd/pkg/runtime/client"
	"github.com/fluxcd/pkg/ssa"
	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	"github.com/weaveworks/weave-gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func GetFluxVersion(log logger.Logger, ctx context.Context, kubeClient *kube.KubeHTTP) (string, error) {
	log.Actionf("Getting Flux version ...")

	listResult := unstructured.UnstructuredList{}

	listResult.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Namespace",
	})

	listOptions := ctrlclient.MatchingLabels{
		coretypes.PartOfLabel: "flux",
	}

	u := unstructured.Unstructured{}

	if err := kubeClient.List(ctx, &listResult, listOptions); err != nil {
		log.Failuref("error getting the list of Flux objects")
		return "", err
	} else {
		for _, item := range listResult.Items {
			if item.GetLabels()[flux.VersionLabelKey] != "" {
				u = item
				break
			}
		}
	}

	labels := u.GetLabels()
	if labels == nil {
		return "", fmt.Errorf("error getting Flux labels")
	}

	fluxVersion := labels[flux.VersionLabelKey]
	if fluxVersion == "" {
		return "", fmt.Errorf("no flux version found")
	}

	return fluxVersion, nil
}

func GetKubeConfigArgs() genericclioptions.RESTClientGetter {
	kubeConfigArgs := genericclioptions.NewConfigFlags(false)

	fromEnv := os.Getenv("FLUX_SYSTEM_NAMESPACE")
	if fromEnv != "" {
		kubeConfigArgs.Namespace = &fromEnv
	}

	kubeConfigArgs.APIServer = nil // prevent AddFlags from configuring --server flag
	kubeConfigArgs.Timeout = nil   // prevent AddFlags from configuring --request-timeout flag, we have --timeout instead

	// Since some subcommands use the `-s` flag as a short version for `--silent`, we manually configure the server flag
	// without the `-s` short version. While we're no longer on par with kubectl's flags, we maintain backwards compatibility
	// on the CLI interface.
	apiServer := ""
	kubeConfigArgs.APIServer = &apiServer

	return kubeConfigArgs
}

func GetKubeClientOptions() *runclient.Options {
	kubeClientOpts := new(runclient.Options)

	return kubeClientOpts
}

func GetKubeClient(log logger.Logger, clusterName string, kubeConfigArgs genericclioptions.RESTClientGetter, kubeClientOpts *runclient.Options) (*kube.KubeHTTP, error) {
	cfg, err := kubeConfigArgs.ToRESTConfig()
	if err != nil {
		log.Failuref("Error getting a restconfig: %v", err.Error())
		return nil, err
	}

	// avoid throttling request when some Flux CRDs are not registered
	cfg.QPS = kubeClientOpts.QPS
	cfg.Burst = kubeClientOpts.Burst

	kubeClient, err := kube.NewKubeHTTPClientWithConfig(cfg, clusterName)
	if err != nil {
		log.Failuref("Kubernetes client initialization failed: %v", err.Error())
		return nil, err
	}

	return kubeClient, nil
}

func newManager(log logger.Logger, ctx context.Context, kubeClient ctrlclient.Client, kubeConfigArgs genericclioptions.RESTClientGetter) (*ssa.ResourceManager, error) {
	restMapper, err := kubeConfigArgs.ToRESTMapper()
	if err != nil {
		log.Failuref("Error getting a restmapper")
		return nil, err
	}

	kubePoller := polling.NewStatusPoller(kubeClient, restMapper, polling.Options{})

	return ssa.NewResourceManager(kubeClient, kubePoller, ssa.Owner{
		Field: "flux",
		Group: "fluxcd.io",
	}), nil
}

func applySet(log logger.Logger, ctx context.Context, kubeClient ctrlclient.Client, kubeConfigArgs genericclioptions.RESTClientGetter, objects []*unstructured.Unstructured) (*ssa.ChangeSet, error) {
	man, err := newManager(log, ctx, kubeClient, kubeConfigArgs)
	if err != nil {
		log.Failuref("Error applying set")
		return nil, err
	}

	return man.ApplyAll(ctx, objects, ssa.DefaultApplyOptions())
}

func waitForSet(log logger.Logger, ctx context.Context, kubeClient ctrlclient.Client, rcg genericclioptions.RESTClientGetter, changeSet *ssa.ChangeSet) error {
	man, err := newManager(log, ctx, kubeClient, rcg)
	if err != nil {
		log.Failuref("Error waiting for set")
		return err
	}

	return man.WaitForSet(changeSet.ToObjMetadataSet(), ssa.WaitOptions{Interval: 2 * time.Second, Timeout: time.Minute})
}

func apply(log logger.Logger, ctx context.Context, kubeClient ctrlclient.Client, kubeConfigArgs genericclioptions.RESTClientGetter, manifestsContent []byte) (string, error) {
	objs, err := ssa.ReadObjects(bytes.NewReader(manifestsContent))
	if err != nil {
		log.Failuref("Error reading Kubernetes objects from the manifests")
		return "", err
	}

	if len(objs) == 0 {
		return "", fmt.Errorf("no Kubernetes objects found in the manifests")
	}

	if err := ssa.SetNativeKindsDefaults(objs); err != nil {
		log.Failuref("Error setting native kinds defaults")
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
		cs, err := applySet(log, ctx, kubeClient, kubeConfigArgs, stageOne)
		if err != nil {
			log.Failuref("Error applying stage one objects")
			return "", err
		}

		changeSet.Append(cs.Entries)
	}

	if err := waitForSet(log, ctx, kubeClient, kubeConfigArgs, changeSet); err != nil {
		log.Failuref("Error waiting for set")
		return "", err
	}

	if len(stageTwo) > 0 {
		cs, err := applySet(log, ctx, kubeClient, kubeConfigArgs, stageTwo)
		if err != nil {
			log.Failuref("Error applying stage two objects")
			return "", err
		}

		changeSet.Append(cs.Entries)
	}

	return changeSet.String(), nil
}
