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
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

// apply is the equivalent of 'kubectl apply --server-side -f'.
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

		if err := waitForSet(log, ctx, kubeClient, kubeConfigArgs, changeSet); err != nil {
			log.Failuref("Error waiting for set")
			return "", err
		}
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
