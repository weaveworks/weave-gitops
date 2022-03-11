package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/core/multicluster"
	core "github.com/weaveworks/weave-gitops/core/server"
	pbapp "github.com/weaveworks/weave-gitops/pkg/api/applications"
	pbprofiles "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/kube"
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
	AppConfig        *ApplicationsConfig
	AppOptions       []ApplicationsOption
	ProfilesConfig   ProfilesConfig
	CoreServerConfig core.CoreServerConfig
	AuthServer       *auth.AuthServer
}

func NewHandlers(ctx context.Context, cfg *Config) (http.Handler, error) {
	mux := runtime.NewServeMux(middleware.WithGrpcErrorLogging(cfg.AppConfig.Logger))
	httpHandler := middleware.WithLogging(cfg.AppConfig.Logger, mux)
	httpHandler = middleware.WithProviderToken(cfg.AppConfig.JwtClient, httpHandler, cfg.AppConfig.Logger)

	restCfg, _, err := kube.RestConfig()
	if err != nil {
		return nil, fmt.Errorf("building rest config: %w", err)
	}

	if AuthEnabled() {
		clustersFetcher, err := multicluster.NewSingleClustersFetcher(restCfg, wego.DefaultNamespace)
		if err != nil {
			return nil, fmt.Errorf("failed fetching clusters: %w", err)
		}

		httpHandler = multicluster.WithClustersClients(clustersFetcher, httpHandler)
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

	if err := core.Hydrate(ctx, mux, cfg.CoreServerConfig); err != nil {
		return nil, fmt.Errorf("could not start up core servers: %w", err)
	}

	return httpHandler, nil
}
