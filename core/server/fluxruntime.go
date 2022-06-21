package server

import (
	"context"
	"errors"
	"fmt"

	coretypes "github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kstatus/status"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

const (
	FluxNamespacePartOf = "flux"
)

var (
	KustomizeNameKey      = fmt.Sprintf("%s/name", kustomizev1.GroupVersion.Group)
	KustomizeNamespaceKey = fmt.Sprintf("%s/namespace", kustomizev1.GroupVersion.Group)
	HelmNameKey           = fmt.Sprintf("%s/name", helmv2.GroupVersion.Group)
	HelmNamespaceKey      = fmt.Sprintf("%s/namespace", helmv2.GroupVersion.Group)

	// ErrFluxNamespaceNotFound no flux namespace found
	ErrFluxNamespaceNotFound = errors.New("could not find flux namespace in cluster")
	// ErrListingDeployments no deployments found
	ErrListingDeployments = errors.New("could not list deployments in namespace")
)

func (cs *coreServer) ListFluxRuntimeObjects(ctx context.Context, msg *pb.ListFluxRuntimeObjectsRequest) (*pb.ListFluxRuntimeObjectsResponse, error) {
	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	var results []*pb.Deployment

	respErrors := []*pb.ListError{}

	for clusterName, nss := range cs.clientsFactory.GetClustersNamespaces() {
		fluxNs := filterFluxNamespace(nss)
		if fluxNs == nil {
			respErrors = append(respErrors, &pb.ListError{ClusterName: clusterName, Namespace: "", Message: ErrFluxNamespaceNotFound.Error()})
			continue
		}

		opts := client.MatchingLabels{
			coretypes.PartOfLabel: FluxNamespacePartOf,
		}

		list := &appsv1.DeploymentList{}

		if err := clustersClient.List(ctx, clusterName, list, opts, client.InNamespace(fluxNs.Name)); err != nil {
			respErrors = append(respErrors, &pb.ListError{ClusterName: clusterName, Namespace: fluxNs.Name, Message: fmt.Sprintf("%s, %s", ErrListingDeployments.Error(), err)})
			continue
		}

		for _, d := range list.Items {
			r := &pb.Deployment{
				Name:        d.Name,
				Namespace:   d.Namespace,
				Conditions:  []*pb.Condition{},
				ClusterName: clusterName,
			}

			for _, cond := range d.Status.Conditions {
				r.Conditions = append(r.Conditions, &pb.Condition{
					Message: cond.Message,
					Reason:  cond.Reason,
					Status:  string(cond.Status),
					Type:    string(cond.Type),
				})
			}

			for _, img := range d.Spec.Template.Spec.Containers {
				r.Images = append(r.Images, img.Image)
			}

			results = append(results, r)
		}
	}

	return &pb.ListFluxRuntimeObjectsResponse{Deployments: results, Errors: respErrors}, nil
}

func filterFluxNamespace(nss []v1.Namespace) *v1.Namespace {
	for _, ns := range nss {
		if val, ok := ns.Labels[coretypes.PartOfLabel]; ok && val == FluxNamespacePartOf {
			return &ns
		}
	}

	return nil
}

func (cs *coreServer) GetReconciledObjects(ctx context.Context, msg *pb.GetReconciledObjectsRequest) (*pb.GetReconciledObjectsResponse, error) {
	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	var opts client.MatchingLabels

	switch msg.AutomationKind {
	case pb.FluxObjectKind_KindKustomization:
		opts = client.MatchingLabels{
			KustomizeNameKey: msg.AutomationName,
		}
	case pb.FluxObjectKind_KindHelmRelease:
		opts = client.MatchingLabels{
			HelmNameKey: msg.AutomationName,
		}
	default:
		return nil, fmt.Errorf("unsupported application kind: %s", msg.AutomationKind.String())
	}

	result := []unstructured.Unstructured{}

	checkDup := map[types.UID]bool{}

	for _, gvk := range msg.Kinds {
		listResult := unstructured.UnstructuredList{}

		listResult.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   gvk.Group,
			Kind:    gvk.Kind,
			Version: gvk.Version,
		})

		if err := clustersClient.List(ctx, msg.ClusterName, &listResult, opts); err != nil {
			if k8serrors.IsForbidden(err) {
				// Our service account (or impersonated user) may not have the ability to see the resource in question,
				// in the given namespace. We pretend it doesn't exist and keep looping.
				// We need logging to make this error more visible.
				continue
			}

			return nil, fmt.Errorf("listing unstructured object: %w", err)
		}

		for _, u := range listResult.Items {
			uid := u.GetUID()

			if !checkDup[uid] {
				result = append(result, u)
				checkDup[uid] = true
			}
		}
	}

	objects := []*pb.UnstructuredObject{}

	for _, obj := range result {
		res, err := status.Compute(&obj)
		if err != nil {
			return nil, fmt.Errorf("could not get status for %s: %w", obj.GetName(), err)
		}

		var images []string

		switch obj.GetKind() {
		case "Deployment":
			images = getDeploymentPodContainerImages(obj.Object)
		}

		objects = append(objects, &pb.UnstructuredObject{
			GroupVersionKind: &pb.GroupVersionKind{
				Group:   obj.GetObjectKind().GroupVersionKind().Group,
				Version: obj.GetObjectKind().GroupVersionKind().GroupVersion().Version,
				Kind:    obj.GetKind(),
			},
			Name:        obj.GetName(),
			Namespace:   obj.GetNamespace(),
			Images:      images,
			Status:      res.Status.String(),
			Uid:         string(obj.GetUID()),
			Conditions:  mapUnstructuredConditions(res),
			ClusterName: msg.GetClusterName(),
		})
	}

	return &pb.GetReconciledObjectsResponse{Objects: objects}, nil
}

