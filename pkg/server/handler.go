package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	pbapp "github.com/weaveworks/weave-gitops/pkg/api/applications"
	pbprofiles "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
)

const (
	AuthEnabledFeatureFlag = "WEAVE_GITOPS_AUTH_ENABLED"
)

func AuthEnabled() bool {
	return os.Getenv(AuthEnabledFeatureFlag) == "true"
}

type Config struct {
	AppConfig      *ApplicationsConfig
	ProfilesConfig ProfilesConfig
	AuthConfig     *auth.AuthConfig
}

func NewHandlers(ctx context.Context, cfg *Config) (http.Handler, error) {
	mux := runtime.NewServeMux(middleware.WithGrpcErrorLogging(cfg.AppConfig.Logger))
	httpHandler := middleware.WithLogging(cfg.AppConfig.Logger, mux)
	httpHandler = middleware.WithProviderToken(cfg.AppConfig.JwtClient, httpHandler, cfg.AppConfig.Logger)

	if AuthEnabled() {
		httpHandler = auth.WithAPIAuth(httpHandler, cfg.AuthConfig)
	}

	appsSrv := NewApplicationsServer(cfg.AppConfig)
	if err := pbapp.RegisterApplicationsHandlerServer(ctx, mux, appsSrv); err != nil {
		return nil, fmt.Errorf("could not register application: %w", err)
	}

	profilesSrv, err := NewProfilesServer(cfg.ProfilesConfig)
	if err != nil {
		return nil, fmt.Errorf("could not create profiles server: %w", err)
	}

	if err := pbprofiles.RegisterProfilesHandlerServer(ctx, mux, profilesSrv); err != nil {
		return nil, fmt.Errorf("could not register profiles: %w", err)
	}

	return httpHandler, nil
}
