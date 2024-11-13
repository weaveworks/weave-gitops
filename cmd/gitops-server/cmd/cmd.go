package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/alexedwards/scs/v2"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	httpmiddleware "github.com/slok/go-http-metrics/middleware"
	httpmiddlewarestd "github.com/slok/go-http-metrics/middleware/std"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/fetcher"
	"github.com/weaveworks/weave-gitops/core/logger"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
	core "github.com/weaveworks/weave-gitops/core/server"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
	"github.com/weaveworks/weave-gitops/pkg/health"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"github.com/weaveworks/weave-gitops/pkg/telemetry"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	k8sMetrics "sigs.k8s.io/controller-runtime/pkg/metrics"
)

const (
	// Allowed login requests per second
	loginRequestRateLimit            = 20
	InsecureNoAuthenticationUserFlag = "insecure-no-authentication-user"
)

// Options contains all the options for the gitops-server command.
type Options struct {
	// System config
	Host                          string
	LogLevel                      string
	NotificationControllerAddress string
	Path                          string
	RoutePrefix                   string
	Port                          string
	AuthMethods                   []string
	// TLS config
	Insecure    bool
	MTLS        bool
	TLSCertFile string
	TLSKeyFile  string
	// Stuff for profiles apparently
	HelmRepoName      string
	HelmRepoNamespace string
	// OIDC
	OIDC       auth.OIDCConfig
	OIDCSecret string
	// Auth
	NoAuthUser string
	// Dev mode
	DevMode bool
	// Metrics
	EnableMetrics  bool
	MetricsAddress string

	UseK8sCachedClients bool
}

var options Options

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Short: "Runs the gitops-server",
		RunE:  runCmd,
	}

	options = Options{
		OIDC: auth.OIDCConfig{
			ClaimsConfig: &auth.ClaimsConfig{},
		},
	}

	// System config
	cmd.Flags().StringVar(&options.Host, "host", server.DefaultHost, "UI host")
	cmd.Flags().StringVar(&options.LogLevel, "log-level", logger.DefaultLogLevel, "log level")
	cmd.Flags().StringVar(&options.NotificationControllerAddress, "notification-controller-address", "", "the address of the notification-controller running in the cluster")
	cmd.Flags().StringVar(&options.RoutePrefix, "route-prefix", "", "Mount the UI and API endpoint under a path prefix, e.g. /weave-gitops")
	cmd.Flags().StringVar(&options.Port, "port", server.DefaultPort, "UI port")
	cmd.Flags().StringSliceVar(&options.AuthMethods, "auth-methods", auth.DefaultAuthMethodStrings(), fmt.Sprintf("Which auth methods to use, valid values are %s", strings.Join(auth.AllUserAuthMethods(), ",")))
	cmd.Flags().BoolVar(&options.UseK8sCachedClients, "use-k8s-cached-clients", false, "Enables the use of cached clients")
	//  TLS
	cmd.Flags().BoolVar(&options.Insecure, "insecure", false, "do not attempt to read TLS certificates")
	cmd.Flags().BoolVar(&options.MTLS, "mtls", false, "disable enforce mTLS")
	cmd.Flags().StringVar(&options.TLSCertFile, "tls-cert-file", "", "filename for the TLS certificate, in-memory generated if omitted")
	cmd.Flags().StringVar(&options.TLSKeyFile, "tls-private-key-file", "", "filename for the TLS key, in-memory generated if omitted")
	// OIDC
	cmd.Flags().StringVar(&options.OIDCSecret, "oidc-secret-name", auth.DefaultOIDCAuthSecretName, "Name of the secret that contains OIDC configuration")
	cmd.Flags().StringVar(&options.OIDC.ClientID, "oidc-client-id", "", "The client ID for the OpenID Connect client")
	cmd.Flags().StringVar(&options.OIDC.ClientSecret, "oidc-client-secret", "", "The client secret to use with OpenID Connect issuer")
	cmd.Flags().StringVar(&options.OIDC.IssuerURL, "oidc-issuer-url", "", "The URL of the OpenID Connect issuer")
	cmd.Flags().StringVar(&options.OIDC.RedirectURL, "oidc-redirect-url", "", "The OAuth2 redirect URL")
	cmd.Flags().DurationVar(&options.OIDC.TokenDuration, "oidc-token-duration", time.Hour, "The duration of the ID token. It should be set in the format: number + time unit (s,m,h) e.g., 20m")
	cmd.Flags().StringVar(&options.OIDC.ClaimsConfig.Username, "oidc-username-claim", auth.ClaimUsername, "JWT claim to use as the user name. By default email, which is expected to be a unique identifier of the end user. Admins can choose other claims, such as sub or name, depending on their provider")
	cmd.Flags().StringVar(&options.OIDC.ClaimsConfig.Groups, "oidc-groups-claim", auth.ClaimGroups, "JWT claim to use as the user's group. If the claim is present it must be an array of strings")
	cmd.Flags().StringSliceVar(&options.OIDC.Scopes, "custom-oidc-scopes", auth.DefaultScopes, "Customise the requested scopes for then OIDC authentication flow - openid will always be requested")
	// OIDC prefixes
	cmd.Flags().StringVar(&options.OIDC.UsernamePrefix, "oidc-username-prefix", "", "Prefix to add to the username when impersonating")
	cmd.Flags().StringVar(&options.OIDC.GroupsPrefix, "oidc-group-prefix", "", "Prefix to add to the groups when impersonating")
	// auth
	cmd.Flags().StringVar(&options.NoAuthUser, InsecureNoAuthenticationUserFlag, "", "A kubernetes user to impersonate for all requests, no authentication will be performed")

	// Metrics
	cmd.Flags().BoolVar(&options.EnableMetrics, "enable-metrics", false, "Starts the metrics listener")
	cmd.Flags().StringVar(&options.MetricsAddress, "metrics-address", ":2112", "If the metrics listener is enabled, bind to this address")

	return cmd
}

