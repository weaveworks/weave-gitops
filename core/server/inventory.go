package server

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/fluxcd/pkg/ssa"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
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

	var entries []*pb.InventoryEntry

	switch msg.Kind {
	case kustomizev1.KustomizationKind:
		entries, err = cs.getKustomizationInventory(ctx, msg.ClusterName, client, msg.Name, msg.Namespace, msg.WithChildren)
		if err != nil {
			return nil, fmt.Errorf("failed getting kustomization inventory: %w", err)
		}
	case helmv2.HelmReleaseKind:
		entries, err = cs.getHelmReleaseInventory(ctx, msg.ClusterName, client, msg.Name, msg.Namespace, msg.WithChildren)
		if err != nil {
			return nil, fmt.Errorf("failed getting helm Release inventory: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown kind: %s", msg.Kind)
	}

	return &pb.GetInventoryResponse{
		Entries: entries,
	}, nil
}

func (cs *coreServer) getKustomizationInventory(ctx context.Context, clusterName string, k8sClient client.Client, name, namespace string, withChildren bool) ([]*pb.InventoryEntry, error) {
	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(kust), kust); err != nil {
		return nil, fmt.Errorf("failed to get kustomization: %w", err)
	}

	if kust.Status.Inventory == nil {
		return nil, nil
	}

	if kust.Status.Inventory.Entries == nil {
		return nil, nil
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

			entry, err := cs.unstructuredToInventoryEntry(ctx, clusterName, k8sClient, obj, namespace, withChildren)
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

func (cs *coreServer) getHelmReleaseInventory(ctx context.Context, clusterName string, k8sClient client.Client, name, namespace string, withChildren bool) ([]*pb.InventoryEntry, error) {
	release := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(release), release); err != nil {
		return nil, fmt.Errorf("failed to get kustomization: %w", err)
	}

	objects, err := getHelmReleaseObjects(ctx, k8sClient, release)
	if err != nil {
		return nil, fmt.Errorf("failed to get helm release objects: %w", err)
	}

	if len(objects) == 0 {
		return []*pb.InventoryEntry{}, nil
	}

	result := []*pb.InventoryEntry{}
	resultMu := sync.Mutex{}

	wg := sync.WaitGroup{}

	for _, o := range objects {
		wg.Add(1)

		go func(obj unstructured.Unstructured) {
			defer wg.Done()

			if obj.GetNamespace() == "" {
				obj.SetNamespace(release.GetReleaseNamespace())
			}

			if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(&obj), &obj); err != nil {
				cs.logger.Error(err, "failed to get object", "entry", obj)
				return
			}

			entry, err := cs.unstructuredToInventoryEntry(ctx, clusterName, k8sClient, obj, namespace, withChildren)
			if err != nil {
				cs.logger.Error(err, "failed converting inventory entry", "entry", obj)
				return
			}

			resultMu.Lock()
			result = append(result, entry)
			resultMu.Unlock()
		}(*o)
	}

	wg.Wait()

	return result, nil
}

// Returns the list of resources applied in the helm chart.
func getHelmReleaseObjects(ctx context.Context, k8sClient client.Client, helmRelease *helmv2.HelmRelease) ([]*unstructured.Unstructured, error) {
	storageNamespace := helmRelease.GetStorageNamespace()

	storageName := helmRelease.GetReleaseName()

	storageVersion := helmRelease.Status.LastReleaseRevision
	if storageVersion < 1 {
		// skip release if it failed to install
		return nil, nil
	}

	storageSecret := &v1.Secret{}
	secretName := fmt.Sprintf("sh.helm.release.v1.%s.v%v", storageName, storageVersion)
	key := client.ObjectKey{
		Name:      secretName,
		Namespace: storageNamespace,
	}

	if err := k8sClient.Get(ctx, key, storageSecret); err != nil {
		return nil, err
	}

	releaseData, releaseFound := storageSecret.Data["release"]
	if !releaseFound {
		return nil, fmt.Errorf("failed to decode the Helm storage object for HelmRelease '%s'", helmRelease.Name)
	}

	byteData, err := base64.StdEncoding.DecodeString(string(releaseData))
	if err != nil {
		return nil, err
	}

	var magicGzip = []byte{0x1f, 0x8b, 0x08}
	if bytes.Equal(byteData[0:3], magicGzip) {
		r, err := gzip.NewReader(bytes.NewReader(byteData))
		if err != nil {
			return nil, err
		}

		defer r.Close()

		uncompressedByteData, err := io.ReadAll(r)
		if err != nil {
			return nil, err
		}

		byteData = uncompressedByteData
	}

	storage := types.HelmReleaseStorage{}
	if err := json.Unmarshal(byteData, &storage); err != nil {
		return nil, fmt.Errorf("failed to decode the Helm storage object for HelmRelease '%s': %w", helmRelease.Name, err)
	}

	objects, err := ssa.ReadObjects(strings.NewReader(storage.Manifest))
	if err != nil {
		return nil, fmt.Errorf("failed to read the Helm storage object for HelmRelease '%s': %w", helmRelease.Name, err)
	}

	return objects, nil
}

func (cs *coreServer) unstructuredToInventoryEntry(ctx context.Context, clusterName string, k8sClient client.Client, unstructuredObj unstructured.Unstructured, ns string, withChildren bool) (*pb.InventoryEntry, error) {
	var err error

	if unstructuredObj.GetKind() == "Secret" {
		unstructuredObj, err = sanitizeUnstructuredSecret(unstructuredObj)
		if err != nil {
			return nil, fmt.Errorf("error sanitizing secrets: %w", err)
		}
	}

	children := []*pb.InventoryEntry{}

	if withChildren {
		children, err = cs.getChildren(ctx, clusterName, k8sClient, unstructuredObj, ns)
		if err != nil {
			return nil, err
		}
	}

	bytes, err := unstructuredObj.MarshalJSON()
	if err != nil {
		return nil, err
	}

	clusterUserNss := cs.clustersManager.GetUserNamespaces(auth.Principal(ctx))
	tenant := GetTenant(unstructuredObj.GetNamespace(), clusterName, clusterUserNss)

	health, err := cs.healthChecker.Check(unstructuredObj)
	if err != nil {
		return nil, err
	}

	entry := &pb.InventoryEntry{
		Payload:     string(bytes),
		Tenant:      tenant,
		ClusterName: clusterName,
		Children:    children,
		Health: &pb.HealthStatus{
			Status:  string(health.Status),
			Message: health.Message,
		},
	}

	return entry, nil
}

func (cs *coreServer) getChildren(ctx context.Context, clusterName string, k8sClient client.Client, parentObj unstructured.Unstructured, ns string) ([]*pb.InventoryEntry, error) {
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
		entry, err := cs.unstructuredToInventoryEntry(ctx, clusterName, k8sClient, c, ns, true)
		if err != nil {
			return nil, err
		}

		children = append(children, entry)
	}

	return children, nil
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

func sanitizeUnstructuredSecret(obj unstructured.Unstructured) (unstructured.Unstructured, error) {
	redactedUnstructured := unstructured.Unstructured{}
	s := &v1.Secret{}

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.UnstructuredContent(), s)
	if err != nil {
		return redactedUnstructured, fmt.Errorf("converting unstructured to helmrelease: %w", err)
	}

	s.Data = map[string][]byte{"redacted": []byte(nil)}

	redactedObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(s)
	if err != nil {
		return redactedUnstructured, fmt.Errorf("converting unstructured to helmrelease: %w", err)
	}

	redactedUnstructured.SetUnstructuredContent(redactedObj)

	return redactedUnstructured, nil
}
