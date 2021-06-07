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
	"golang.org/x/oauth2"
)

type Server struct {
	logger logrus.FieldLogger
	oauth  oauth2.Config
}

func NewServer(oauthConfig oauth2.Config) http.Handler {
	defaultHooks := twirp.ChainHooks(middleware.LoggingHooks(), middleware.MetricsHooks())

	gitops := Server{
		logger: logrus.New(),
		oauth:  oauthConfig,
	}

	s := pb.NewGitOpsServer(&gitops, defaultHooks)

	return s
}

func (s *Server) Login(ctx context.Context, msg *pb.LoginReq) (*pb.LoginRes, error) {

	redirectUrl := s.oauth.AuthCodeURL(msg.State)

	return &pb.LoginRes{RedirectUrl: redirectUrl}, nil
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
