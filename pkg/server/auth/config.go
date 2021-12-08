package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-logr/logr"
	"golang.org/x/oauth2"
)

// AuthConfig holds auth configuration parameters.
type AuthConfig struct {
	logger             logr.Logger
	provider           *oidc.Provider
	clientID           string
	clientSecret       string
	redirectURL        string
	issueSecureCookies bool
	cookieDuration     time.Duration
	client             *http.Client
}

// NewAuthConfig creates a new AuthConfig object.
func NewAuthConfig(ctx context.Context, oidcIssuerURL, oidcClientId, oidcClientSecret, oidcRedirectURL string, oidcIssueSecureCookies bool, oidcCookieDuration time.Duration, client *http.Client, logger logr.Logger) (*AuthConfig, error) {
	provider, err := oidc.NewProvider(ctx, oidcIssuerURL)
	if err != nil {
		return nil, fmt.Errorf("could not create provider: %w", err)
	}

	return &AuthConfig{
		logger:             logger,
		provider:           provider,
		clientID:           oidcClientId,
		clientSecret:       oidcClientSecret,
		redirectURL:        oidcRedirectURL,
		issueSecureCookies: oidcIssueSecureCookies,
		cookieDuration:     oidcCookieDuration,
		client:             client,
	}, nil
}

// Logger returns the logger instance
func (c *AuthConfig) Logger() logr.Logger {
	return c.logger
}

// SetRedirectURL is used to set the redirect URL. This is meant to be used
// in unit tests only.
func (c *AuthConfig) SetRedirectURL(url string) {
	c.redirectURL = url
}

func (c *AuthConfig) verifier() *oidc.IDTokenVerifier {
	return c.provider.Verifier(&oidc.Config{ClientID: c.clientID})
}

func (c *AuthConfig) oauth2Config(scopes []string) *oauth2.Config {
	// Ensure "openid" scope is always present.
	if !contains(scopes, oidc.ScopeOpenID) {
		scopes = append(scopes, oidc.ScopeOpenID)
	}

	// Request "offline_access" scope for refresh tokens.
	if !contains(scopes, oidc.ScopeOfflineAccess) {
		scopes = append(scopes, oidc.ScopeOfflineAccess)
	}

	// Request "email" scope to get user's email address.
	if !contains(scopes, ScopeEmail) {
		scopes = append(scopes, ScopeEmail)
	}

	return &oauth2.Config{
		ClientID:     c.clientID,
		ClientSecret: c.clientSecret,
		Endpoint:     c.provider.Endpoint(),
		RedirectURL:  c.redirectURL,
		Scopes:       scopes,
	}
}

func (c *AuthConfig) callback() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		var (
			token *oauth2.Token
			state SessionState
		)

		ctx := oidc.ClientContext(r.Context(), c.client)

		switch r.Method {
		case http.MethodGet:
			// Authorization redirect callback from OAuth2 auth flow.
			if errMsg := r.FormValue("error"); errMsg != "" {
				c.logger.Info("authz redirect callback failed", "error", errMsg, "error_description", r.FormValue("error_description"))
				http.Error(rw, "", http.StatusBadRequest)

				return
			}

			code := r.FormValue("code")
			if code == "" {
				c.logger.Info("code value was empty")
				http.Error(rw, "", http.StatusBadRequest)

				return
			}

			cookie, err := r.Cookie(StateCookieName)
			if err != nil {
				c.logger.Error(err, "cookie was not found in the request", "cookie", StateCookieName)
				http.Error(rw, "", http.StatusBadRequest)

				return
			}

			if state := r.FormValue("state"); state != cookie.Value {
				c.logger.Info("cookie value does not match state value")
				http.Error(rw, "", http.StatusBadRequest)

				return
			}

			b, err := base64.StdEncoding.DecodeString(cookie.Value)
			if err != nil {
				c.logger.Error(err, "cannot base64 decode cookie", "cookie", StateCookieName, "cookie_value", cookie.Value)
				http.Error(rw, "", http.StatusInternalServerError)

				return
			}

			if err := json.Unmarshal(b, &state); err != nil {
				c.logger.Error(err, "failed to unmarshal state to JSON")
				http.Error(rw, "", http.StatusInternalServerError)

				return
			}

			token, err = c.oauth2Config(nil).Exchange(ctx, code)
			if err != nil {
				c.logger.Error(err, "failed to exchange auth code for token")
				http.Error(rw, "", http.StatusInternalServerError)

				return
			}
		default:
			http.Error(rw, fmt.Sprintf("method not implemented: %s", r.Method), http.StatusBadRequest)

			return
		}

		rawIDToken, ok := token.Extra("id_token").(string)
		if !ok {
			http.Error(rw, "no id_token in token response", http.StatusInternalServerError)
			return
		}

		_, err := c.verifier().Verify(r.Context(), rawIDToken)
		if err != nil {
			http.Error(rw, fmt.Sprintf("failed to verify ID token: %v", err), http.StatusInternalServerError)
			return
		}

		// Issue ID token cookie
		http.SetCookie(rw, c.createCookie(IDTokenCookieName, rawIDToken))

		// Some OIDC providers may not include a refresh token
		if token.RefreshToken != "" {
			// Issue refresh token cookie
			http.SetCookie(rw, c.createCookie(RefreshTokenCookieName, token.RefreshToken))
		}

		// Clear state cookie
		http.SetCookie(rw, c.clearCookie(StateCookieName))

		http.Redirect(rw, r, state.ReturnURL, http.StatusSeeOther)
	}
}

func (c *AuthConfig) createCookie(name, value string) *http.Cookie {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().UTC().Add(c.cookieDuration),
		HttpOnly: true,
	}

	if c.issueSecureCookies {
		cookie.Secure = true
	}

	return cookie
}

func (c *AuthConfig) clearCookie(name string) *http.Cookie {
	cookie := &http.Cookie{
		Name:    name,
		Value:   "",
		Path:    "/",
		Expires: time.Unix(0, 0),
	}

	return cookie
}

// SessionState represents the state that needs to be persisted between
// the AuthN request from the Relying Party (RP) to the authorization
// endpoint of the OpenID Provider (OP) and the AuthN response back from
// the OP to the RP's callback URL. This state could be persisted server-side
// in a data store such as Redis but we prefer to operate stateless so we
// store this in a cookie instead. The cookie value and the value of the
// "state" parameter passed in the AuthN request are identical and set to
// the base64-encoded, JSON serialised state.
//
// https://openid.net/specs/openid-connect-core-1_0.html#Overview
// https://auth0.com/docs/configure/attack-protection/state-parameters#alternate-redirect-method
// https://community.auth0.com/t/state-parameter-and-user-redirection/8387/2
type SessionState struct {
	Nonce     string `json:"n"`
	ReturnURL string `json:"return_url"`
}

func contains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}

	return false
}
