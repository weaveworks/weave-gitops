package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/gitops-server/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/gitops-server/core/clustersmngr/fetcher"
	core "github.com/weaveworks/weave-gitops/gitops-server/core/server"
	"github.com/weaveworks/weave-gitops/gitops-server/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/gitops-server/pkg/server/middleware"
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
	CoreServerConfig core.CoreServerConfig
	AuthServer       *auth.AuthServer
}

func NewHandlers(ctx context.Context, log logr.Logger, cfg *Config) (http.Handler, error) {
	mux := runtime.NewServeMux(middleware.WithGrpcErrorLogging(log))
	httpHandler := middleware.WithLogging(log, mux)

	if AuthEnabled() {
		clustersFetcher, err := fetcher.NewSingleClusterFetcher(cfg.CoreServerConfig.RestCfg)
		if err != nil {
			return nil, fmt.Errorf("failed fetching clusters: %w", err)
		}

		httpHandler = clustersmngr.WithClustersClient(clustersFetcher, httpHandler)
		httpHandler = auth.WithAPIAuth(httpHandler, cfg.AuthServer, PublicRoutes)
	}

	if err := core.Hydrate(ctx, mux, cfg.CoreServerConfig); err != nil {
		return nil, fmt.Errorf("could not start up core servers: %w", err)
	}

	return httpHandler, nil
}
