package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/core/server"
	"github.com/weaveworks/weave-gitops/pkg/kube"
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
	AppConfig  *ApplicationsConfig
	AuthServer *auth.AuthServer
}

func NewHandlers(ctx context.Context, cfg *Config) (http.Handler, error) {
	mux := runtime.NewServeMux(middleware.WithGrpcErrorLogging(cfg.AppConfig.Logger))
	httpHandler := middleware.WithLogging(cfg.AppConfig.Logger, mux)

	if AuthEnabled() {
		httpHandler = auth.WithAPIAuth(httpHandler, cfg.AuthServer)
	}

	restCfg, _, err := kube.RestConfig()
	if err != nil {
		return nil, fmt.Errorf("building rest config: %w", err)
	}

	if err := server.Hydrate(ctx, mux, restCfg); err != nil {
		return nil, fmt.Errorf("could not start up core servers: %w", err)
	}

	return httpHandler, nil
}
