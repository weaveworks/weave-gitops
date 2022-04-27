package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/go-logr/logr"
	"google.golang.org/grpc/codes"
	grpcStatus "google.golang.org/grpc/status"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	commonkube "github.com/weaveworks/weave-gitops/common/pkg/kube"
	pb "github.com/weaveworks/weave-gitops/for-enterprise/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/for-enterprise/pkg/server/middleware"
	"github.com/weaveworks/weave-gitops/gitops-server/pkg/kube"
	"github.com/weaveworks/weave-gitops/gitops/pkg/flux"
	"github.com/weaveworks/weave-gitops/gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/gitops/pkg/runner"
	"github.com/weaveworks/weave-gitops/gitops/pkg/services"
	"github.com/weaveworks/weave-gitops/gitops/pkg/services/auth"
)

var (
	ErrEmptyAccessToken = errors.New("access token is empty")
	ErrBadProvider      = errors.New("wrong provider name")
)

type applicationServer struct {
	pb.UnimplementedApplicationsServer

	factory      services.Factory
	jwtClient    auth.JWTClient
	log          logr.Logger
	ghAuthClient auth.GithubAuthClient
	glAuthClient auth.GitlabAuthClient
	clientGetter kube.ClientGetter
}

// An ApplicationsConfig allows for the customization of an ApplicationsServer.
// Use the DefaultConfig() to use the default dependencies.
type ApplicationsConfig struct {
	Logger           logr.Logger
	Factory          services.Factory
	JwtClient        auth.JWTClient
	GithubAuthClient auth.GithubAuthClient
	GitlabAuthClient auth.GitlabAuthClient
	ClusterConfig    commonkube.ClusterConfig
}

// NewApplicationsServer creates a grpc Applications server
func NewApplicationsServer(cfg *ApplicationsConfig, setters ...ApplicationsOption) pb.ApplicationsServer {
	configGetter := kube.NewImpersonatingConfigGetter(cfg.ClusterConfig.DefaultConfig, false)
	clientGetter := kube.NewDefaultClientGetter(configGetter, cfg.ClusterConfig.ClusterName)

	args := &ApplicationsOptions{
		ClientGetter: clientGetter,
	}

	for _, setter := range setters {
		setter(args)
	}

	return &applicationServer{
		jwtClient:    cfg.JwtClient,
		log:          cfg.Logger,
		factory:      cfg.Factory,
		ghAuthClient: cfg.GithubAuthClient,
		glAuthClient: cfg.GitlabAuthClient,
		clientGetter: args.ClientGetter,
	}
}

// DefaultApplicationsConfig creates a populated config with the dependencies for a Server
func DefaultApplicationsConfig(log logr.Logger) (*ApplicationsConfig, error) {
	rand.Seed(time.Now().UnixNano())
	secretKey := rand.String(20)
	envSecretKey := os.Getenv("GITOPS_JWT_ENCRYPTION_SECRET")

	if envSecretKey != "" {
		secretKey = envSecretKey
	}

	jwtClient := auth.NewJwtClient(secretKey)
	clusterName := commonkube.InClusterConfigClusterName()

	rest, err := config.GetConfig()
	if err != nil {
		return nil, fmt.Errorf("could not create client config: %w", err)
	}

	fluxClient := flux.New(&runner.CLIRunner{})

	return &ApplicationsConfig{
		Logger:           log.WithName("app-server"),
		Factory:          services.NewFactory(fluxClient, log.WithName("services")),
		JwtClient:        jwtClient,
		GithubAuthClient: auth.NewGithubAuthClient(http.DefaultClient),
		GitlabAuthClient: auth.NewGitlabAuthClient(http.DefaultClient),
		ClusterConfig: commonkube.ClusterConfig{
			DefaultConfig: rest,
			ClusterName:   clusterName,
		},
	}, nil
}

func (s *applicationServer) GetGithubDeviceCode(ctx context.Context, msg *pb.GetGithubDeviceCodeRequest) (*pb.GetGithubDeviceCodeResponse, error) {
	res, err := s.ghAuthClient.GetDeviceCode()
	if err != nil {
		return nil, fmt.Errorf("error doing github code request: %w", err)
	}

	return &pb.GetGithubDeviceCodeResponse{
		UserCode:      res.UserCode,
		ValidationURI: res.VerificationURI,
		DeviceCode:    res.DeviceCode,
		Interval:      int32(res.Interval),
	}, nil
}

