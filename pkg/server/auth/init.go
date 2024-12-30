package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/alexedwards/scs/v2"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/weaveworks/weave-gitops/core/logger"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
)

// AuthParams provides the configuration for the AuthServer.
type AuthParams struct {
	OIDCConfig        OIDCConfig
	OIDCSecretName    string
	AuthMethodStrings []string
	NoAuthUser        string
	Namespace         string
	SessionManager    *scs.SessionManager
}

// InitAuthServer creates a new AuthServer and configures it for the correct
// authentication methods.
func InitAuthServer(ctx context.Context, log logr.Logger, rawKubernetesClient ctrlclient.Client, authParams AuthParams) (*AuthServer, error) {
	log.V(logger.LogLevelDebug).Info("Parsing authentication methods", "methods", authParams.AuthMethodStrings)
	authMethods, err := ParseAuthMethodArray(authParams.AuthMethodStrings)
	if err != nil {
		return nil, err
	}

	if authParams.NoAuthUser != "" {
		log.V(logger.LogLevelWarn).Info("Anonymous mode enabled", "noAuthUser", authParams.NoAuthUser)
		authMethods = map[AuthMethod]bool{Anonymous: true}
	}

	log.Info("Registering authentication methods", "methods", toAuthMethodStrings(authMethods))

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication methods set")
	}

	oidcConfig := authParams.OIDCConfig
	if authMethods[OIDC] {
		if authParams.OIDCSecretName != DefaultOIDCAuthSecretName {
			log.V(logger.LogLevelDebug).Info("Reading OIDC configuration from alternate secret",
				"name", authParams.OIDCSecretName,
				"namespace", authParams.Namespace,
			)
		}

		// If OIDC auth secret is found prefer that over CLI parameters
		var secret corev1.Secret
		if err := rawKubernetesClient.Get(ctx, types.NamespacedName{
			Name:      authParams.OIDCSecretName,
			Namespace: authParams.Namespace,
		}, &secret); err == nil {
			if oidcConfig.ClientSecret != "" && secret.Data["clientSecret"] != nil { // 'Data' is a byte array
				log.V(logger.LogLevelWarn).Info("OIDC client configured by both CLI and secret. CLI values will be overridden.")
			}

			oidcConfig = NewOIDCConfigFromSecret(secret)
		} else if err != nil {
			log.V(logger.LogLevelDebug).Info("Could not read OIDC secret",
				"name", authParams.OIDCSecretName,
				"namespace", authParams.Namespace,
				"error", err,
			)
		}

		if oidcConfig.ClientSecret != "" {
			log.V(logger.LogLevelDebug).Info("OIDC config",
				"IssuerURL", oidcConfig.IssuerURL,
				"ClientID", oidcConfig.ClientID,
				"ClientSecretLength", len(oidcConfig.ClientSecret),
				"RedirectURL", oidcConfig.RedirectURL,
				"TokenDuration", oidcConfig.TokenDuration,
			)
		}

		if _, err := url.Parse(oidcConfig.IssuerURL); err != nil {
			return nil, fmt.Errorf("invalid issuer URL: %w", err)
		}

		if _, err := url.Parse(oidcConfig.RedirectURL); err != nil {
			return nil, fmt.Errorf("invalid redirect URL: %w", err)
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

	if featureflags.IsSet("WEAVE_GITOPS_FEATURE_DEV_MODE") {
		log.V(logger.LogLevelWarn).Info("Dev-mode is enabled. This should be used for local work only.")
		tsv.SetDevMode(true)
	}

	authServer, err := NewAuthServer(ctx, &AuthServerConfig{
		Log:                 log.WithName("auth-server"),
		client:              http.DefaultClient,
		kubernetesClient:    rawKubernetesClient,
		tokenSignerVerifier: tsv,
		authMethods:         authMethods,
		OIDCConfig:          oidcConfig,
		namespace:           authParams.Namespace,
		noAuthUser:          authParams.NoAuthUser,
		SessionManager:      authParams.SessionManager,
	})
	if err != nil {
		return nil, fmt.Errorf("could not create auth server: %w", err)
	}

	authParams.SessionManager.Lifetime = oidcConfig.TokenDuration

	return authServer, err
}
