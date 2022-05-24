package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	core "github.com/weaveworks/weave-gitops/core/server"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
)

var (
	PublicRoutes = []string{
		"/v1/featureflags",
	}
)

type Config struct {
	AppConfig        *ApplicationsConfig
	AppOptions       []ApplicationsOption
	CoreServerConfig core.CoreServerConfig
	AuthServer       *auth.AuthServer
}

func NewHandlers(ctx context.Context, log logr.Logger, cfg *Config) (http.Handler, error) {
	mux := runtime.NewServeMux(middleware.WithGrpcErrorLogging(log))

	if err := core.Hydrate(ctx, mux, cfg.CoreServerConfig); err != nil {
		return nil, fmt.Errorf("could not start up core servers: %w", err)
	}

	httpHandler := clustersmngr.WithClustersClient(cfg.CoreServerConfig.ClientsFactory, mux)

	httpHandler = auth.WithAPIAuth(httpHandler, cfg.AuthServer, PublicRoutes)

	return httpHandler, nil
}
