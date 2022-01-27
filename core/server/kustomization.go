package server

import (
	"context"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	stypes "github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (as *appServer) AddKustomization(ctx context.Context, msg *pb.AddKustomizationReq) (*pb.AddKustomizationRes, error) {
	if msg.SourceRef == nil {
		return nil, status.Errorf(codes.InvalidArgument, "sourceRef is required")
	}

	k8s, err := as.k8s.Client(ctx)
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
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	list := &kustomizev1.KustomizationList{}

	var opts client.MatchingLabels
	if msg.AppName != "" {
		opts = client.MatchingLabels{
			"app.kubernetes.io/part-of": msg.AppName,
		}
	}

	if err := k8s.List(ctx, list, &opts, client.InNamespace(msg.Namespace)); err != nil {
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
	k8s, err := as.k8s.Client(ctx)
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
