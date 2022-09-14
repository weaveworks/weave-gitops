package auth

import (
	"context"
	"fmt"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/logger"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
	corev1 "k8s.io/api/core/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

func InitAuthServer(ctx context.Context, log logr.Logger, rawKubernetesClient ctrlclient.Client, oidcConfig OIDCConfig, oidcSecret string, namespace string, authMethodStrings []string) (*AuthServer, error) {
	log.V(logger.LogLevelDebug).Info("Registering authentication methods", "methods", authMethodStrings)

	authMethods, err := ParseAuthMethodArray(authMethodStrings)
	if err != nil {
		return nil, err
	}

	if len(authMethods) == 0 {
		return nil, fmt.Errorf("no authentication methods set")
	}

	if authMethods[OIDC] {
		if oidcSecret != DefaultOIDCAuthSecretName {
			log.V(logger.LogLevelDebug).Info("Reading OIDC configuration from alternate secret", "secretName", oidcSecret)
		}

		// If OIDC auth secret is found prefer that over CLI parameters
		var secret corev1.Secret
		if err := rawKubernetesClient.Get(ctx, ctrlclient.ObjectKey{
			Namespace: namespace,
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

	if featureflags.Get("WEAVE_GITOPS_FEATURE_DEV_MODE") == "true" {
		log.V(logger.LogLevelWarn).Info("Dev-mode is enabled. This should be used for local work only.")
		tsv.SetDevMode(true)
	}

	authCfg, err := NewAuthServerConfig(log, oidcConfig, rawKubernetesClient, tsv, namespace, authMethods)
	if err != nil {
		return nil, err
	}

	authServer, err := NewAuthServer(ctx, authCfg)
	if err != nil {
		return nil, fmt.Errorf("could not create auth server: %w", err)
	}

	return authServer, err
}
