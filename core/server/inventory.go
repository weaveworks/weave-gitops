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

	helmv2 "github.com/fluxcd/helm-controller/api/v2"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1"
	"github.com/fluxcd/pkg/ssa/utils"
	"github.com/go-logr/logr"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/cli-utils/pkg/object"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"

	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/health"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

// ObjectWithChildren is a recursive data structure containing a tree of Unstructured
// values.
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

	defaultNS := msg.Namespace

	switch msg.Kind {
	case kustomizev1.KustomizationKind:
		inventoryRefs, err = cs.getKustomizationInventory(ctx, client, msg.Name, msg.Namespace)
		if err != nil {
			return nil, fmt.Errorf("failed getting kustomization inventory: %w", err)
		}
	case helmv2.HelmReleaseKind:
		hr, err := cs.getHelmRelease(ctx, client, msg.Name, msg.Namespace)
		if err != nil {
			return nil, fmt.Errorf("failed getting Helm Release for inventory: %w", err)
		}
		inventoryRefs, err = cs.getHelmReleaseInventory(ctx, client, hr)
		if err != nil {
			return nil, fmt.Errorf("failed getting Helm Release inventory: %w", err)
		}
		defaultNS = defaultNSFromHelmRelease(hr)
	default:
		gvk, err := cs.primaryKinds.Lookup(msg.Kind)
		if err != nil {
			return nil, err
		}
		inventoryRefs, err = getFluxLikeInventory(ctx, client, msg.Name, msg.Namespace, *gvk)
		if err != nil {
			return nil, fmt.Errorf("failed getting flux like inventory: %w", err)
		}
	}

	objsWithChildren := getObjectsWithChildren(ctx, defaultNS, inventoryRefs, client, msg.WithChildren, cs.logger)

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

func (cs *coreServer) getHelmRelease(ctx context.Context, k8sClient client.Client, name, namespace string) (*helmv2.HelmRelease, error) {
	release := &helmv2.HelmRelease{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(release), release); err != nil {
		return nil, fmt.Errorf("failed to get helm release: %w", err)
	}

	return release, nil
}

func (cs *coreServer) getHelmReleaseInventory(ctx context.Context, k8sClient client.Client, hr *helmv2.HelmRelease) ([]*unstructured.Unstructured, error) {
	objects, err := getHelmReleaseObjects(ctx, k8sClient, hr)
	if err != nil {
		return nil, fmt.Errorf("failed to get helm release objects: %w", err)
	}

	return objects, nil
}

