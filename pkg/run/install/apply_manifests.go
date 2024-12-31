package install

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/fluxcd/cli-utils/pkg/object"
	"github.com/fluxcd/pkg/ssa"
	"github.com/fluxcd/pkg/ssa/normalize"
	"github.com/fluxcd/pkg/ssa/utils"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/pkg/logger"
)

type ResourceManagerForApply interface {
	ApplyAll(ctx context.Context, objects []*unstructured.Unstructured, opts ssa.ApplyOptions) (*ssa.ChangeSet, error)
	WaitForSet(set object.ObjMetadataSet, opts ssa.WaitOptions) error
}

func NewManager(ctx context.Context, log logger.Logger, kubeClient ctrlclient.Client, kubeConfigArgs genericclioptions.RESTClientGetter) (*ssa.ResourceManager, error) {
	return ssa.NewResourceManager(kubeClient, nil, ssa.Owner{
		Field: "flux",
		Group: "fluxcd.io",
	}), nil
}

// apply is the equivalent of 'kubectl apply --server-side -f'.
func apply(ctx context.Context, log logger.Logger, manager ResourceManagerForApply, manifestsContent []byte) (string, error) { //nolint:unparam
	objs, err := utils.ReadObjects(bytes.NewReader(manifestsContent))
	if err != nil {
		log.Failuref("Error reading Kubernetes objects from the manifests")
		return "", err
	}

	if len(objs) == 0 {
		return "", fmt.Errorf("no Kubernetes objects found in the manifests")
	}

	if err := normalize.UnstructuredList(objs); err != nil {
		log.Failuref("Error setting the list of resources to apply")
		return "", err
	}

	changeSet := ssa.NewChangeSet()

	// contains only CRDs and Namespaces
	var stageOne []*unstructured.Unstructured

	// contains all objects except for CRDs and Namespaces
	var stageTwo []*unstructured.Unstructured

	for _, u := range objs {
		if utils.IsClusterDefinition(u) {
			stageOne = append(stageOne, u)
		} else {
			stageTwo = append(stageTwo, u)
		}
	}

	if len(stageOne) > 0 {
		cs, err := applySet(ctx, manager, stageOne)
		if err != nil {
			log.Failuref("Error applying stage one objects")
			return "", err
		}

		changeSet.Append(cs.Entries)

		if err := waitForSet(manager, changeSet); err != nil {
			log.Failuref("Error waiting for set")
			return "", err
		}
	}

	if len(stageTwo) > 0 {
		cs, err := applySet(ctx, manager, stageTwo)
		if err != nil {
			log.Failuref("Error applying stage two objects")
			return "", err
		}

		changeSet.Append(cs.Entries)
	}

	return changeSet.String(), nil
}

func applySet(ctx context.Context, manager ResourceManagerForApply, objects []*unstructured.Unstructured) (*ssa.ChangeSet, error) {
	return manager.ApplyAll(ctx, objects, ssa.DefaultApplyOptions())
}

func waitForSet(manager ResourceManagerForApply, changeSet *ssa.ChangeSet) error {
	return manager.WaitForSet(changeSet.ToObjMetadataSet(), ssa.WaitOptions{Interval: 2 * time.Second, Timeout: time.Minute})
}
