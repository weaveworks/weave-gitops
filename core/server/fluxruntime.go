package server

import (
	"context"
	"fmt"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kstatus/status"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
)

var (
	KustomizeNameKey      = fmt.Sprintf("%s/name", kustomizev1.GroupVersion.Group)
	KustomizeNamespaceKey = fmt.Sprintf("%s/namespace", kustomizev1.GroupVersion.Group)
	HelmNameKey           = fmt.Sprintf("%s/name", helmv2.GroupVersion.Group)
	HelmNamespaceKey      = fmt.Sprintf("%s/namespace", helmv2.GroupVersion.Group)
)

func (cs *coreServer) ListFluxRuntimeObjects(ctx context.Context, msg *pb.ListFluxRuntimeObjectsRequest) (*pb.ListFluxRuntimeObjectsResponse, error) {
	k8s, err := cs.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	l := &appsv1.DeploymentList{}

	if err := list(ctx, k8s, temporarilyEmptyAppName, msg.Namespace, l); err != nil {
		return nil, err
	}

	result := []*pb.Deployment{}

	for _, d := range l.Items {
		r := &pb.Deployment{
			Name:       d.Name,
			Namespace:  d.Namespace,
			Conditions: []*pb.Condition{},
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

		result = append(result, r)
	}

	return &pb.ListFluxRuntimeObjectsResponse{Deployments: result}, nil
}

func (cs *coreServer) GetReconciledObjects(ctx context.Context, msg *pb.GetReconciledObjectsRequest) (*pb.GetReconciledObjectsResponse, error) {
	k8s, err := cs.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	var opts client.MatchingLabels

	switch msg.AutomationKind {
	case pb.AutomationKind_KustomizationAutomation:
		opts = client.MatchingLabels{
			KustomizeNameKey:      msg.AutomationName,
			KustomizeNamespaceKey: msg.Namespace,
		}
	case pb.AutomationKind_HelmReleaseAutomation:
		opts = client.MatchingLabels{
			HelmNameKey:      msg.AutomationName,
			HelmNamespaceKey: msg.Namespace,
		}
	default:
		return nil, fmt.Errorf("unsupported application kind: %s", msg.AutomationKind.String())
	}

	result := []unstructured.Unstructured{}

	for _, gvk := range msg.Kinds {
		l := unstructured.UnstructuredList{}

		l.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   gvk.Group,
			Kind:    gvk.Kind,
			Version: gvk.Version,
		})

		if err := k8s.List(ctx, &l, opts, client.InNamespace(msg.Namespace)); err != nil {
			if k8serrors.IsForbidden(err) {
				// Our service account (or impersonated user) may not have the ability to see the resource in question,
				// in the given namespace.
				// We pretend it doesn't exist and keep looping.
				continue
			}

			return nil, fmt.Errorf("listing unstructured object: %w", err)
		}

		result = append(result, l.Items...)
	}

	objects := []*pb.UnstructuredObject{}

	for _, obj := range result {
		res, err := status.Compute(&obj)

		if err != nil {
			return nil, fmt.Errorf("could not get status for %s: %w", obj.GetName(), err)
		}

		objects = append(objects, &pb.UnstructuredObject{
			GroupVersionKind: &pb.GroupVersionKind{
				Group:   obj.GetObjectKind().GroupVersionKind().Group,
				Version: obj.GetObjectKind().GroupVersionKind().GroupVersion().Version,
				Kind:    obj.GetKind(),
			},
			Name:       obj.GetName(),
			Namespace:  obj.GetNamespace(),
			Status:     res.Status.String(),
			Uid:        string(obj.GetUID()),
			Conditions: mapUnstructuredConditions(res),
		})
	}

	return &pb.GetReconciledObjectsResponse{Objects: objects}, nil
}

func (cs *coreServer) GetChildObjects(ctx context.Context, msg *pb.GetChildObjectsRequest) (*pb.GetChildObjectsResponse, error) {
	k8s, err := cs.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	l := unstructured.UnstructuredList{}

	l.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   msg.GroupVersionKind.Group,
		Version: msg.GroupVersionKind.Version,
		Kind:    msg.GroupVersionKind.Kind,
	})

	if err := list(ctx, k8s, "", msg.Namespace, &l); err != nil {
		return nil, fmt.Errorf("could not get unstructured object: %s", err)
	}

	objects := []*pb.UnstructuredObject{}

Items:
	for _, obj := range l.Items {

		refs := obj.GetOwnerReferences()

		for _, ref := range refs {
			if ref.UID != types.UID(msg.ParentUid) {
				// Assuming all owner references have the same parent UID,
				// this is not the child we are looking for.
				// Skip the rest of the operations in Items loops.
				continue Items
			}
		}

		statusResult, err := status.Compute(&obj)
		if err != nil {
			return nil, fmt.Errorf("could not get status for %s: %w", obj.GetName(), err)
		}
		objects = append(objects, &pb.UnstructuredObject{
			GroupVersionKind: &pb.GroupVersionKind{
				Group:   obj.GetObjectKind().GroupVersionKind().Group,
				Version: obj.GetObjectKind().GroupVersionKind().GroupVersion().Version,
				Kind:    obj.GetKind(),
			},
			Name:       obj.GetName(),
			Namespace:  obj.GetNamespace(),
			Status:     statusResult.Status.String(),
			Uid:        string(obj.GetUID()),
			Conditions: mapUnstructuredConditions(statusResult),
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
