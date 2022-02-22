package run

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/go-logr/zapr"
	"github.com/mattn/go-isatty"
	"github.com/pkg/browser"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/weaveworks/weave-gitops/cmd/gitops/cmderrors"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/cache"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/server/tls"
)

// Options contains all the options for the `ui run` command.
type Options struct {
	Port                          string
	Host                          string
	HelmRepoNamespace             string
	HelmRepoName                  string
	ProfileCacheLocation          string
	WatcherMetricsBindAddress     string
	WatcherHealthzBindAddress     string
	WatcherPort                   int
	Path                          string
	LoggingEnabled                bool
	OIDC                          OIDCAuthenticationOptions
	NotificationControllerAddress string
	TLSCert                       string
	TLSKey                        string
	NoTLS                         bool
}

// OIDCAuthenticationOptions contains the OIDC authentication options for the
// `ui run` command.
type OIDCAuthenticationOptions struct {
	IssuerURL     string
	ClientID      string
	ClientSecret  string
	RedirectURL   string
	TokenDuration time.Duration
}

var options Options

// NewCommand returns the `ui run` command
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "run [--log]",
		Short:   "Runs gitops ui",
		PreRunE: preRunCmd,
		RunE:    runCmd,
	}

	options = Options{}

	cmd.Flags().BoolVarP(&options.LoggingEnabled, "log", "l", false, "enable logging for the ui")
	cmd.Flags().StringVar(&options.Host, "host", server.DefaultHost, "UI host")
	cmd.Flags().StringVar(&options.Port, "port", server.DefaultPort, "UI port")
	cmd.Flags().StringVar(&options.Path, "path", "", "Path url")
	cmd.Flags().StringVar(&options.HelmRepoNamespace, "helm-repo-namespace", "default", "the namespace of the Helm Repository resource to scan for profiles")
	cmd.Flags().StringVar(&options.HelmRepoName, "helm-repo-name", "weaveworks-charts", "the name of the Helm Repository resource to scan for profiles")
	cmd.Flags().StringVar(&options.ProfileCacheLocation, "profile-cache-location", "/tmp/helm-cache", "the location where the cache Profile data lives")
	cmd.Flags().StringVar(&options.WatcherHealthzBindAddress, "watcher-healthz-bind-address", ":9981", "bind address for the healthz service of the watcher")
	cmd.Flags().StringVar(&options.WatcherMetricsBindAddress, "watcher-metrics-bind-address", ":9980", "bind address for the metrics service of the watcher")
	cmd.Flags().StringVar(&options.NotificationControllerAddress, "notification-controller-address", "", "the address of the notification-controller running in the cluster")
	cmd.Flags().IntVar(&options.WatcherPort, "watcher-port", 9443, "the port on which the watcher is running")

	cmd.Flags().StringVar(&options.TLSCert, "tls-cert-file", "", "filename for the TLS certficate, in-memory generated if omitted")
	cmd.Flags().StringVar(&options.TLSKey, "tls-private-key", "", "filename for the TLS key, in-memory generated if omitted")
	cmd.Flags().BoolVar(&options.NoTLS, "no-tls", false, "do not attempt to read TLS certificates")

	if server.AuthEnabled() {
		cmd.Flags().StringVar(&options.OIDC.IssuerURL, "oidc-issuer-url", "", "The URL of the OpenID Connect issuer")
		cmd.Flags().StringVar(&options.OIDC.ClientID, "oidc-client-id", "", "The client ID for the OpenID Connect client")
		cmd.Flags().StringVar(&options.OIDC.ClientSecret, "oidc-client-secret", "", "The client secret to use with OpenID Connect issuer")
		cmd.Flags().StringVar(&options.OIDC.RedirectURL, "oidc-redirect-url", "", "The OAuth2 redirect URL")
		cmd.Flags().DurationVar(&options.OIDC.TokenDuration, "oidc-token-duration", time.Hour, "The duration of the ID token. It should be set in the format: number + time unit (s,m,h) e.g., 20m")
	}

	return cmd
}

func preRunCmd(cmd *cobra.Command, args []string) error {
	issuerURL := options.OIDC.IssuerURL
	clientID := options.OIDC.ClientID
	clientSecret := options.OIDC.ClientSecret
	redirectURL := options.OIDC.RedirectURL

	if issuerURL != "" || clientID != "" || clientSecret != "" || redirectURL != "" {
		if issuerURL == "" {
			return cmderrors.ErrNoIssuerURL
		}

		if clientID == "" {
			return cmderrors.ErrNoClientID
		}

		if clientSecret == "" {
			return cmderrors.ErrNoClientSecret
		}

		if redirectURL == "" {
			return cmderrors.ErrNoRedirectURL
		}
	}

	return nil
}

