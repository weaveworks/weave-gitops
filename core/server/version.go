package server

import (
	"context"

	pb "github.com/weaveworks/weave-gitops/pkg/api/core"
)

// Variables that we'll set @ build time
var (
	Version   = "v0.0.0"
	GitCommit = ""
	Branch    = ""
	Buildtime = ""
)

func (cs *coreServer) GetVersion(ctx context.Context, msg *pb.GetVersionRequest) (*pb.GetVersionResponse, error) {
	return &pb.GetVersionResponse{
		Semver:    Version,
		Commit:    GitCommit,
		Branch:    Branch,
		BuildTime: Buildtime,
	}, nil
}
