package server

import (
	"context"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	appsv1 "k8s.io/api/apps/v1"
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
