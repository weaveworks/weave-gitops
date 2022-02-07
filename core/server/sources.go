package server

import (
	"context"
	"fmt"

	helmv2 "github.com/fluxcd/helm-controller/api/v2beta1"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (as *appServer) AddGitRepository(ctx context.Context, msg *pb.AddGitRepositoryReq) (*pb.AddGitRepositoryRes, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	if msg.Reference == nil {
		return nil, status.Error(codes.InvalidArgument, "missing 'reference' field")
	}

	u, err := gitproviders.NewRepoURL(msg.Url)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid url: %s", err.Error())
	}

	msg.Url = u.String()

	src := types.ProtoToGitRepository(msg)

	if err := k8s.Create(ctx, src); err != nil {
		return nil, fmt.Errorf("creating git repo: %w", err)
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

	opts := getMatchingLabels(msg.AppName)

	if err := k8s.List(ctx, list, &opts, client.InNamespace(msg.Namespace)); err != nil {
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

	opts := getMatchingLabels(msg.AppName)

	if err := k8s.List(ctx, list, &opts, client.InNamespace(msg.Namespace)); err != nil {
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

func (as *appServer) AddHelmChart(ctx context.Context, msg *pb.AddHelmChartReq) (*pb.AddHelmChartRes, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	src := types.ProtoToHelmChart(msg)

	if err := k8s.Create(ctx, &src); err != nil {
		return nil, status.Errorf(codes.Internal, "creating source for helm chart %q: %s", msg.HelmChart.Name, err.Error())
	}

	return &pb.AddHelmChartRes{
		Success:   true,
		HelmChart: types.HelmChartToProto(&src),
	}, nil
}

func (as *appServer) ListHelmCharts(ctx context.Context, msg *pb.ListHelmChartReq) (*pb.ListHelmChartRes, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	list := &sourcev1.HelmChartList{}

	opts := getMatchingLabels(msg.AppName)

	if err := k8s.List(ctx, list, &opts, client.InNamespace(msg.Namespace)); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to list helm charts for app %s: %s", msg.AppName, err.Error())
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get helm repository list: %s", err.Error())
	}

	var results []*pb.HelmChart
	for _, repository := range list.Items {
		results = append(results, types.HelmChartToProto(&repository))
	}

	return &pb.ListHelmChartRes{
		HelmCharts: results,
	}, nil
}

func (as *appServer) AddBucket(ctx context.Context, msg *pb.AddBucketReq) (*pb.AddBucketRes, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	src := types.ProtoToBucket(msg)

	if err := k8s.Create(ctx, &src); err != nil {
		return nil, status.Errorf(codes.Internal, "creating source for bucket %q: %s", msg.Bucket.Name, err.Error())
	}

	return &pb.AddBucketRes{
		Success: true,
		Bucket:  types.BucketToProto(&src),
	}, nil
}

func (as *appServer) ListBuckets(ctx context.Context, msg *pb.ListBucketReq) (*pb.ListBucketRes, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	list := &sourcev1.BucketList{}

	opts := getMatchingLabels(msg.AppName)

	if err := k8s.List(ctx, list, &opts, client.InNamespace(msg.Namespace)); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to list helm charts for app %s: %s", msg.AppName, err.Error())
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get helm repository list: %s", err.Error())
	}

	var results []*pb.Bucket
	for _, repository := range list.Items {
		results = append(results, types.BucketToProto(&repository))
	}

	return &pb.ListBucketRes{
		Buckets: results,
	}, nil
}

func (as *appServer) AddHelmRelease(ctx context.Context, msg *pb.AddHelmReleaseReq) (*pb.AddHelmReleaseRes, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	src := types.ProtoToHelmRelease(msg)

	if err := k8s.Create(ctx, &src); err != nil {
		return nil, status.Errorf(codes.Internal, "creating helm release %q: %s", msg.HelmRelease.Name, err.Error())
	}

	return &pb.AddHelmReleaseRes{
		Success:     true,
		HelmRelease: types.HelmReleaseToProto(&src),
	}, nil
}

func (as *appServer) ListHelmReleases(ctx context.Context, msg *pb.ListHelmReleaseReq) (*pb.ListHelmReleaseRes, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	list := &helmv2.HelmReleaseList{}

	opts := getMatchingLabels(msg.AppName)

	if err := k8s.List(ctx, list, &opts, client.InNamespace(msg.Namespace)); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to list helm releases for app %s: %s", msg.AppName, err.Error())
	}

	if err != nil {
		return nil, status.Errorf(codes.Internal, "unable to get helm repository list: %s", err.Error())
	}

	var results []*pb.HelmRelease
	for _, repository := range list.Items {
		results = append(results, types.HelmReleaseToProto(&repository))
	}

	return &pb.ListHelmReleaseRes{
		HelmReleases: results,
	}, nil
}