func runCmd(cmd *cobra.Command, args []string) error {
	var log = logrus.New()

	mux := http.NewServeMux()

	mux.Handle("/health/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))

		if err != nil {
			log.Errorf("error writing health check: %s", err)
		}
	}))

	assetFS := getAssets()
	assetHandler := http.FileServer(http.FS(assetFS))
	redirector := createRedirector(assetFS, log)

	appConfig, err := server.DefaultApplicationsConfig()
	if err != nil {
		return fmt.Errorf("could not create http client: %w", err)
	}

	if !options.LoggingEnabled {
		appConfig.Logger = zapr.NewLogger(zap.NewNop())
	}

	rest, clusterName, err := kube.RestConfig()
	if err != nil {
		return fmt.Errorf("could not create client config: %w", err)
	}

	_, rawClient, err := kube.NewKubeHTTPClientWithConfig(rest, clusterName)
	if err != nil {
		return fmt.Errorf("could not create kube http client: %w", err)
	}

	profileCache, err := cache.NewCache(options.ProfileCacheLocation)
	if err != nil {
		return fmt.Errorf("failed to create cacher: %w", err)
	}

	if options.NotificationControllerAddress == "" {
		namespace, _ := cmd.Flags().GetString("namespace")
		options.NotificationControllerAddress = fmt.Sprintf("http://notification-controller.%s.svc.cluster.local./", namespace)
	}

	profileWatcher, err := watcher.NewWatcher(watcher.Options{
		KubeClient:                    rawClient,
		Cache:                         profileCache,
		MetricsBindAddress:            options.WatcherMetricsBindAddress,
		HealthzBindAddress:            options.WatcherHealthzBindAddress,
		NotificationControllerAddress: options.NotificationControllerAddress,
		WatcherPort:                   options.WatcherPort,
	})
	if err != nil {
		return fmt.Errorf("failed to start the watcher: %w", err)
	}

	go func() {
		if err := profileWatcher.StartWatcher(); err != nil {
			log.Error(err, "failed to start profile watcher")
			os.Exit(1)
		}
	}()

	profilesConfig := server.NewProfilesConfig(kube.ClusterConfig{
		DefaultConfig: rest,
		ClusterName:   clusterName,
	}, profileCache, options.HelmRepoNamespace, options.HelmRepoName)

	var authServer *auth.AuthServer

	if server.AuthEnabled() {
		_, err := url.Parse(options.OIDC.IssuerURL)
		if err != nil {
			return fmt.Errorf("invalid issuer URL: %w", err)
		}

		_, err = url.Parse(options.OIDC.RedirectURL)
		if err != nil {
			return fmt.Errorf("invalid redirect URL: %w", err)
		}

		tsv, err := auth.NewHMACTokenSignerVerifier(options.OIDC.TokenDuration)
		if err != nil {
			return fmt.Errorf("could not create HMAC token signer: %w", err)
		}

		srv, err := auth.NewAuthServer(cmd.Context(), appConfig.Logger, http.DefaultClient,
			auth.AuthConfig{
				OIDCConfig: auth.OIDCConfig{
					IssuerURL:     options.OIDC.IssuerURL,
					ClientID:      options.OIDC.ClientID,
					ClientSecret:  options.OIDC.ClientSecret,
					RedirectURL:   options.OIDC.RedirectURL,
					TokenDuration: options.OIDC.TokenDuration,
				},
			}, rawClient, tsv,
		)
		if err != nil {
			return fmt.Errorf("could not create auth server: %w", err)
		}

		appConfig.Logger.Info("Registering callback route")
		auth.RegisterAuthServer(mux, "/oauth2", srv)

		authServer = srv
	}

	appAndProfilesHandlers, err := server.NewHandlers(context.Background(), &server.Config{AppConfig: appConfig, ProfilesConfig: profilesConfig, AuthServer: authServer})
	if err != nil {
		return fmt.Errorf("could not create handler: %w", err)
	}

	mux.Handle("/v1/", appAndProfilesHandlers)

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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
	}))

	addr := net.JoinHostPort(options.Host, options.Port)

	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	go func() {
		log.Infof("Serving on %s", addr)

		if err := ListenAndServe(srv, options, log); err != nil {
			log.Error(err, "server exited")
			os.Exit(1)
		}
	}()

	if isatty.IsTerminal(os.Stdout.Fd()) {
		scheme := "https"
		if options.NoTLS {
			scheme = "http"
		}

		url := fmt.Sprintf("%s://%s/%s", scheme, addr, options.Path)

		log.Printf("Opening browser at %s", url)

		if err := browser.OpenURL(url); err != nil {
			return fmt.Errorf("failed to open the browser: %w", err)
		}
	}

	// graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

	defer func() {
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("Server Shutdown Failed: %w", err)
	}

	return nil
}

func ListenAndServe(srv *http.Server, options Options, log logrus.FieldLogger) error {
	if options.NoTLS {
		log.Info("TLS connections disabled")
		return srv.ListenAndServe()
	}

	if options.TLSCert == "" && options.TLSKey == "" {
		log.Info("TLS cert and key not specified, generating and using in-memory keys")

		tlsConfig, err := tls.TLSConfig([]string{"localhost", "0.0.0.0", "127.0.0.1"})
		if err != nil {
			return fmt.Errorf("failed to generate a TLSConfig: %w", err)
		}

		srv.TLSConfig = tlsConfig
		// if TLSCert and TLSKey are both empty (""), ListenAndServeTLS will ignore
		// and happily use the TLSConfig supplied above
		return srv.ListenAndServeTLS("", "")
	}

	if options.TLSCert == "" || options.TLSKey == "" {
		return cmderrors.ErrNoTLSCertOrKey
	}

	log.Infof("Using TLS from %q and %q", options.TLSCert, options.TLSKey)

	return srv.ListenAndServeTLS(options.TLSCert, options.TLSKey)
}

//go:embed dist/*
var static embed.FS

func getAssets() fs.FS {
	f, err := fs.Sub(static, "dist")

	if err != nil {
		panic(err)
	}

	return f
}

// A redirector ensures that index.html always gets served.
// The JS router will take care of actual navigation once the index.html page lands.
func createRedirector(fsys fs.FS, log logrus.FieldLogger) http.HandlerFunc {
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