func (s *applicationServer) GetGithubAuthStatus(ctx context.Context, msg *pb.GetGithubAuthStatusRequest) (*pb.GetGithubAuthStatusResponse, error) {
	token, err := s.ghAuthClient.GetDeviceCodeAuthStatus(msg.DeviceCode)
	if err == auth.ErrAuthPending {
		return nil, grpcStatus.Error(codes.Unauthenticated, err.Error())
	} else if err != nil {
		return nil, fmt.Errorf("error getting github device code status: %w", err)
	}

	t, err := s.jwtClient.GenerateJWT(auth.ExpirationTime, gitproviders.GitProviderGitHub, token)
	if err != nil {
		return nil, fmt.Errorf("could not generate token: %w", err)
	}

	return &pb.GetGithubAuthStatusResponse{AccessToken: t}, nil
}

// Authenticate generates and returns a jwt token using git provider name and git provider token
func (s *applicationServer) Authenticate(_ context.Context, msg *pb.AuthenticateRequest) (*pb.AuthenticateResponse, error) {
	if !strings.HasPrefix(github.DefaultDomain, msg.ProviderName) &&
		!strings.HasPrefix(gitlab.DefaultDomain, msg.ProviderName) {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "%s expected github or gitlab, got %s", ErrBadProvider, msg.ProviderName)
	}

	if msg.AccessToken == "" {
		return nil, grpcStatus.Error(codes.InvalidArgument, ErrEmptyAccessToken.Error())
	}

	token, err := s.jwtClient.GenerateJWT(auth.ExpirationTime, gitproviders.GitProviderName(msg.GetProviderName()), msg.GetAccessToken())
	if err != nil {
		return nil, grpcStatus.Errorf(codes.Internal, "error generating jwt token. %s", err)
	}

	return &pb.AuthenticateResponse{Token: token}, nil
}

func (s *applicationServer) ParseRepoURL(ctx context.Context, msg *pb.ParseRepoURLRequest) (*pb.ParseRepoURLResponse, error) {
	u, err := gitproviders.NewRepoURL(msg.Url)
	if err != nil {
		return nil, grpcStatus.Errorf(codes.InvalidArgument, "could not parse url: %s", err.Error())
	}

	return &pb.ParseRepoURLResponse{
		Name:     u.RepositoryName(),
		Owner:    u.Owner(),
		Provider: toProtoProvider(u.Provider()),
	}, nil
}

func (s *applicationServer) GetGitlabAuthURL(ctx context.Context, msg *pb.GetGitlabAuthURLRequest) (*pb.GetGitlabAuthURLResponse, error) {
	u, err := s.glAuthClient.AuthURL(ctx, msg.RedirectUri)
	if err != nil {
		return nil, fmt.Errorf("could not get gitlab auth url: %w", err)
	}

	return &pb.GetGitlabAuthURLResponse{Url: u.String()}, nil
}

func (s *applicationServer) AuthorizeGitlab(ctx context.Context, msg *pb.AuthorizeGitlabRequest) (*pb.AuthorizeGitlabResponse, error) {
	tokenState, err := s.glAuthClient.ExchangeCode(ctx, msg.RedirectUri, msg.Code)
	if err != nil {
		return nil, fmt.Errorf("could not exchange code: %w", err)
	}

	token, err := s.jwtClient.GenerateJWT(tokenState.ExpiresInSeconds, gitproviders.GitProviderGitLab, tokenState.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("could not generate token: %w", err)
	}

	return &pb.AuthorizeGitlabResponse{Token: token}, nil
}

func (s *applicationServer) ValidateProviderToken(ctx context.Context, msg *pb.ValidateProviderTokenRequest) (*pb.ValidateProviderTokenResponse, error) {
	token, err := middleware.ExtractProviderToken(ctx)
	if err != nil {
		return nil, grpcStatus.Error(codes.Unauthenticated, err.Error())
	}

	v, err := findValidator(msg.Provider, s)
	if err != nil {
		return nil, grpcStatus.Error(codes.InvalidArgument, err.Error())
	}

	if err := v.ValidateToken(ctx, token.AccessToken); err != nil {
		return nil, grpcStatus.Error(codes.InvalidArgument, err.Error())
	}

	return &pb.ValidateProviderTokenResponse{
		Valid: true,
	}, nil
}

func toProtoProvider(p gitproviders.GitProviderName) pb.GitProvider {
	switch p {
	case gitproviders.GitProviderGitHub:
		return pb.GitProvider_GitHub
	case gitproviders.GitProviderGitLab:
		return pb.GitProvider_GitLab
	}

	return pb.GitProvider_Unknown
}

func findValidator(provider pb.GitProvider, s *applicationServer) (auth.ProviderTokenValidator, error) {
	switch provider {
	case pb.GitProvider_GitHub:
		return s.ghAuthClient, nil
	case pb.GitProvider_GitLab:
		return s.glAuthClient, nil
	}

	return nil, fmt.Errorf("unknown git provider %s", provider)
}
