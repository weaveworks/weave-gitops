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

	objs, err := cs.getKustomizationInventory(ctx, client, msg.Name, msg.Namespace)
	if err != nil {
		return nil, fmt.Errorf("failed getting kustomization inventory: %w", err)
	}

	entries := []*pb.InventoryEntry{}
	for _, o := range objs {
		bytes, err := o.MarshalJSON()
		if err != nil {
			return nil, err
		}

		entry := &pb.InventoryEntry{
			Payload: string(bytes),
		}

		entries = append(entries, entry)
	}

	return &pb.GetInventoryResponse{
		Entries: entries,
	}, nil
}

func (cs *coreServer) getKustomizationInventory(ctx context.Context, k8sClient client.Client, name string, namespace string) ([]*unstructured.Unstructured, error) {
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
		return []*unstructured.Unstructured{}, nil
	}

	result := []*unstructured.Unstructured{}
	resultMu := sync.Mutex{}

	wg := sync.WaitGroup{}

	for _, e := range kust.Status.Inventory.Entries {
		wg.Add(1)

		go func(entry kustomizev1.ResourceRef) {
			defer wg.Done()

			obj, err := inventoryEntryToUnstructured(entry)
			if err != nil {
				cs.logger.Error(err, "failed converting inventory entry", "entry", entry)
				return
			}

			if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
				cs.logger.Error(err, "failed to get object", "entry", entry)
				return
			}

			resultMu.Lock()
			result = append(result, obj)
			resultMu.Unlock()
		}(e)
	}

	wg.Wait()

	return result, nil
}

func inventoryEntryToUnstructured(entry kustomizev1.ResourceRef) (*unstructured.Unstructured, error) {
	objMetadata, err := object.ParseObjMetadata(entry.ID)
	if err != nil {
		return nil, err
	}

	u := &unstructured.Unstructured{}
	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   objMetadata.GroupKind.Group,
		Kind:    objMetadata.GroupKind.Kind,
		Version: entry.Version,
	})
	u.SetName(objMetadata.Name)
	u.SetNamespace(objMetadata.Namespace)

	return u, nil
}
