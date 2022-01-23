package server

import (
	"context"

	"github.com/weaveworks/weave-gitops/core/clientset"
	"github.com/weaveworks/weave-gitops/core/gitops/app"
	"github.com/weaveworks/weave-gitops/core/gitops/kustomize"
	"github.com/weaveworks/weave-gitops/core/gitops/source"
	stypes "github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

type appServer struct {
	pb.UnimplementedAppsServer

	clientSet  clientset.Set
	appCreator app.KubeCreator
	appFetcher app.KubeFetcher

	kustCreator kustomize.KubeCreator
	kustFetcher kustomize.Fetcher

	sourceCreator source.KubeCreator
	sourceFetcher source.KubeFetcher
}

func NewAppServer(clientSet clientset.Set, appCreator app.KubeCreator, kustCreator kustomize.KubeCreator, sourceCreator source.KubeCreator, fetcher app.KubeFetcher, kustFetcher kustomize.Fetcher, sourceFetcher source.KubeFetcher) pb.AppsServer {
	return &appServer{
		clientSet:     clientSet,
		appCreator:    appCreator,
		kustCreator:   kustCreator,
		sourceCreator: sourceCreator,
		appFetcher:    fetcher,
		kustFetcher:   kustFetcher,
		sourceFetcher: sourceFetcher,
	}
}

func (as *appServer) AddApp(_ context.Context, msg *pb.AddAppRequest) (*pb.AddAppResponse, error) {
	k8sRestClient, err := as.clientSet.AppClient()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
	}

	app, err := as.appCreator.Create(context.Background(), k8sRestClient, stypes.AppAddProtoToCustomResource(msg))
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

func (as *appServer) GetApp(_ context.Context, msg *pb.GetAppRequest) (*pb.GetAppResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "")
}

func (as *appServer) ListApps(_ context.Context, msg *pb.ListAppRequest) (*pb.ListAppResponse, error) {
	k8sRestClient, err := as.clientSet.AppClient()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
	}

	app, err := as.appFetcher.List(context.Background(), k8sRestClient, msg.Namespace, metav1.ListOptions{})
	if k8serrors.IsUnauthorized(err) {
		return nil, status.Errorf(codes.PermissionDenied, "")
	} else if k8serrors.IsNotFound(err) {
		return nil, status.Errorf(codes.NotFound, "")
	}

	var results []*pb.App
	for _, item := range app.Items {
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

	k8sRestClient, err := as.clientSet.KustomizationClient()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
	}

	kust := stypes.ProtoToKustomization(msg)
	k, err := as.kustCreator.Create(context.Background(), k8sRestClient, &kust)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to create kustomization: %s", err.Error())
	}

	return &pb.AddKustomizationRes{
		Success:       true,
		Kustomization: stypes.KustomizationToProto(k),
	}, nil
}

func (as *appServer) ListKustomizations(ctx context.Context, msg *pb.ListKustomizationsReq) (*pb.ListKustomizationsRes, error) {
	k8sRestClient, err := as.clientSet.KustomizationClient()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
	}

	appNameLabel, err := labels.NewRequirement("app.kubernetes.io/part-of", selection.Equals, []string{msg.AppName})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to make label for list filter")
	}

	kustomizations, err := as.kustFetcher.List(context.Background(), k8sRestClient, msg.Namespace, metav1.ListOptions{
		LabelSelector: appNameLabel.String(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to create new app: %s", err.Error())
	}

	var results []*pb.Kustomization
	for _, kustomization := range kustomizations.Items {
		results = append(results, stypes.KustomizationToProto(&kustomization))
	}

	return &pb.ListKustomizationsRes{
		Kustomizations: results,
	}, nil
}

func (as *appServer) RemoveKustomizations(ctx context.Context, msg *pb.RemoveKustomizationReq) (*pb.RemoveKustomizationRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "")
}

func (as *appServer) AddGitRepository(ctx context.Context, msg *pb.AddGitRepositoryReq) (*pb.AddGitRepositoryRes, error) {
	k8sRestClient, err := as.clientSet.SourceClient()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
	}

	gr, err := as.sourceCreator.CreateGitRepository(context.Background(), k8sRestClient, stypes.ProtoToGitRepository(msg))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to create kustomization: %s", err.Error())
	}

	return &pb.AddGitRepositoryRes{
		Success:       true,
		GitRepository: stypes.GitRepositoryToProto(gr),
	}, nil
}

func (as *appServer) ListGitRepositories(ctx context.Context, msg *pb.ListGitRepositoryReq) (*pb.ListGitRepositoryRes, error) {
	k8sRestClient, err := as.clientSet.SourceClient()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
	}

	appNameLabel, err := labels.NewRequirement("app.kubernetes.io/part-of", selection.Equals, []string{msg.AppName})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to make label for list filter")
	}

	repositories, err := as.sourceFetcher.ListGitRepositories(context.Background(), k8sRestClient, msg.Namespace, metav1.ListOptions{
		LabelSelector: appNameLabel.String(),
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get git repository list: %s", err.Error())
	}

	var results []*pb.GitRepository
	for _, repository := range repositories.Items {
		results = append(results, stypes.GitRepositoryToProto(&repository))
	}

	return &pb.ListGitRepositoryRes{
		GitRepositories: results,
	}, nil
}