func (cs *coreServer) GetChildObjects(ctx context.Context, msg *pb.GetChildObjectsRequest) (*pb.GetChildObjectsResponse, error) {
	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	listResult := unstructured.UnstructuredList{}

	listResult.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   msg.GroupVersionKind.Group,
		Version: msg.GroupVersionKind.Version,
		Kind:    msg.GroupVersionKind.Kind,
	})

	if err := clustersClient.List(ctx, msg.ClusterName, &listResult); err != nil {
		return nil, fmt.Errorf("could not get unstructured object: %s", err)
	}

	objects := []*pb.UnstructuredObject{}

ItemsLoop:
	for _, obj := range listResult.Items {
		refs := obj.GetOwnerReferences()
		if len(refs) == 0 {
			// Ignore items without OwnerReference.
			// for example: dev-weave-gitops-test-connection
			continue ItemsLoop
		}

		for _, ref := range refs {
			if ref.UID != types.UID(msg.ParentUid) {
				// Assuming all owner references have the same parent UID,
				// this is not the child we are looking for.
				// Skip the rest of the operations in Items loops.
				continue ItemsLoop
			}
		}

		statusResult, err := status.Compute(&obj)
		if err != nil {
			return nil, fmt.Errorf("could not get status for %s: %w", obj.GetName(), err)
		}

		var images []string

		switch obj.GetKind() {
		case "Pod":
			images = getPodContainerImages(obj.Object)
		case "ReplicaSet":
			replicasFound := status.GetIntField(obj.UnstructuredContent(), ".status.replicas", 0)

			if replicasFound == 0 {
				continue ItemsLoop
			}
			images = getReplicaSetPodContainerImages(obj.Object)
		}

		objects = append(objects, &pb.UnstructuredObject{
			GroupVersionKind: &pb.GroupVersionKind{
				Group:   obj.GetObjectKind().GroupVersionKind().Group,
				Version: obj.GetObjectKind().GroupVersionKind().GroupVersion().Version,
				Kind:    obj.GetKind(),
			},
			Images:      images,
			Name:        obj.GetName(),
			Namespace:   obj.GetNamespace(),
			Status:      statusResult.Status.String(),
			Uid:         string(obj.GetUID()),
			Conditions:  mapUnstructuredConditions(statusResult),
			ClusterName: msg.GetClusterName(),
		})
	}

	return &pb.GetChildObjectsResponse{Objects: objects}, nil
}

func mapUnstructuredConditions(result *status.Result) []*pb.Condition {
	conds := []*pb.Condition{}

	if result.Status == status.CurrentStatus {
		conds = append(conds, &pb.Condition{Type: "Ready", Status: "True", Message: result.Message})
	}

	return conds
}

func getContainerImages(containers []interface{}) []string {
	images := []string{}

	for _, item := range containers {
		container, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		image, ok, _ := unstructured.NestedString(container, "image")
		if ok {
			images = append(images, image)
		}
	}

	return images
}

func getPodContainerImages(obj map[string]interface{}) []string {
	containers, _, _ := unstructured.NestedSlice(obj, "spec", "containers")

	return getContainerImages(containers)
}

func getReplicaSetPodContainerImages(obj map[string]interface{}) []string {
	containers, _, _ := unstructured.NestedSlice(
		obj,
		"spec", "template", "spec", "containers",
	)

	return getContainerImages(containers)
}

func getDeploymentPodContainerImages(obj map[string]interface{}) []string {
	containers, _, _ := unstructured.NestedSlice(
		obj,
		"spec", "template", "spec", "containers",
	)

	return getContainerImages(containers)
}
