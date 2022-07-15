package server

import (
	"context"
	"fmt"

	"github.com/weaveworks/weave-gitops/core/server/types"
	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (cs *coreServer) GetObject(ctx context.Context, msg *pb.GetObjectRequest) (*pb.GetObjectResponse, error) {
	clustersClient, err := cs.clientsFactory.GetImpersonatedClient(ctx, auth.Principal(ctx))
	if err != nil {
		return nil, fmt.Errorf("error getting impersonating client: %w", err)
	}

	gvk, err := cs.primaryKinds.Lookup(msg.Kind)
	if err != nil {
		return nil, err
	}

	obj := unstructured.Unstructured{}
	obj.SetGroupVersionKind(*gvk)

	key := client.ObjectKey{
		Name:      msg.Name,
		Namespace: msg.Namespace,
	}

	if err := clustersClient.Get(ctx, msg.ClusterName, key, &obj); err != nil {
		return nil, err
	}

	res, err := types.K8sObjectToProto(&obj, msg.ClusterName)

	if err != nil {
		return nil, fmt.Errorf("converting kustomization to proto: %w", err)
	}

	return &pb.GetObjectResponse{Object: res}, nil
}
