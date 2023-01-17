package server

import (
	"context"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
)

func (cs *coreServer) IsCRDAvailable(ctx context.Context, msg *pb.IsCRDAvailableRequest) (*pb.IsCRDAvailableResponse, error) {
	return &pb.IsCRDAvailableResponse{
		Clusters: cs.crd.IsAvailableOnClusters(msg.Name),
	}, nil
}
