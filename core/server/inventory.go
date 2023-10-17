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
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/health"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/object"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// an object that can store unstructued and its children
type ObjectWithChildren struct {
	Object   *unstructured.Unstructured
	Children []*ObjectWithChildren
}

func (cs *coreServer) GetInventory(ctx context.Context, msg *pb.GetInventoryRequest) (*pb.GetInventoryResponse, error) {
	clustersClient, err := cs.clustersManager.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	client, err := clustersClient.Scoped(msg.ClusterName)
	if err != nil {
		return nil, fmt.Errorf("error getting scoped client for cluster=%s: %w", msg.ClusterName, err)
	}

	var inventoryRefs []*unstructured.Unstructured

	switch msg.Kind {
	case kustomizev1.KustomizationKind:
		inventoryRefs, err = cs.getKustomizationInventory(ctx, client, msg.Name, msg.Namespace)
		if err != nil {
			return nil, fmt.Errorf("failed getting kustomization inventory: %w", err)
		}
	case helmv2.HelmReleaseKind:
		inventoryRefs, err = cs.getHelmReleaseInventory(ctx, client, msg.Name, msg.Namespace)
		if err != nil {
			return nil, fmt.Errorf("failed getting helm Release inventory: %w", err)
		}
	default:
		return nil, fmt.Errorf("unknown kind: %s", msg.Kind)
	}

	objsWithChildren, err := GetObjectsWithChildren(ctx, inventoryRefs, client, msg.WithChildren, cs.logger)
	if err != nil {
		return nil, fmt.Errorf("failed getting objects with children: %w", err)
	}

	entries := []*pb.InventoryEntry{}
	clusterUserNamespaces := cs.clustersManager.GetUserNamespaces(auth.Principal(ctx))
	for _, oc := range objsWithChildren {
		entry, err := unstructuredToInventoryEntry(msg.ClusterName, *oc, clusterUserNamespaces, cs.healthChecker)
		if err != nil {
			return nil, fmt.Errorf("failed converting inventory entry: %w", err)
		}
		entries = append(entries, entry)
	}

	return &pb.GetInventoryResponse{
		Entries: entries,
	}, nil
}

func (cs *coreServer) getKustomizationInventory(ctx context.Context, k8sClient client.Client, name, namespace string) ([]*unstructured.Unstructured, error) {
	ks := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(ks), ks); err != nil {
		return nil, fmt.Errorf("failed to get kustomization: %w", err)
	}

	if ks.Status.Inventory == nil {
		return nil, nil
	}

	if ks.Status.Inventory.Entries == nil {
		return nil, nil
	}

	objects := []*unstructured.Unstructured{}
	for _, ref := range ks.Status.Inventory.Entries {
		obj, err := ResourceRefToUnstructured(ref.ID, ref.Version)
		if err != nil {
			cs.logger.Error(err, "failed converting inventory entry", "entry", ref)
			return nil, err
		}
		objects = append(objects, &obj)
	}

	return objects, nil
}

func (cs *coreServer) getHelmReleaseInventory(ctx context.Context, k8sClient client.Client, name, namespace string) ([]*unstructured.Unstructured, error) {
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

	// FIXME: do we need this?
	for _, obj := range objects {
		if obj.GetNamespace() == "" {
			obj.SetNamespace(namespace)
		}
	}

	return objects, nil
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

	if helmRelease.Spec.KubeConfig != nil {
		// helmrelease secret is on another cluster so we cannot inspect it to figure out the inventory and version and other things
		return nil, nil
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

func unstructuredToInventoryEntry(clusterName string, objWithChildren ObjectWithChildren, clusterUserNamespaces map[string][]v1.Namespace, healthChecker health.HealthChecker) (*pb.InventoryEntry, error) {
	unstructuredObj := *objWithChildren.Object
	if unstructuredObj.GetKind() == "Secret" {
		var err error
		unstructuredObj, err = SanitizeUnstructuredSecret(unstructuredObj)
		if err != nil {
			return nil, fmt.Errorf("error sanitizing secrets: %w", err)
		}
	}
	bytes, err := unstructuredObj.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal unstructured object: %w", err)
	}

	tenant := GetTenant(unstructuredObj.GetNamespace(), clusterName, clusterUserNamespaces)

	health, err := healthChecker.Check(unstructuredObj)
	if err != nil {
		return nil, fmt.Errorf("failed to check health: %w", err)
	}

	children := []*pb.InventoryEntry{}
	for _, c := range objWithChildren.Children {
		child, err := unstructuredToInventoryEntry(clusterName, *c, clusterUserNamespaces, healthChecker)
		if err != nil {
			return nil, fmt.Errorf("failed converting child inventory entry: %w", err)
		}
		children = append(children, child)
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

func GetObjectsWithChildren(ctx context.Context, objects []*unstructured.Unstructured, k8sClient client.Client, withChildren bool, logger logr.Logger) ([]*ObjectWithChildren, error) {
	result := []*ObjectWithChildren{}
	resultMu := sync.Mutex{}

	wg := sync.WaitGroup{}

	for _, o := range objects {
		wg.Add(1)

		go func(obj unstructured.Unstructured) {
			defer wg.Done()

			if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(&obj), &obj); err != nil {
				logger.Error(err, "failed to get object", "entry", obj)
				return
			}

			children := []*ObjectWithChildren{}
			if withChildren {
				var err error
				children, err = GetChildren(ctx, k8sClient, obj)
				if err != nil {
					logger.Error(err, "failed getting children", "entry", obj)
					return
				}
			}

			entry := &ObjectWithChildren{
				Object:   &obj,
				Children: children,
			}

			resultMu.Lock()
			result = append(result, entry)
			resultMu.Unlock()
		}(*o)
	}

	wg.Wait()

	return result, nil
}

func GetChildren(ctx context.Context, k8sClient client.Client, parentObj unstructured.Unstructured) ([]*ObjectWithChildren, error) {
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
		return []*ObjectWithChildren{}, nil
	}

	if err := k8sClient.List(ctx, &listResult, client.InNamespace(parentObj.GetNamespace())); err != nil {
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

	children := []*ObjectWithChildren{}

	for _, c := range unstructuredChildren {
		var err error
		children, err = GetChildren(ctx, k8sClient, c)
		if err != nil {
			return nil, err
		}

		entry := &ObjectWithChildren{
			Object:   &c,
			Children: children,
		}
		children = append(children, entry)
	}

	return children, nil
}

func ResourceRefToUnstructured(id, version string) (unstructured.Unstructured, error) {
	u := unstructured.Unstructured{}

	objMetadata, err := object.ParseObjMetadata(id)
	if err != nil {
		return u, err
	}

	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   objMetadata.GroupKind.Group,
		Kind:    objMetadata.GroupKind.Kind,
		Version: version,
	})
	u.SetName(objMetadata.Name)
	u.SetNamespace(objMetadata.Namespace)

	return u, nil
}

func SanitizeUnstructuredSecret(obj unstructured.Unstructured) (unstructured.Unstructured, error) {
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
