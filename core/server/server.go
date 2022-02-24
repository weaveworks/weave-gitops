package server

import (
	"context"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func Hydrate(ctx context.Context, mux *runtime.ServeMux, config *rest.Config) error {
	appsServer := NewCoreServer(config)
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

func NewCoreServer(cfg *rest.Config) pb.CoreServer {
	return &coreServer{
		k8s: placeholderClientGetter{cfg: cfg},
	}
}

func list(ctx context.Context, k8s client.Client, appName, namespace string, list client.ObjectList) error {
	opts := getMatchingLabels(appName)
	err := k8s.List(ctx, list, &opts, client.InNamespace(namespace))
	err = wrapK8sAPIError("list resource", err)

	return err
}

func get(ctx context.Context, k8s client.Client, name, namespace string, obj client.Object) error {
	err := k8s.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, obj)
	err = wrapK8sAPIError("get resource", err)

	return err
}

func wrapK8sAPIError(msg string, err error) error {
	if k8serrors.IsUnauthorized(err) {
		return status.Errorf(codes.PermissionDenied, err.Error())
	} else if k8serrors.IsNotFound(err) {
		return status.Errorf(codes.NotFound, err.Error())
	} else if err != nil {
		return fmt.Errorf("%s: %w", msg, err)
	}

	return nil
}

func doClientError(err error) error {
	return status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
}
