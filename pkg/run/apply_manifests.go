package run

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/fluxcd/pkg/ssa"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	"sigs.k8s.io/cli-utils/pkg/kstatus/polling"
	"sigs.k8s.io/cli-utils/pkg/object"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

type ResourceManagerForApply interface {
	ApplyAll(ctx context.Context, objects []*unstructured.Unstructured, opts ssa.ApplyOptions) (*ssa.ChangeSet, error)
	WaitForSet(set object.ObjMetadataSet, opts ssa.WaitOptions) error
}

func NewManager(log logger.Logger, ctx context.Context, kubeClient ctrlclient.Client, kubeConfigArgs genericclioptions.RESTClientGetter) (*ssa.ResourceManager, error) {
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

// apply is the equivalent of 'kubectl apply --server-side -f'.
func apply(log logger.Logger, ctx context.Context, manager ResourceManagerForApply, manifestsContent []byte) (string, error) {
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
		cs, err := applySet(log, ctx, manager, stageOne)
		if err != nil {
			log.Failuref("Error applying stage one objects")
			return "", err
		}

		changeSet.Append(cs.Entries)

		if err := waitForSet(log, ctx, manager, changeSet); err != nil {
			log.Failuref("Error waiting for set")
			return "", err
		}
	}

	if len(stageTwo) > 0 {
		cs, err := applySet(log, ctx, manager, stageTwo)
		if err != nil {
			log.Failuref("Error applying stage two objects")
			return "", err
		}

		changeSet.Append(cs.Entries)
	}

	return changeSet.String(), nil
}

func applySet(log logger.Logger, ctx context.Context, manager ResourceManagerForApply, objects []*unstructured.Unstructured) (*ssa.ChangeSet, error) {
	return manager.ApplyAll(ctx, objects, ssa.DefaultApplyOptions())
}

func waitForSet(log logger.Logger, ctx context.Context, manager ResourceManagerForApply, changeSet *ssa.ChangeSet) error {
	return manager.WaitForSet(changeSet.ToObjMetadataSet(), ssa.WaitOptions{Interval: 2 * time.Second, Timeout: time.Minute})
}
