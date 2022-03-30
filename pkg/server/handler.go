package server

import (
	"context"
	"fmt"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	core "github.com/weaveworks/weave-gitops/core/server"
	pbauth "github.com/weaveworks/weave-gitops/pkg/api/gitauth"
	pbprofiles "github.com/weaveworks/weave-gitops/pkg/api/profiles"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/server/gitprovider"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"github.com/weaveworks/weave-gitops/pkg/server/profiles"
	"k8s.io/client-go/rest"
)

const (
	AuthEnabledFeatureFlag = "WEAVE_GITOPS_AUTH_ENABLED"
	// Allowed login requests per second
	loginRequestRateLimit = 20
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
	GitAuthConfig    *gitprovider.GitAuthConfig
	GitAuthOptions   []gitprovider.AuthOption
	ProfilesConfig   profiles.ProfilesConfig
	CoreServerConfig core.CoreServerConfig
	AuthServerConfig auth.AuthConfig
}

func RegisterHandlers(ctx context.Context, log logr.Logger, hmux *http.ServeMux, cfg *Config) (http.Handler, error) {
	rmux := runtime.NewServeMux(middleware.WithGrpcErrorLogging(log))
	httpHandler := middleware.WithLogging(log, rmux)

	if err := AddGRPCHandlers(ctx, rmux, cfg); err != nil {
		return nil, err
	}

	if AuthEnabled() {
		return AddHTTPHandlers(ctx, httpHandler, hmux, cfg.CoreServerConfig.RestCfg, cfg.AuthServerConfig)
	}

	return httpHandler, nil
}

func AddHTTPHandlers(ctx context.Context, h http.Handler, mux *http.ServeMux, rest *rest.Config, authCfg auth.AuthConfig) (http.Handler, error) {
	clustersFetcher, err := clustersmngr.NewSingleClusterFetcher(rest)
	if err != nil {
		return nil, fmt.Errorf("failed fetching clusters: %w", err)
	}

	h = clustersmngr.WithClustersClient(clustersFetcher, h)

	srv, err := auth.NewAuthServer(ctx, authCfg)
	if err != nil {
		return nil, fmt.Errorf("could not create auth server: %w", err)
	}

	if err := auth.RegisterAuthServer(mux, "/oauth2", srv, loginRequestRateLimit); err != nil {
		return nil, fmt.Errorf("failed to register auth routes: %w", err)
	}

	return auth.WithAPIAuth(h, srv, PublicRoutes), nil
}

func AddGRPCHandlers(ctx context.Context, mux *runtime.ServeMux, cfg *Config) error {
	gitAuthSrv := gitprovider.NewGitAuthServer(cfg.GitAuthConfig, cfg.GitAuthOptions...)
	if err := pbauth.RegisterGitProviderAuthHandlerServer(ctx, mux, gitAuthSrv); err != nil {
		return fmt.Errorf("could not register git auth handlers: %w", err)
	}

	profilesSrv := profiles.NewProfilesServer(cfg.ProfilesConfig)
	if err := pbprofiles.RegisterProfilesHandlerServer(ctx, mux, profilesSrv); err != nil {
		return fmt.Errorf("could not register profiles handlers: %w", err)
	}

	return core.Hydrate(ctx, mux, cfg.CoreServerConfig)
}

func Health(log logr.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		_, err := w.Write([]byte("ok"))

		if err != nil {
			log.Error(err, "error writing health check")
		}
	})
}

func StaticAssets(assetFS fs.FS, log logr.Logger) http.Handler {
	assetHandler := http.FileServer(http.FS(assetFS))
	redirector := newRedirector(assetFS, log)

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Assume anything with a file extension in the name is a static asset.
		extension := filepath.Ext(req.URL.Path)
		// We use the golang http.FileServer for static file requests.
		// This will return a 404 on normal page requests, ie /some-page.
		// Redirect all non-file requests to index.html, where the JS routing will take over.
		if extension == "" {
			redirector(w, req)
			return
		}
		assetHandler.ServeHTTP(w, req)
	})
}

func newRedirector(fsys fs.FS, log logr.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		indexPage, err := fsys.Open("index.html")

		if err != nil {
			log.Error(err, "could not open index.html page")
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		stat, err := indexPage.Stat()
		if err != nil {
			log.Error(err, "could not get index.html stat")
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		bt := make([]byte, stat.Size())
		_, err = indexPage.Read(bt)

		if err != nil {
			log.Error(err, "could not read index.html")
			w.WriteHeader(http.StatusInternalServerError)

			return
		}

		_, err = w.Write(bt)

		if err != nil {
			log.Error(err, "error writing index.html")
			w.WriteHeader(http.StatusInternalServerError)

			return
		}
	}
}
