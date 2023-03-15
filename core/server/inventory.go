package server

import (
	"context"
	"fmt"
	"sync"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/object"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cs *coreServer) GetInventory(ctx context.Context, msg *pb.GetInventoryRequest) (*pb.GetInventoryResponse, error) {
	clustersClient, err := cs.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	client, err := clustersClient.Scoped(msg.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting scoped client for cluster=%s: %w", msg.ClusterName, err)
	}

	entries, err := cs.getKustomizationInventory(ctx, client, msg.Name, msg.Namespace, msg.WithChildren)
	if err != nil {
		return nil, fmt.Errorf("failed getting kustomization inventory: %w", err)
	}

	return &pb.GetInventoryResponse{
		Entries: entries,
	}, nil
}

func getChildren(ctx context.Context, k8sClient client.Client, parentObj unstructured.Unstructured, ns string) ([]*pb.InventoryEntry, error) {
	listResult := unstructured.UnstructuredList{}

	switch parentObj.GetObjectKind().GroupVersionKind().Kind {
	case "Deployment", "StatefulSet":
		listResult.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "apps",
			Version: "v1",
			Kind:    "ReplicaSet",
		})
	case "ReplicaSet":
		listResult.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   "",
			Version: "v1",
			Kind:    "Pod",
		})
	default:
		return []*pb.InventoryEntry{}, nil
	}

	if err := k8sClient.List(ctx, &listResult, client.InNamespace(ns)); err != nil {
		return nil, fmt.Errorf("could not get unstructured object: %s", err)
	}

	unstructuredChildren := []unstructured.Unstructured{}

	for _, o := range listResult.Items {
		refs := o.GetOwnerReferences()
		if len(refs) == 0 {
			// Ignore items without OwnerReference.
			// for example: dev-weave-gitops-test-connection
			continue
		}
		for _, ref := range refs {
			if ref.UID == parentObj.GetUID() {
				unstructuredChildren = append(unstructuredChildren, o)
			}
		}
	}

	children := []*pb.InventoryEntry{}

	for _, c := range unstructuredChildren {
		entry, err := unstructuredToInventoryEntry(ctx, k8sClient, c, ns, true)
		if err != nil {
			return nil, err
		}

		children = append(children, entry)
	}

	return children, nil
}

func unstructuredToInventoryEntry(ctx context.Context, k8sClient client.Client, obj unstructured.Unstructured, ns string, withChildren bool) (*pb.InventoryEntry, error) {
	bytes, err := obj.MarshalJSON()
	if err != nil {
		return nil, err
	}

	children := []*pb.InventoryEntry{}

	if withChildren {
		children, err = getChildren(ctx, k8sClient, obj, ns)
		if err != nil {
			return nil, err
		}
	}

	entry := &pb.InventoryEntry{
		Payload:  string(bytes),
		Children: children,
	}

	return entry, nil
}

func (cs *coreServer) getKustomizationInventory(ctx context.Context, k8sClient client.Client, name string, namespace string, withChildren bool) ([]*pb.InventoryEntry, error) {
	kust := &kustomizev1.Kustomization{
		ObjectMeta: v1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(kust), kust); err != nil {
		return nil, fmt.Errorf("failed to get kustomization: %w", err)
	}

	if kust.Status.Inventory.Entries == nil {
		return []*pb.InventoryEntry{}, nil
	}

	result := []*pb.InventoryEntry{}
	resultMu := sync.Mutex{}

	wg := sync.WaitGroup{}

	for _, e := range kust.Status.Inventory.Entries {
		wg.Add(1)

		go func(ref kustomizev1.ResourceRef) {
			defer wg.Done()

			obj, err := resourceRefToUnstructured(ref)
			if err != nil {
				cs.logger.Error(err, "failed converting inventory entry", "entry", ref)
				return
			}

			if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(&obj), &obj); err != nil {
				cs.logger.Error(err, "failed to get object", "entry", ref)
				return
			}

			entry, err := unstructuredToInventoryEntry(ctx, k8sClient, obj, namespace, withChildren)
			if err != nil {
				cs.logger.Error(err, "failed converting inventory entry", "entry", ref)
				return
			}

			resultMu.Lock()
			result = append(result, entry)
			resultMu.Unlock()
		}(e)
	}

	wg.Wait()

	return result, nil
}

func resourceRefToUnstructured(entry kustomizev1.ResourceRef) (unstructured.Unstructured, error) {
	u := unstructured.Unstructured{}

	objMetadata, err := object.ParseObjMetadata(entry.ID)
	if err != nil {
		return u, err
	}

	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   objMetadata.GroupKind.Group,
		Kind:    objMetadata.GroupKind.Kind,
		Version: entry.Version,
	})
	u.SetName(objMetadata.Name)
	u.SetNamespace(objMetadata.Namespace)

	return u, nil
}
