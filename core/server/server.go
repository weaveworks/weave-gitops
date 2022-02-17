package server

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	appsv1 "k8s.io/api/apps/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Hydrate(ctx context.Context, mux *runtime.ServeMux, config *rest.Config) error {
	appsServer := NewAppServer(config)
	if err := pb.RegisterCoreHandlerServer(ctx, mux, appsServer); err != nil {
		return fmt.Errorf("could not register new app server: %w", err)
	}

	return nil
}

// Create the scheme once and re-use it on every call.
// This shouldn't need to change between requests(?)
var scheme = kube.CreateScheme()

const temporarilyEmptyAppName = ""

type coreServer struct {
	pb.UnimplementedCoreServer

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

func NewAppServer(cfg *rest.Config) pb.CoreServer {
	return &coreServer{
		k8s: placeholderClientGetter{cfg: cfg},
	}
}

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

func list(ctx context.Context, k8s client.Client, appName, namespace string, list client.ObjectList) error {
	opts := getMatchingLabels(appName)

	err := k8s.List(ctx, list, &opts, client.InNamespace(namespace))

	if err != nil {
		return status.Errorf(codes.Internal, "unable to create new app: %s", err.Error())
	}

	if k8serrors.IsUnauthorized(err) {
		return status.Errorf(codes.PermissionDenied, err.Error())
	} else if k8serrors.IsNotFound(err) {
		return status.Errorf(codes.NotFound, err.Error())
	}

	return nil
}

func doClientError(err error) error {
	return status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
}
