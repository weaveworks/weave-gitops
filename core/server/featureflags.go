package server

import (
	"context"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
)

func (cs *coreServer) GetFeatureFlags(ctx context.Context, msg *pb.GetFeatureFlagsRequest) (*pb.GetFeatureFlagsResponse, error) {
	return &pb.GetFeatureFlagsResponse{
		Flags: featureflags.GetFlags(),
	}, nil
}
