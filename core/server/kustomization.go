package server

import (
	"context"
	"fmt"
	"sync"

	kustomizev1 "github.com/fluxcd/kustomize-controller/api/v1beta2"
	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/app"
	"github.com/weaveworks/weave-gitops/pkg/kube"
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

	kust := types.ProtoToKustomization(msg)

	if err := k8s.Create(ctx, &kust); err != nil {
		return nil, status.Errorf(codes.Internal, "creating kustomization for app %q: %s", msg.AppName, err.Error())
	}

	return &pb.AddKustomizationRes{
		Success:       true,
		Kustomization: types.KustomizationToProto(&kust),
	}, nil
}

func (as *appServer) ListKustomizations(ctx context.Context, msg *pb.ListKustomizationsReq) (*pb.ListKustomizationsRes, error) {
	k8s, err := as.k8s.Client(ctx)
	if err != nil {
		return nil, doClientError(err)
	}

	list := &kustomizev1.KustomizationList{}

	opts := getMatchingLabels(msg.AppName)

	if err := k8s.List(ctx, list, &opts, client.InNamespace(msg.Namespace)); err != nil {
		return nil, status.Errorf(codes.Internal, "unable to create new app: %s", err.Error())
	}

	var results []*pb.Kustomization
	for _, kustomization := range list.Items {
		results = append(results, types.KustomizationToProto(&kustomization))
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

func (as *appServer) ListKustomizationsForClusters(ctx context.Context, msg *pb.ListKustomizationsForClustersReq) (*pb.ListKustomizationsForClustersRes, error) {
	clusters := msg.Clusters

	result := []*pb.RemoteKustomization{}

	wg := &sync.WaitGroup{}
	wg.Add(len(msg.Clusters))

	errs := make(chan error)

	for _, c := range clusters {
		go func(ctx context.Context, cluster string, r []*pb.RemoteKustomization) {
			defer wg.Done()

			shallowCopy := *as.k8s.cfg

			cfg, err := as.remoteK8s.GetByName(ctx, cluster)
			if err != nil {
				errs <- fmt.Errorf("getting cluster config %s: %w\n", cluster, err)
				return
			}

			cfg.TLSClientConfig = shallowCopy.TLSClientConfig

			_, c, err := kube.NewKubeHTTPClientWithConfig(cfg, "")
			if err != nil {
				errs <- fmt.Errorf("getting client for %s: %w\n", cluster, err)
				return
			}

			list := &kustomizev1.KustomizationList{}
			if err := c.List(ctx, list, client.InNamespace(msg.Namespace)); err != nil {
				errs <- fmt.Errorf("listing kustomizations for %s: %w\n", cluster, err)
				return
			}

			for _, k := range list.Items {
				result = append(result, &pb.RemoteKustomization{
					ClusterName:   cluster,
					Kustomization: types.KustomizationToProto(&k),
				})
			}
		}(ctx, c, result)
	}

	if err := <-errs; err != nil {
		return nil, err
	}

	wg.Wait()

	return &pb.ListKustomizationsForClustersRes{
		Kustomizations: result,
	}, nil
}