// Returns the list of resources applied in the helm chart.
func getHelmReleaseObjects(ctx context.Context, k8sClient client.Client, helmRelease *helmv2.HelmRelease) ([]*unstructured.Unstructured, error) {
	secretName := secretNameFromHelmRelease(helmRelease)
	if secretName == nil {
		// skip release if it failed to install
		return nil, nil
	}
	storageSecret := &v1.Secret{}
	if helmRelease.Spec.KubeConfig != nil {
		// helmrelease secret is on another cluster so we cannot inspect it to figure out the inventory and version and other things
		return nil, nil
	}

	if err := k8sClient.Get(ctx, *secretName, storageSecret); err != nil {
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

	magicGzip := []byte{0x1f, 0x8b, 0x08}
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

	objects, err := utils.ReadObjects(strings.NewReader(storage.Manifest))
	if err != nil {
		return nil, fmt.Errorf("failed to read the Helm storage object for HelmRelease '%s': %w", helmRelease.Name, err)
	}

	return objects, nil
}

func unstructuredToInventoryEntry(clusterName string, objWithChildren ObjectWithChildren, clusterUserNamespaces map[string][]v1.Namespace, healthChecker health.HealthChecker) (*pb.InventoryEntry, error) {
	unstructuredObj := *objWithChildren.Object
	if unstructuredObj.GetKind() == "Secret" {
		var err error
		unstructuredObj, err = sanitizeUnstructuredSecret(unstructuredObj)
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

// GetObjectsWithChildren returns objects with their children populated if withChildren is true.
// Objects are retrieved in parallel.
// Children are retrieved recusively, e.g. Deployment -> ReplicaSet -> Pod
func getObjectsWithChildren(ctx context.Context, defaultNS string, objects []*unstructured.Unstructured, k8sClient client.Client, withChildren bool, logger logr.Logger) []*ObjectWithChildren {
	result := []*ObjectWithChildren{}

	var (
		isNamespacedGVK = map[string]bool{}
		resultMu        sync.Mutex
		gvkMu           sync.RWMutex
		wg              sync.WaitGroup
		err             error
	)

	for _, o := range objects {
		wg.Add(1)

		go func(obj unstructured.Unstructured) {
			defer wg.Done()

			// Set the namespace of the object if it is not set.
			if obj.GetNamespace() == "" {
				// Manifest does not contain the namespace of the release.
				// Figure out if the object is namespaced if the namespace is not
				// explicitly set, and configure the namespace accordingly.
				objGVK := obj.GetObjectKind().GroupVersionKind().String()
				gvkMu.RLock()
				namespaced, ok := isNamespacedGVK[objGVK]
				gvkMu.RUnlock()

				if !ok {
					namespaced, err = apiutil.IsObjectNamespaced(&obj, k8sClient.Scheme(), k8sClient.RESTMapper())
					if err != nil {
						logger.Error(err, "failed to determine if resource is namespace scoped", "kind", obj.GetObjectKind().GroupVersionKind().Kind)
						return
					}

					// Cache the result, so we don't have to do this for every object
					gvkMu.Lock()
					isNamespacedGVK[objGVK] = namespaced
					gvkMu.Unlock()
				}

				if namespaced {
					obj.SetNamespace(defaultNS)
				}
			}

			if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(&obj), &obj); err != nil {
				logger.Error(err, "failed to get object", "entry", obj)
				return
			}

			children := []*ObjectWithChildren{}
			if withChildren {
				var err error
				children, err = getChildren(ctx, k8sClient, obj)
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
			defer resultMu.Unlock()
			result = append(result, entry)
		}(*o)
	}

	wg.Wait()

	return result
}

func getChildren(ctx context.Context, k8sClient client.Client, parentObj unstructured.Unstructured) ([]*ObjectWithChildren, error) {
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
		return nil, fmt.Errorf("could not get unstructured object: %w", err)
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
		children, err = getChildren(ctx, k8sClient, c)
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

// ResourceRefToUnstructured converts a flux like resource entry pair of (id, version) into a unstructured object
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

// sanitizeUnstructuredSecret redacts the data field of a Secret object
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

func getFluxLikeInventory(ctx context.Context, k8sClient client.Client, name, namespace string, gvk schema.GroupVersionKind) ([]*unstructured.Unstructured, error) {
	// Create an unstructured object with the desired GVK (GroupVersionKind)
	obj := &unstructured.Unstructured{}
	obj.SetGroupVersionKind(gvk)
	obj.SetName(name)
	obj.SetNamespace(namespace)

	// Get the object from the Kubernetes cluster
	if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
		return nil, fmt.Errorf("failed to get kustomization: %w", err)
	}

	return parseInventoryFromUnstructured(obj)
}

func parseInventoryFromUnstructured(obj *unstructured.Unstructured) ([]*unstructured.Unstructured, error) {
	content := obj.UnstructuredContent()

	// Check if status.inventory is present
	unstructuredInventory, found, err := unstructured.NestedMap(content, "status", "inventory")
	if err != nil {
		return nil, fmt.Errorf("error getting status.inventory from object: %w", err)
	}
	if !found {
		return nil, fmt.Errorf("status.inventory not found in object %s/%s", obj.GetNamespace(), obj.GetName())
	}

	resourceInventory := &kustomizev1.ResourceInventory{}
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(unstructuredInventory, resourceInventory)
	if err != nil {
		return nil, fmt.Errorf("error converting inventory to resource inventory: %w", err)
	}

	objects := []*unstructured.Unstructured{}
	for _, ref := range resourceInventory.Entries {
		u, err := ResourceRefToUnstructured(ref.ID, ref.Version)
		if err != nil {
			return nil, fmt.Errorf("error converting resource ref to unstructured: %w", err)
		}
		objects = append(objects, &u)
	}

	return objects, nil
}

const helmSecretNameFmt = "sh.helm.release.v1.%s.v%v" // #nosec G101

func secretNameFromHelmRelease(helmRelease *helmv2.HelmRelease) *client.ObjectKey {
	if latest := helmRelease.Status.History.Latest(); latest != nil {
		return &client.ObjectKey{
			Name:      fmt.Sprintf(helmSecretNameFmt, latest.Name, latest.Version),
			Namespace: helmRelease.GetStorageNamespace(),
		}
	}

	return nil
}

func defaultNSFromHelmRelease(helmRelease *helmv2.HelmRelease) string {
	if latest := helmRelease.Status.History.Latest(); latest != nil {
		return latest.Namespace
	}

	return ""
}