func runCmd(cmd *cobra.Command, args []string) error {
	log, err := logger.New(options.LogLevel, options.Insecure)
	if err != nil {
		return err
	}

	log.Info("Version", "version", core.Version, "git-commit", core.GitCommit, "branch", core.Branch, "buildtime", core.Buildtime)

	featureflags.SetFromEnv(os.Environ())

	if cmd.Flags().Changed(InsecureNoAuthenticationUserFlag) && options.NoAuthUser == "" {
		return fmt.Errorf("%s flag set but no user specified", InsecureNoAuthenticationUserFlag)
	}

	mux := http.NewServeMux()

	mux.Handle("/health/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))

		if err != nil {
			log.Error(err, "error writing health check")
		}
	}))

	clusterName := kube.InClusterConfigClusterName()

	rest, err := config.GetConfig()
	if err != nil {
		return fmt.Errorf("could not create client config: %w", err)
	}

	scheme, err := kube.CreateScheme()
	if err != nil {
		return fmt.Errorf("could not create scheme: %w", err)
	}

	rawClient, err := client.New(rest, client.Options{
		Scheme: scheme,
	})
	if err != nil {
		return fmt.Errorf("could not create kube http client: %w", err)
	}

	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)

	namespace, _, err := kubeConfig.Namespace()
	if err != nil {
		return fmt.Errorf("couldn't get current namespace")
	}

	sessionManager := scs.New()
	// TODO: Make this configurable
	sessionManager.Lifetime = 24 * time.Hour
	authServer, err := auth.InitAuthServer(cmd.Context(), log, rawClient, auth.AuthParams{
		OIDCConfig:        options.OIDC,
		OIDCSecretName:    options.OIDCSecret,
		AuthMethodStrings: options.AuthMethods,
		NoAuthUser:        options.NoAuthUser,
		Namespace:         namespace,
		SessionManager:    sessionManager,
	})
	if err != nil {
		return fmt.Errorf("could not initialise authentication server: %w", err)
	}

	log.Info("Registering auth routes")

	if err := auth.RegisterAuthServer(mux, "/oauth2", authServer, loginRequestRateLimit); err != nil {
		return fmt.Errorf("failed to register auth routes: %w", err)
	}

	ctx := context.Background()

	oidcPrefixes := kube.UserPrefixes{
		UsernamePrefix: options.OIDC.UsernamePrefix,
		GroupsPrefix:   options.OIDC.GroupsPrefix,
	}

	// Incorporate values from authServer.AuthConfig.OIDCConfig
	if authServer.OIDCConfig.UsernamePrefix != "" {
		log.V(logger.LogLevelWarn).Info("OIDC username prefix configured by both CLI and secret. Secret values will take precedence.")
		oidcPrefixes.UsernamePrefix = authServer.OIDCConfig.UsernamePrefix
	}
	if authServer.OIDCConfig.GroupsPrefix != "" {
		log.V(logger.LogLevelWarn).Info("OIDC groups prefix configured by both CLI and secret. Secret values will take precedence.")
		oidcPrefixes.GroupsPrefix = authServer.OIDCConfig.GroupsPrefix
	}

	cl, err := cluster.NewSingleCluster(cluster.DefaultCluster, rest, scheme, oidcPrefixes, cluster.DefaultKubeConfigOptions...)
	if err != nil {
		return fmt.Errorf("failed to create cluster client; %w", err)
	}

	if featureflags.Get("WEAVE_GITOPS_FEATURE_TELEMETRY") == "true" {
		err := telemetry.InitTelemetry(ctx, cl)
		if err != nil {
			// If there's an error turning on telemetry, that's not a
			// thing that should interrupt anything else
			log.Info("Couldn't enable telemetry", "error", err)
		}
	}

	log.Info("Using cached clients", "enabled", options.UseK8sCachedClients)

	if options.UseK8sCachedClients {
		cl = cluster.NewDelegatingCacheCluster(cl, rest, scheme)
	}

	fetcher := fetcher.NewSingleClusterFetcher(cl)

	clustersManager := clustersmngr.NewClustersManager([]clustersmngr.ClusterFetcher{fetcher}, nsaccess.NewChecker(nsaccess.DefautltWegoAppRules), log)
	clustersManager.Start(ctx)

	healthChecker := health.NewHealthChecker()

	coreConfig, err := core.NewCoreConfig(log, rest, clusterName, clustersManager, healthChecker)
	if err != nil {
		return fmt.Errorf("could not create core config: %w", err)
	}

	appAndProfilesHandlers, err := server.NewHandlers(ctx, log,
		&server.Config{
			CoreServerConfig: coreConfig,
			AuthServer:       authServer,
		},
		sessionManager,
	)
	if err != nil {
		return fmt.Errorf("could not create handler: %w", err)
	}

	mux.Handle("/v1/", gziphandler.GzipHandler(appAndProfilesHandlers))

	// Static asset handling
	assetFS := getAssets()
	assertFSHandler := http.FileServer(http.FS(assetFS))
	redirectHandler := server.IndexHTMLHandler(assetFS, log, options.RoutePrefix)
	assetHandler := server.AssetHandler(assertFSHandler, redirectHandler)
	mux.Handle("/", gziphandler.GzipHandler(assetHandler))

	if options.RoutePrefix != "" {
		mux = server.WithRoutePrefix(mux, options.RoutePrefix)
	}

	handler := http.Handler(mux)

	if options.EnableMetrics {
		mdlw := httpmiddleware.New(httpmiddleware.Config{
			Recorder: metrics.NewRecorder(metrics.Config{}),
		})
		handler = httpmiddlewarestd.Handler("", mdlw, handler)
	}

	handler = middleware.WithLogging(log, handler)

	handler = sessionManager.LoadAndSave(handler)

	addr := net.JoinHostPort(options.Host, options.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: handler,
	}

	go func() {
		log.Info("Starting server", "address", addr)

		if err := listenAndServe(log, srv, options); err != nil {
			log.Error(err, "server exited")
			os.Exit(1)
		}
	}()

	var metricsServer *http.Server

	if options.EnableMetrics {
		metricsMux := http.NewServeMux()
		gatherers := prometheus.Gatherers{
			prometheus.DefaultGatherer,
			k8sMetrics.Registry,
			clustersmngr.Registry,
		}
		metricsMux.Handle("/metrics", promhttp.HandlerFor(gatherers, promhttp.HandlerOpts{}))

		metricsServer = &http.Server{
			Addr:    options.MetricsAddress,
			Handler: metricsMux,
		}

		go func() {
			log.Info("Starting metrics endpoint", "address", metricsServer.Addr)

			if err := metricsServer.ListenAndServe(); err != nil {
				log.Error(err, "Error starting metrics endpoint, continuing anyway")
			}
		}()
	}

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)

	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	if options.EnableMetrics {
		if err := metricsServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("metrics server shutdown failed: %w", err)
		}
	}

	return nil
}

func listenAndServe(log logr.Logger, srv *http.Server, options Options) error {
	if options.Insecure {
		log.Info("TLS connections disabled")
		return srv.ListenAndServe()
	}

	if options.TLSCertFile == "" || options.TLSKeyFile == "" {
		return cmderrors.ErrNoTLSCertOrKey
	}

	if options.MTLS {
		caCert, err := os.ReadFile(options.TLSCertFile)
		if err != nil {
			return fmt.Errorf("failed reading cert file %s. %s", options.TLSCertFile, err)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		srv.TLSConfig = &tls.Config{
			ClientCAs:  caCertPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		}
	} else {
		log.Info("Using TLS", "cert_file", options.TLSCertFile, "key_file", options.TLSKeyFile)
	}

	// if tlsCert and tlsKey are both empty (""), ListenAndServeTLS will ignore
	// and happily use the TLSConfig supplied above
	return srv.ListenAndServeTLS(options.TLSCertFile, options.TLSKeyFile)
}

func getAssets() fs.FS {
	exec, err := os.Executable()
	if err != nil {
		panic(err)
	}

	f := os.DirFS(path.Join(path.Dir(exec), "dist"))

	return f
}
