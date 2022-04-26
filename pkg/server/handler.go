package server

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/fetcher"
	core "github.com/weaveworks/weave-gitops/core/server"
	pbapp "github.com/weaveworks/weave-gitops/pkg/api/applications"
	pbprofiles "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
)

const (
	AuthEnabledFeatureFlag = "WEAVE_GITOPS_AUTH_ENABLED"
	TlsDisabledFeatureFlag = "WEAVE_GITOPS_TLS_DISABLED"
	DevModeFeatureFlag     = "WEAVE_GITOPS_DEV_MODE_ENABLED"
	ClusterUserAuthFlag    = "CLUSTER_USER_AUTH"
	OidcAuthFlag           = "OIDC_AUTH"
)

var (
	PublicRoutes = []string{
		"/v1/featureflags",
	}
)

func AuthEnabled() bool {
	return os.Getenv(AuthEnabledFeatureFlag) == "true"
}

func TlsEnabled() bool {
	return os.Getenv(TlsDisabledFeatureFlag) != "true"
}

type Config struct {
	AppConfig        *ApplicationsConfig
	AppOptions       []ApplicationsOption
	ProfilesConfig   ProfilesConfig
	CoreServerConfig core.CoreServerConfig
	AuthServer       *auth.AuthServer
}

func NewHandlers(ctx context.Context, log logr.Logger, cfg *Config) (http.Handler, error) {
	mux := runtime.NewServeMux(middleware.WithGrpcErrorLogging(log))
	httpHandler := middleware.WithLogging(log, mux, TlsEnabled())

	if AuthEnabled() {
		clustersFetcher, err := fetcher.NewSingleClusterFetcher(cfg.CoreServerConfig.RestCfg)
		if err != nil {
			return nil, fmt.Errorf("failed fetching clusters: %w", err)
		}

		httpHandler = clustersmngr.WithClustersClient(clustersFetcher, httpHandler)
		httpHandler = auth.WithAPIAuth(httpHandler, cfg.AuthServer, PublicRoutes)
	}

	appsSrv := NewApplicationsServer(cfg.AppConfig, cfg.AppOptions...)
	if err := pbapp.RegisterApplicationsHandlerServer(ctx, mux, appsSrv); err != nil {
		return nil, fmt.Errorf("could not register application: %w", err)
	}

	profilesSrv := NewProfilesServer(log, cfg.ProfilesConfig)

	if err := pbprofiles.RegisterProfilesHandlerServer(ctx, mux, profilesSrv); err != nil {
		return nil, fmt.Errorf("could not register profiles: %w", err)
	}

	if err := core.Hydrate(ctx, mux, cfg.CoreServerConfig); err != nil {
		return nil, fmt.Errorf("could not start up core servers: %w", err)
	}

	return httpHandler, nil
}
