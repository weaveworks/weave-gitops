package server

import (
	"context"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (as *appServer) AddGitRepository(ctx context.Context, msg *pb.AddGitRepositoryReq) (*pb.AddGitRepositoryRes, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	src := types.ProtoToGitRepository(msg)

	if err := k8s.Create(ctx, src); err != nil {
		return nil, status.Errorf(codes.Internal, "creating source for app %q: %s", msg.AppName, err.Error())
	}

	return &pb.AddGitRepositoryRes{
		Success:       true,
		GitRepository: types.GitRepositoryToProto(src),
	}, nil
}

func (as *appServer) ListGitRepositories(ctx context.Context, msg *pb.ListGitRepositoryReq) (*pb.ListGitRepositoryRes, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	list := &sourcev1.GitRepositoryList{}

	if msg.AppName == "" {
		err = k8s.List(ctx, list)
	} else {
		opts := client.MatchingLabels{
			types.PartOfLabel: msg.AppName,
		}
		err = k8s.List(ctx, list, opts)
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get git repository list: %s", err.Error())
	}

	var results []*pb.GitRepository
	for _, repository := range list.Items {
		results = append(results, types.GitRepositoryToProto(&repository))
	}

	return &pb.ListGitRepositoryRes{
		GitRepositories: results,
	}, nil
}

func (as *appServer) AddHelmRepository(ctx context.Context, msg *pb.AddHelmRepositoryReq) (*pb.AddHelmRepositoryRes, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	src := types.ProtoToHelmRepository(msg)

	if err := k8s.Create(ctx, &src); err != nil {
		return nil, status.Errorf(codes.Internal, "creating source for helm repository %q: %s", msg.Name, err.Error())
	}

	return &pb.AddHelmRepositoryRes{
		Success:        true,
		HelmRepository: types.HelmRepositoryToProto(&src),
	}, nil
}

func (as *appServer) ListHelmRepositories(ctx context.Context, msg *pb.ListHelmRepositoryReq) (*pb.ListHelmRepositoryRes, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	list := &sourcev1.HelmRepositoryList{}

	var opts client.MatchingLabels
	if msg.AppName != "" {
		opts = client.MatchingLabels{
			types.PartOfLabel: msg.AppName,
		}
	}

	if err := k8s.List(ctx, list, &opts); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to list helm repositories for app %s: %s", msg.AppName, err.Error())
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get helm repository list: %s", err.Error())
	}

	var results []*pb.HelmRepository
	for _, repository := range list.Items {
		results = append(results, types.HelmRepositoryToProto(&repository))
	}

	return &pb.ListHelmRepositoryRes{
		HelmRepositories: results,
	}, nil
}
