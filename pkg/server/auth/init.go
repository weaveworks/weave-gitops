package auth

import (
  "context"
  "fmt"
  "net/http"
  "net/url"
  "time"

  "github.com/coreos/go-oidc/v3/oidc"
  "github.com/go-logr/logr"
  "github.com/weaveworks/weave-gitops/api/v1alpha1"
  "github.com/weaveworks/weave-gitops/core/logger"
  "github.com/weaveworks/weave-gitops/pkg/featureflags"
  corev1 "k8s.io/api/core/v1"
  "sigs.k8s.io/controller-runtime/pkg/client"
  ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func InitAuthServer(ctx context.Context, log logr.Logger, rawKubernetesClient ctrlclient.Client, oidcConfig OIDCConfig, oidcSecret, devUser string, devMode bool, authMethodStrings []string) (*AuthServer, error) {
  log.V(logger.LogLevelDebug).Info("Registering authentication methods", "methods", authMethodStrings)

  authMethods, err := ParseAuthMethodArray(authMethodStrings)
  if err != nil {
    return nil, err
  }

  if len(authMethods) == 0 {
    return nil, fmt.Errorf("No authentication methods set")
  }

  if authMethods[OIDC] {
    if oidcSecret != DefaultOIDCAuthSecretName {
      log.V(logger.LogLevelDebug).Info("Reading OIDC configuration from alternate secret", "secretName", oidcSecret)
    }

    // If OIDC auth secret is found prefer that over CLI parameters
    var secret corev1.Secret
    if err := rawKubernetesClient.Get(ctx, client.ObjectKey{
      Namespace: v1alpha1.DefaultNamespace,
      Name:      oidcSecret,
    }, &secret); err == nil {
      if oidcConfig.ClientSecret != "" && secret.Data["clientSecret"] != nil { // 'Data' is a byte array
        log.V(logger.LogLevelWarn).Info("OIDC client configured by both CLI and secret. CLI values will be overridden.")
      }

      oidcConfig = NewOIDCConfigFromSecret(secret)
    } else if err != nil {
      log.V(logger.LogLevelDebug).Info("Could not read OIDC secret", "secretName", oidcSecret, "error", err)
    }

    if oidcConfig.ClientSecret != "" {
      log.V(logger.LogLevelDebug).Info("OIDC config", "IssuerURL", oidcConfig.IssuerURL, "ClientID", oidcConfig.ClientID, "ClientSecretLength", len(oidcConfig.ClientSecret), "RedirectURL", oidcConfig.RedirectURL, "TokenDuration", oidcConfig.TokenDuration)
    }
  } else {
    // Make sure there is no OIDC config if it's not an enabled authorization method
    // the TokenDuration needs to be set so cookies can use it
    oidcConfig = OIDCConfig{TokenDuration: defaultCookieDuration}
  }

  tsv, err := NewHMACTokenSignerVerifier(oidcConfig.TokenDuration)
  if err != nil {
    return nil, fmt.Errorf("could not create HMAC token signer: %w", err)
  }

  if devMode {
    log.V(logger.LogLevelWarn).Info("Dev-mode is enabled. This should be used for local work only.")
    tsv.SetDevMode(devUser)
  }

  authCfg, err := NewAuthServerConfig(log, oidcConfig, rawKubernetesClient, tsv, authMethods)
  if err != nil {
    return nil, err
  }

  authServer, err := NewAuthServer(ctx, authCfg)
  if err != nil {
    return nil, fmt.Errorf("could not create auth server: %w", err)
  }
  return authServer, err
}

func NewOIDCConfigFromSecret(secret corev1.Secret) OIDCConfig {
  cfg := OIDCConfig{
    IssuerURL:    string(secret.Data["issuerURL"]),
    ClientID:     string(secret.Data["clientID"]),
    ClientSecret: string(secret.Data["clientSecret"]),
    RedirectURL:  string(secret.Data["redirectURL"]),
  }

  tokenDuration, err := time.ParseDuration(string(secret.Data["tokenDuration"]))
  if err != nil {
    tokenDuration = defaultCookieDuration
  }

  cfg.TokenDuration = tokenDuration

  return cfg
}

func NewAuthServerConfig(log logr.Logger, oidcCfg OIDCConfig, kubernetesClient ctrlclient.Client, tsv TokenSignerVerifier, authMethods map[AuthMethod]bool) (AuthConfig, error) {
  if authMethods[OIDC] {
    if _, err := url.Parse(oidcCfg.IssuerURL); err != nil {
      return AuthConfig{}, fmt.Errorf("invalid issuer URL: %w", err)
    }

    if _, err := url.Parse(oidcCfg.RedirectURL); err != nil {
      return AuthConfig{}, fmt.Errorf("invalid redirect URL: %w", err)
    }
  }

  return AuthConfig{
    Log:                 log.WithName("auth-server"),
    client:              http.DefaultClient,
    kubernetesClient:    kubernetesClient,
    tokenSignerVerifier: tsv,
    config:              oidcCfg,
    authMethods:         authMethods,
  }, nil
}

// NewAuthServer creates a new AuthServer object.
func NewAuthServer(ctx context.Context, cfg AuthConfig) (*AuthServer, error) {
  if cfg.authMethods[UserAccount] {
    var secret corev1.Secret
    err := cfg.kubernetesClient.Get(ctx, client.ObjectKey{
      Namespace: v1alpha1.DefaultNamespace,
      Name:      ClusterUserAuthSecretName,
    }, &secret)

    if err != nil {
      return nil, fmt.Errorf("Could not get secret for cluster user, %w", err)
    } else {
      featureflags.Set(FeatureFlagClusterUser, FeatureFlagSet)
    }
  } else {
    featureflags.Set(FeatureFlagClusterUser, "false")
  }

  var provider *oidc.Provider

  if cfg.config.IssuerURL == "" {
    featureflags.Set(FeatureFlagOIDCAuth, "false")
  } else if cfg.authMethods[OIDC] {
    var err error

    provider, err = oidc.NewProvider(ctx, cfg.config.IssuerURL)
    if err != nil {
      return nil, fmt.Errorf("could not create provider: %w", err)
    }
    featureflags.Set(FeatureFlagOIDCAuth, FeatureFlagSet)
  }

  if featureflags.Get(FeatureFlagOIDCAuth) != FeatureFlagSet && featureflags.Get(FeatureFlagClusterUser) != FeatureFlagSet {
    return nil, fmt.Errorf("Neither OIDC auth or local auth enabled, can't start")
  }

  return &AuthServer{cfg, provider}, nil
}
