package server

import (
	"context"
	"fmt"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/weaveworks/weave-gitops/api/v1alpha2"
	stypes "github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type k8sClientCreator func(ctx context.Context, cfg *rest.Config) (client.Client, error)
type appServer struct {
	pb.UnimplementedAppsServer

	createK8s k8sClientCreator
	restCfg   *rest.Config
	scheme    *runtime.Scheme
}

var scheme = kube.CreateScheme()

func NewAppServer(cfg *rest.Config) pb.AppsServer {
	return &appServer{
		createK8s: func(_ context.Context, override *rest.Config) (client.Client, error) {
			// Future proofing here to ensure we can extract any auth data from context in the future.
			// `rest.Config` is the second arg to support connecting to different clusters in the future.
			restCfg := cfg
			if override != nil {
				restCfg = override
			}

			rawClient, err := client.New(restCfg, client.Options{
				Scheme: scheme,
			})

			if err != nil {
				return nil, fmt.Errorf("kubernetes client initialization failed: %w", err)
			}

			return rawClient, err
		},
	}
}

func (as *appServer) AddApp(ctx context.Context, msg *pb.AddAppRequest) (*pb.AddAppResponse, error) {
	k8s, err := as.createK8s(ctx, nil)
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
	k8s, err := as.createK8s(ctx, nil)
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
	k8s, err := as.createK8s(ctx, nil)
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

func (as *appServer) RemoveApp(_ context.Context, msg *pb.RemoveAppRequest) (*pb.RemoveAppResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "")
}

func (as *appServer) AddKustomization(ctx context.Context, msg *pb.AddKustomizationReq) (*pb.AddKustomizationRes, error) {
	if msg.SourceRef == nil {
		return nil, status.Errorf(codes.InvalidArgument, "sourceRef is required")
	}

	k8s, err := as.createK8s(ctx, nil)
	if err != nil {
		return nil, doClientError(err)
	}

	kust := stypes.ProtoToKustomization(msg)

	if err := k8s.Create(ctx, &kust); err != nil {
		return nil, status.Errorf(codes.Internal, "creating kustomization for app %q: %s", msg.AppName, err.Error())
	}

	return &pb.AddKustomizationRes{
		Success:       true,
		Kustomization: stypes.KustomizationToProto(&kust),
	}, nil
}

func (as *appServer) ListKustomizations(ctx context.Context, msg *pb.ListKustomizationsReq) (*pb.ListKustomizationsRes, error) {
	k8s, err := as.createK8s(ctx, nil)
	if err != nil {
		return nil, doClientError(err)
	}

	list := &kustomizev1.KustomizationList{}

	opts := client.MatchingLabels{
		"app.kubernetes.io/part-of": msg.AppName,
	}

	if err := k8s.List(ctx, list, &opts); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to create new app: %s", err.Error())
	}

	var results []*pb.Kustomization
	for _, kustomization := range list.Items {
		results = append(results, stypes.KustomizationToProto(&kustomization))
	}

	return &pb.ListKustomizationsRes{
		Kustomizations: results,
	}, nil
}

func (as *appServer) RemoveKustomizations(ctx context.Context, msg *pb.RemoveKustomizationReq) (*pb.RemoveKustomizationRes, error) {
	k8s, err := as.createK8s(ctx, nil)
	if err != nil {
		return nil, doClientError(err)
	}

	kust := &kustomizev1.Kustomization{
		ObjectMeta: metav1.ObjectMeta{
			Name:      msg.KustomizationName,
			Namespace: msg.Namespace,
		},
	}
	if err := k8s.Delete(ctx, kust); err != nil {
		return nil, err
	}

	return &pb.RemoveKustomizationRes{Success: true}, nil
}

func (as *appServer) AddGitRepository(ctx context.Context, msg *pb.AddGitRepositoryReq) (*pb.AddGitRepositoryRes, error) {
	k8s, err := as.createK8s(ctx, nil)
	if err != nil {
		return nil, doClientError(err)
	}

	src := stypes.ProtoToGitRepository(msg)

	if err := k8s.Create(ctx, src); err != nil {
		return nil, status.Errorf(codes.Internal, "creating source for app %q: %s", msg.AppName, err.Error())
	}

	return &pb.AddGitRepositoryRes{
		Success:       true,
		GitRepository: stypes.GitRepositoryToProto(src),
	}, nil
}

func (as *appServer) ListGitRepositories(ctx context.Context, msg *pb.ListGitRepositoryReq) (*pb.ListGitRepositoryRes, error) {
	k8s, err := as.createK8s(ctx, nil)
	if err != nil {
		return nil, doClientError(err)
	}

	list := &sourcev1.GitRepositoryList{}

	opts := client.MatchingLabels{
		"app.kubernetes.io/part-of": msg.AppName,
	}

	if err := k8s.List(ctx, list, opts); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get git repository list: %s", err.Error())
	}

	var results []*pb.GitRepository
	for _, repository := range list.Items {
		results = append(results, stypes.GitRepositoryToProto(&repository))
	}

	return &pb.ListGitRepositoryRes{
		GitRepositories: results,
	}, nil
}

func doClientError(err error) error {
	return status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
}
