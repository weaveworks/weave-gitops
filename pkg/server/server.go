package server

import (
	"context"
	"net/http"

	"github.com/twitchtv/twirp"
	pb "github.com/weaveworks/weave-gitops/pkg/rpc/gitops"
)

type Server struct {
}

func NewServer() http.Handler {
	defaultHooks := twirp.ChainHooks(LoggingHooks())

	gitops := Server{}

	s := pb.NewGitOpsServer(&gitops, defaultHooks)

	return withAuth(s)
}

func (s *Server) ListApplications(ctx context.Context, msg *pb.ListApplicationsReq) (*pb.ListApplicationsRes, error) {
	apps := make([]*pb.Application, 1)

	apps[0] = &pb.Application{Name: "my-cool-app"}

	return &pb.ListApplicationsRes{Applications: apps}, nil
}
