package server

import (
	"context"
	"fmt"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/kustomize/kstatus/status"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"
	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
)

var (
	KustomizeNameKey      = fmt.Sprintf("%s/name", kustomizev1.GroupVersion.Group)
	KustomizeNamespaceKey = fmt.Sprintf("%s/namespace", kustomizev1.GroupVersion.Group)
	HelmNameKey           = fmt.Sprintf("%s/name", helmv2.GroupVersion.Group)
	HelmNamespaceKey      = fmt.Sprintf("%s/namespace", helmv2.GroupVersion.Group)
)

func (as *coreServer) ListFluxRuntimeObjects(ctx context.Context, msg *pb.ListFluxRuntimeObjectsRequest) (*pb.ListFluxRuntimeObjectsResponse, error) {
	k8s, err := as.k8s.Client(ctx)
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
		list := unstructured.UnstructuredList{}

		list.SetGroupVersionKind(schema.GroupVersionKind{
			Group:   gvk.Group,
			Kind:    gvk.Kind,
			Version: gvk.Version,
		})

		if err := k8s.List(ctx, &list, opts, client.InNamespace(msg.Namespace)); err != nil {
			return nil, fmt.Errorf("could not get unstructured list: %s", err)
		}

		result = append(result, list.Items...)
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
			Name:      obj.GetName(),
			Namespace: obj.GetNamespace(),
			Status:    res.Status.String(),
			Uid:       string(obj.GetUID()),
		})
	}

	return &pb.GetReconciledObjectsResponse{Objects: objects}, nil
}
