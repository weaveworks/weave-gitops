package server

import (
	"context"

	"github.com/weaveworks/weave-gitops/core/clientset"
	"github.com/weaveworks/weave-gitops/core/gitops/kustomize"
	"github.com/weaveworks/weave-gitops/core/gitops/source"
	stypes "github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type fluxServer struct {
	pb.UnimplementedFluxServer

	clientSet clientset.Set
	creator   kustomize.KubeCreator
	fetcher   kustomize.Fetcher

	sourceCreator source.KubeCreator
	sourceFetcher source.KubeFetcher
}

func NewFluxServer(clientSet clientset.Set, creator kustomize.KubeCreator, sourceCreator source.KubeCreator, fetcher kustomize.Fetcher, sourceFetcher source.KubeFetcher) pb.FluxServer {
	return &fluxServer{
		clientSet:     clientSet,
		creator:       creator,
		fetcher:       fetcher,
		sourceCreator: sourceCreator,
		sourceFetcher: sourceFetcher,
	}
}

func (fs *fluxServer) AddKustomization(ctx context.Context, msg *pb.AddKustomizationReq) (*pb.AddKustomizationRes, error) {
	k8sRestClient, err := fs.clientSet.KustomizationClient()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
	}

	kust := stypes.ProtoToKustomization(msg)
	k, err := fs.creator.Create(context.Background(), k8sRestClient, &kust)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to create kustomization: %s", err.Error())
	}

	return &pb.AddKustomizationRes{
		Success:       true,
		Kustomization: stypes.KustomizationToProto(k),
	}, nil
}

func (fs *fluxServer) ListKustomizations(_ context.Context, msg *pb.ListKustomizationsReq) (*pb.ListKustomizationsRes, error) {
	k8sRestClient, err := fs.clientSet.KustomizationClient()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
	}

	kustomizations, err := fs.fetcher.List(context.Background(), k8sRestClient, msg.Namespace, metav1.ListOptions{})
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

func (fs *fluxServer) RemoveKustomization(_ context.Context, msg *pb.RemoveKustomizationReq) (*pb.RemoveKustomizationRes, error) {
	return nil, status.Errorf(codes.Unimplemented, "")
}

func (as *fluxServer) AddGitRepository(ctx context.Context, msg *pb.AddGitRepositoryReq) (*pb.AddGitRepositoryRes, error) {
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

func (as *fluxServer) ListGitRepositories(ctx context.Context, msg *pb.ListGitRepositoryReq) (*pb.ListGitRepositoryRes, error) {
	k8sRestClient, err := as.clientSet.SourceClient()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to make k8s rest client: %s", err.Error())
	}

	repositories, err := as.sourceFetcher.ListGitRepositories(context.Background(), k8sRestClient, msg.Namespace, metav1.ListOptions{})
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
