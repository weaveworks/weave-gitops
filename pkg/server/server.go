package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
	"github.com/twitchtv/twirp"
	"github.com/weaveworks/weave-gitops/pkg/cmdimpl"
	"github.com/weaveworks/weave-gitops/pkg/middleware"
	pb "github.com/weaveworks/weave-gitops/pkg/rpc/gitops"
)

type Server struct {
	logger logrus.FieldLogger
}

func NewServer() http.Handler {
	defaultHooks := twirp.ChainHooks(middleware.LoggingHooks())

	gitops := Server{
		logger: logrus.New(),
	}

	s := pb.NewGitOpsServer(&gitops, defaultHooks)

	return middleware.WithAuth(s)
}

func (s *Server) AddApplication(ctx context.Context, msg *pb.AddApplicationReq) (*pb.AddApplicationRes, error) {

	params := cmdimpl.AddParamSet{
		Dir:            msg.Dir,
		Name:           msg.Name,
		Owner:          msg.Owner,
		Url:            msg.Url,
		Path:           msg.Path,
		Branch:         msg.Branch,
		PrivateKey:     msg.PrivateKey,
		DeploymentType: msg.DeploymentType.String(),
		Namespace:      msg.Namespace,
	}

	s.logger.Debug("a debug message")

	if err := cmdimpl.Add([]string{msg.Dir}, params); err != nil {
		return nil, fmt.Errorf("could not add application: %v", err)
	}

	return &pb.AddApplicationRes{Application: &pb.Application{Name: msg.Name}}, nil
}
