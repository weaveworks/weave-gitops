package cmd

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/fs"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"syscall"
	"time"

	"github.com/NYTimes/gziphandler"
	"github.com/go-logr/logr"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	metrics "github.com/slok/go-http-metrics/metrics/prometheus"
	httpmiddleware "github.com/slok/go-http-metrics/middleware"
	httpmiddlewarestd "github.com/slok/go-http-metrics/middleware/std"
	"github.com/spf13/cobra"
	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/fetcher"
	"github.com/weaveworks/weave-gitops/core/logger"
	"github.com/weaveworks/weave-gitops/core/nsaccess"
	core "github.com/weaveworks/weave-gitops/core/server"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

const (
	// Allowed login requests per second
	loginRequestRateLimit = 20
)

// Options contains all the options for the gitops-server command.
type Options struct {
	// System config
	Host                          string
	LogLevel                      string
	NotificationControllerAddress string
	Path                          string
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
	// Dev mode
	DevMode bool
	// Metrics
	EnableMetrics  bool
	MetricsAddress string
}

var options Options

func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Short: "Runs the gitops-server",
		RunE:  runCmd,
	}

	options = Options{}

	// System config
	cmd.Flags().StringVar(&options.Host, "host", server.DefaultHost, "UI host")
	cmd.Flags().StringVar(&options.LogLevel, "log-level", logger.DefaultLogLevel, "log level")
	cmd.Flags().StringVar(&options.NotificationControllerAddress, "notification-controller-address", "", "the address of the notification-controller running in the cluster")
	cmd.Flags().StringVar(&options.Path, "path", "", "Path url")
	cmd.Flags().StringVar(&options.Port, "port", server.DefaultPort, "UI port")
	cmd.Flags().StringSliceVar(&options.AuthMethods, "auth-methods", auth.DefaultAuthMethodStrings(), "Which auth methods to use, valid values are 'oidc' and 'user-account'")
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
	// Dev mode
	cmd.Flags().BoolVar(&options.DevMode, "dev-mode", false, "Enables development mode - this disables verifying cookie signatures, do not use")
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

	mux := http.NewServeMux()

	mux.Handle("/health/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))

		if err != nil {
			log.Error(err, "error writing health check")
		}
	}))

	assetFS := getAssets()
	assetHandler := http.FileServer(http.FS(assetFS))
	redirector := createRedirector(assetFS, log)
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
		return fmt.Errorf("Couldn't get current namespace")
	}

	authServer, err := auth.InitAuthServer(cmd.Context(), log, rawClient, options.OIDC, options.OIDCSecret, options.DevMode, namespace, options.AuthMethods)

	if err != nil {
		return fmt.Errorf("could not initialise authentication server: %w", err)
	}

	log.Info("Registering auth routes")

	if err := auth.RegisterAuthServer(mux, "/oauth2", authServer, loginRequestRateLimit); err != nil {
		return fmt.Errorf("failed to register auth routes: %w", err)
	}

	ctx := context.Background()

	fetcher := fetcher.NewSingleClusterFetcher(rest)

	clusterClientsFactory := clustersmngr.NewClientFactory(fetcher, nsaccess.NewChecker(nsaccess.DefautltWegoAppRules), log, scheme, clustersmngr.NewClustersClientsPool)
	clusterClientsFactory.Start(ctx)

	coreConfig := core.NewCoreConfig(log, rest, clusterName, clusterClientsFactory)

	appConfig, err := server.DefaultApplicationsConfig(log)
	if err != nil {
		return fmt.Errorf("could not create http client: %w", err)
	}

	appAndProfilesHandlers, err := server.NewHandlers(ctx, log,
		&server.Config{
			AppConfig:        appConfig,
			CoreServerConfig: coreConfig,
			AuthServer:       authServer,
		},
	)
	if err != nil {
		return fmt.Errorf("could not create handler: %w", err)
	}

	mux.Handle("/v1/", gziphandler.GzipHandler(appAndProfilesHandlers))

	mux.Handle("/", gziphandler.GzipHandler(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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
	})))

	handler := http.Handler(mux)

	if options.EnableMetrics {
		mdlw := httpmiddleware.New(httpmiddleware.Config{
			Recorder: metrics.NewRecorder(metrics.Config{}),
		})
		handler = httpmiddlewarestd.Handler("", mdlw, mux)
	}

	handler = middleware.WithLogging(log, handler)

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
		metricsMux.Handle("/metrics", promhttp.Handler())

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

	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("Server shutdown failed: %w", err)
	}

	if options.EnableMetrics {
		if err := metricsServer.Shutdown(ctx); err != nil {
			return fmt.Errorf("Metrics server shutdown failed: %w", err)
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
		caCert, err := ioutil.ReadFile(options.TLSCertFile)
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

// A redirector ensures that index.html always gets served.
// The JS router will take care of actual navigation once the index.html page lands.
func createRedirector(fsys fs.FS, log logr.Logger) http.HandlerFunc {
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
