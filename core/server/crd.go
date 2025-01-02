package server

import (
	"context"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
)

// IsCRDAvailable returns with a hashmap where the keys are the names of
// the clusters, and the value is a boolean indicating whether given CRD is
// installed or not on that cluster.
func (cs *coreServer) IsCRDAvailable(ctx context.Context, msg *pb.IsCRDAvailableRequest) (*pb.IsCRDAvailableResponse, error) {
	return &pb.IsCRDAvailableResponse{
		Clusters: cs.crd.IsAvailableOnClusters(ctx, msg.Name),
	}, nil
}
