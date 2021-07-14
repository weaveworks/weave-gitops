package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/middleware"
	"golang.org/x/oauth2"
	"k8s.io/apimachinery/pkg/types"
)

type server struct {
	pb.UnimplementedApplicationsServer

	kube kube.Kube
}

func NewApplicationsServer(kubeSvc kube.Kube) pb.ApplicationsServer {

	return &server{
		kube: kubeSvc,
	}
}

func (s *server) ListApplications(ctx context.Context, msg *pb.ListApplicationsRequest) (*pb.ListApplicationsResponse, error) {
	apps, err := s.kube.GetApplications(ctx, msg.GetNamespace())
	if err != nil {
		return nil, err
	}

	if apps == nil {
		return &pb.ListApplicationsResponse{
			Applications: []*pb.Application{},
		}, nil
	}

	list := []*pb.Application{}
	for _, a := range apps {
		list = append(list, &pb.Application{Name: a.Name})
	}
	return &pb.ListApplicationsResponse{
		Applications: list,
	}, nil
}

func (s *server) GetApplication(ctx context.Context, msg *pb.GetApplicationRequest) (*pb.GetApplicationResponse, error) {
	app, err := s.kube.GetApplication(ctx, types.NamespacedName{Name: msg.Name, Namespace: msg.Namespace})

	if err != nil {
		return nil, fmt.Errorf("could not get application: %s", err)
	}

	return &pb.GetApplicationResponse{Application: &pb.Application{
		Name: app.Name,
		Url:  app.Spec.URL,
		Path: app.Spec.Path,
	}}, nil
}

func (s *server) GetAuthenticationProviders(ctx context.Context, msg *pb.GetAuthenticationProvidersRequest) (*pb.GetAuthenticationProvidersResponse, error) {
	providers := gitproviders.GetSupportedProviders()

	oauthProviders := []*pb.OauthProvider{}
	for _, name := range providers {
		gp, _ := gitproviders.New(name)
		url := gp.OauthConfig().AuthCodeURL("state", oauth2.AccessTypeOffline)
		oauthProviders = append(oauthProviders, &pb.OauthProvider{
			Name:    string(name),
			AuthUrl: url,
		})
	}

	return &pb.GetAuthenticationProvidersResponse{Providers: oauthProviders}, nil
}

func (s *server) Authenticate(ctx context.Context, msg *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
	gh, err := gitproviders.New(gitproviders.ProviderName(msg.ProviderName))
	if err != nil {
		return nil, makeHTTPError(http.StatusBadRequest, err)
	}

	conf := gh.OauthConfig()

	token, err := conf.Exchange(ctx, msg.Code)
	if err != nil {
		return nil, fmt.Errorf("could not exchange code: %w", err)
	}

	user, err := gh.GetUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("could not get user: %w", err)
	}

	return &pb.AuthenticateResponse{
		User:  &pb.User{Email: user.Email},
		Token: token.AccessToken,
	}, nil
}

func (s *server) GetUser(ctx context.Context, msg *pb.GetUserRequest) (*pb.GetUserResponse, error) {
	token, err := middleware.ExtractToken(ctx)
	if err != nil {
		return nil, makeHTTPError(http.StatusUnauthorized, err)
	}

	gp, err := gitproviders.New(gitproviders.ProviderNameGithub)
	if err != nil {
		return nil, makeHTTPError(http.StatusBadRequest, err)
	}

	user, err := gp.GetUser(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("could not get user: %w", err)
	}

	return &pb.GetUserResponse{
		User: &pb.User{Email: user.Email},
	}, nil

}

func makeHTTPError(status int, err error) error {
	return &runtime.HTTPStatusError{HTTPStatus: status, Err: err}
}
