package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/core/server"
	pbapp "github.com/weaveworks/weave-gitops/pkg/api/applications"
	pbprofiles "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
)

const (
	AuthEnabledFeatureFlag = "WEAVE_GITOPS_AUTH_ENABLED"
)

var (
	PublicRoutes = []string{
		"/v1/featureflags",
	}
)

func AuthEnabled() bool {
	return os.Getenv(AuthEnabledFeatureFlag) == "true"
}

type Config struct {
	AppConfig      *ApplicationsConfig
	AppOptions     []ApplicationsOption
	ProfilesConfig ProfilesConfig
	AuthServer     *auth.AuthServer
}

func NewHandlers(ctx context.Context, cfg *Config) (http.Handler, error) {
	mux := runtime.NewServeMux(middleware.WithGrpcErrorLogging(cfg.AppConfig.Logger))
	httpHandler := middleware.WithLogging(cfg.AppConfig.Logger, mux)
	httpHandler = middleware.WithProviderToken(cfg.AppConfig.JwtClient, httpHandler, cfg.AppConfig.Logger)

	if AuthEnabled() {
		httpHandler = auth.WithAPIAuth(httpHandler, cfg.AuthServer, PublicRoutes)
	}

	appsSrv := NewApplicationsServer(cfg.AppConfig, cfg.AppOptions...)
	if err := pbapp.RegisterApplicationsHandlerServer(ctx, mux, appsSrv); err != nil {
		return nil, fmt.Errorf("could not register application: %w", err)
	}

	profilesSrv := NewProfilesServer(cfg.ProfilesConfig)

	if err := pbprofiles.RegisterProfilesHandlerServer(ctx, mux, profilesSrv); err != nil {
		return nil, fmt.Errorf("could not register profiles: %w", err)
	}

	if err := server.Hydrate(ctx, mux); err != nil {
		return nil, fmt.Errorf("could not start up core servers: %w", err)
	}

	return httpHandler, nil
}
