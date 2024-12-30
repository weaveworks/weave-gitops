package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	core "github.com/weaveworks/weave-gitops/core/server"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
)

var PublicRoutes = []string{
	"/v1/featureflags",
}

type Config struct {
	CoreServerConfig core.CoreServerConfig
	AuthServer       *auth.AuthServer
}

// NewHandlers creates and returns a new server configured to serve the core
// application.
func NewHandlers(ctx context.Context, log logr.Logger, cfg *Config, sm auth.SessionManager) (http.Handler, error) {
	mux := runtime.NewServeMux(middleware.WithGrpcErrorLogging(log))

	if err := core.Hydrate(ctx, mux, cfg.CoreServerConfig); err != nil {
		return nil, fmt.Errorf("could not start up core servers: %w", err)
	}

	httpHandler := auth.WithAPIAuth(mux, cfg.AuthServer, PublicRoutes, sm)

	return httpHandler, nil
}
