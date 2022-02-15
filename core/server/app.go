package server

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/api/v1alpha2"
	stypes "github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"github.com/weaveworks/weave-gitops/pkg/kube"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	appsv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Create the scheme once and re-use it on every call.
// This shouldn't need to change between requests(?)
var scheme = kube.CreateScheme()

type appServer struct {
	pb.UnimplementedAppsServer

	k8s placeholderClientGetter
}

// This struct is only here to avoid a circular import with the `server` package.
// This is meant to match the ClientGetter interface.
// Since we are in a prototyping phase, it didn't make sense to move and import that code just yet.
type placeholderClientGetter struct {
	cfg *rest.Config
}

func (p placeholderClientGetter) Client(ctx context.Context) (client.Client, error) {
	return client.New(p.cfg, client.Options{
		Scheme: scheme,
	})
}

func NewAppServer(cfg *rest.Config) pb.AppsServer {
	return &appServer{
		k8s: placeholderClientGetter{cfg: cfg},
	}
}

func (as *appServer) AddApp(ctx context.Context, msg *pb.AddAppRequest) (*pb.AddAppResponse, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	app := stypes.AppAddProtoToCustomResource(msg)

	err = k8s.Create(ctx, app)

	if k8serrors.IsUnauthorized(err) {
		return nil, status.Errorf(codes.PermissionDenied, err.Error())
	} else if k8serrors.IsNotFound(err) {
		return nil, status.Errorf(codes.NotFound, err.Error())
	} else if k8serrors.IsConflict(err) {

	} else if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to create new app: %s", err.Error())
	}

	return &pb.AddAppResponse{
		App:     stypes.AppCustomResourceToProto(app),
		Success: true,
	}, nil
}

func (as *appServer) GetApp(ctx context.Context, msg *pb.GetAppRequest) (*pb.GetAppResponse, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	obj := &v1alpha2.Application{}

	if err := k8s.Get(ctx, types.NamespacedName{Name: msg.AppName, Namespace: msg.Namespace}, obj); err != nil {
		return nil, status.Errorf(codes.Internal, "getting app: %s", err.Error())
	}

	return &pb.GetAppResponse{App: stypes.AppCustomResourceToProto(obj)}, nil
}

func (as *appServer) ListApps(ctx context.Context, msg *pb.ListAppRequest) (*pb.ListAppResponse, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	list := &v1alpha2.ApplicationList{}

	err = k8s.List(ctx, list, client.InNamespace(msg.Namespace))
	if k8serrors.IsUnauthorized(err) {
		return nil, status.Errorf(codes.PermissionDenied, "")
	} else if k8serrors.IsNotFound(err) {
		return nil, status.Errorf(codes.NotFound, "")
	}

	var results []*pb.App
	for _, item := range list.Items {
		results = append(results, stypes.AppCustomResourceToProto(&item))
	}

	return &pb.ListAppResponse{
		Apps: results,
	}, nil
}

func (as *appServer) ListFluxRuntimeObjects(ctx context.Context, msg *pb.ListFluxRuntimeObjectsReq) (*pb.ListFluxRuntimeObjectsRes, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	list := &appsv1.DeploymentList{}

	if err := k8s.List(ctx, list, client.InNamespace(msg.Namespace)); err != nil {
		return nil, fmt.Errorf("listing deployments: %w", err)
	}

	result := []*pb.Deployment{}

	for _, d := range list.Items {
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

	return &pb.ListFluxRuntimeObjectsRes{Deployments: result}, nil
}

func (as *appServer) RemoveApp(_ context.Context, msg *pb.RemoveAppRequest) (*pb.RemoveAppResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "")
}

func doClientError(err error) error {
	return status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
}
