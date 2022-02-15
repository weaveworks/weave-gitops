package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-logr/logr"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	corev1 "k8s.io/api/core/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	LoginOIDC     string = "oidc"
	LoginUsername string = "username"
)

// OIDCConfig is used to configure an AuthServer to interact with
// an OIDC issuer.
type OIDCConfig struct {
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURL  string
}

// CookieConfig is used to configure the cookies that get issued
// from the OIDC issuer once the OAuth2 process flow completes.
type CookieConfig struct {
	CookieDuration     time.Duration
	IssueSecureCookies bool
}

// AuthConfig is used to configure an AuthServer.
type AuthConfig struct {
	OIDCConfig
	CookieConfig
}

// AuthServer interacts with an OIDC issuer to handle the OAuth2 process flow.
type AuthServer struct {
	logger              logr.Logger
	client              *http.Client
	provider            *oidc.Provider
	config              AuthConfig
	kubernetesClient    ctrlclient.Client
	tokenSignerVerifier TokenSignerVerifier
}

// Form data submitted by client
type LoginRequest struct {
	Password string `json:"password"`
}

type UserInfo struct {
	Email  string   `json:"email"`
	Groups []string `json:"groups"`
}

// NewAuthServer creates a new AuthServer object.
func NewAuthServer(ctx context.Context, logger logr.Logger, client *http.Client, config AuthConfig, kubernetesClient ctrlclient.Client, tokenSignerVerifier TokenSignerVerifier) (*AuthServer, error) {
	provider, err := oidc.NewProvider(ctx, config.IssuerURL)
	if err != nil {
		return nil, fmt.Errorf("could not create provider: %w", err)
	}

	hmacSecret := make([]byte, 64)

	_, err = rand.Read(hmacSecret)
	if err != nil {
		return nil, fmt.Errorf("could not generate random HMAC secret: %w", err)
	}

	return &AuthServer{
		logger:              logger,
		client:              client,
		provider:            provider,
		config:              config,
		kubernetesClient:    kubernetesClient,
		tokenSignerVerifier: tokenSignerVerifier,
	}, nil
}

// SetRedirectURL is used to set the redirect URL. This is meant to be used
// in unit tests only.
func (s *AuthServer) SetRedirectURL(url string) {
	s.config.RedirectURL = url
}

func (s *AuthServer) verifier() *oidc.IDTokenVerifier {
	return s.provider.Verifier(&oidc.Config{ClientID: s.config.ClientID})
}

func (s *AuthServer) oauth2Config(scopes []string) *oauth2.Config {
	// Ensure "openid" scope is always present.
	if !contains(scopes, oidc.ScopeOpenID) {
		scopes = append(scopes, oidc.ScopeOpenID)
	}

	// Request "offline_access" scope for refresh tokens.
	if !contains(scopes, oidc.ScopeOfflineAccess) {
		scopes = append(scopes, oidc.ScopeOfflineAccess)
	}

	// Request "email" scope to get user's email address.
	if !contains(scopes, scopeEmail) {
		scopes = append(scopes, scopeEmail)
	}

	// Request "groups" scope to get user's groups.
	if !contains(scopes, scopeGroups) {
		scopes = append(scopes, scopeGroups)
	}

	return &oauth2.Config{
		ClientID:     s.config.ClientID,
		ClientSecret: s.config.ClientSecret,
		Endpoint:     s.provider.Endpoint(),
		RedirectURL:  s.config.RedirectURL,
		Scopes:       scopes,
	}
}

func (s *AuthServer) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	var (
		token *oauth2.Token
		state SessionState
	)

	ctx := oidc.ClientContext(r.Context(), s.client)

	switch r.Method {
	case http.MethodGet:
		// Authorization redirect callback from OAuth2 auth flow.
		if errMsg := r.FormValue("error"); errMsg != "" {
			s.logger.Info("authz redirect callback failed", "error", errMsg, "error_description", r.FormValue("error_description"))
			http.Error(rw, "", http.StatusBadRequest)

			return
		}

		code := r.FormValue("code")
		if code == "" {
			s.logger.Info("code value was empty")
			http.Error(rw, "", http.StatusBadRequest)

			return
		}

		cookie, err := r.Cookie(StateCookieName)
		if err != nil {
			s.logger.Error(err, "cookie was not found in the request", "cookie", StateCookieName)
			http.Error(rw, "", http.StatusBadRequest)

			return
		}

		if state := r.FormValue("state"); state != cookie.Value {
			s.logger.Info("cookie value does not match state value")
			http.Error(rw, "", http.StatusBadRequest)

			return
		}

		b, err := base64.StdEncoding.DecodeString(cookie.Value)
		if err != nil {
			s.logger.Error(err, "cannot base64 decode cookie", "cookie", StateCookieName, "cookie_value", cookie.Value)
			http.Error(rw, "", http.StatusInternalServerError)

			return
		}

		if err := json.Unmarshal(b, &state); err != nil {
			s.logger.Error(err, "failed to unmarshal state to JSON")
			http.Error(rw, "", http.StatusInternalServerError)

			return
		}

		token, err = s.oauth2Config(nil).Exchange(ctx, code)
		if err != nil {
			s.logger.Error(err, "failed to exchange auth code for token")
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

	_, err := s.verifier().Verify(r.Context(), rawIDToken)
	if err != nil {
		http.Error(rw, fmt.Sprintf("failed to verify ID token: %v", err), http.StatusInternalServerError)
		return
	}

	// Issue ID token cookie
	http.SetCookie(rw, s.createCookie(IDTokenCookieName, rawIDToken))

	// Some OIDC providers may not include a refresh token
	if token.RefreshToken != "" {
		// Issue refresh token cookie
		http.SetCookie(rw, s.createCookie(RefreshTokenCookieName, token.RefreshToken))
	}

	// Clear state cookie
	http.SetCookie(rw, s.clearCookie(StateCookieName))

	http.Redirect(rw, r, state.ReturnURL, http.StatusSeeOther)
}

func (s *AuthServer) SignIn() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			s.logger.Info("Only POST requests allowed")
			rw.WriteHeader(http.StatusMethodNotAllowed)

			return
		}

		var loginRequest LoginRequest

		err := json.NewDecoder(r.Body).Decode(&loginRequest)
		if err != nil {
			s.logger.Error(err, "Failed to decode from JSON")
			http.Error(rw, fmt.Sprintf("Failed to read request body."), http.StatusBadRequest)

			return
		}

		var hashedSecret corev1.Secret

		if err := s.kubernetesClient.Get(r.Context(), ctrlclient.ObjectKey{
			Namespace: "wego-system",
			Name:      "admin-password-hash",
		}, &hashedSecret); err != nil {
			s.logger.Error(err, "Failed to query for the secret")
			http.Error(rw, fmt.Sprint("Please ensure that a password has been set."), http.StatusBadRequest)

			return
		}

		if err := bcrypt.CompareHashAndPassword(hashedSecret.Data["password"], []byte(loginRequest.Password)); err != nil {
			s.logger.Error(err, "Failed to compare hash with password")
			rw.WriteHeader(http.StatusUnauthorized)

			return
		}

		signed, err := s.tokenSignerVerifier.Sign()
		if err != nil {
			s.logger.Error(err, "Failed to create and sign token")
			rw.WriteHeader(http.StatusInternalServerError)

			return
		}

		http.SetCookie(rw, s.createCookie(IDTokenCookieName, signed))
		rw.WriteHeader(http.StatusOK)
	}
}

// Call the OIDC User Info endpoint and return information to the client
// Read the cookie and extract the token to then interogate the OIDC provider for the User Info
func (s *AuthServer) UserInfo() http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(IDTokenCookieName)
		if err != nil {
			http.Error(rw, fmt.Sprintf("failed to read cookie: %v", err), http.StatusBadRequest)
			return
		}

		claims, err := s.tokenSignerVerifier.Verify(c.Value)
		if err != nil {
			s.logger.Error(err, "Failed to verify token", "token", c.Value)
		} else {
			ui := UserInfo{
				Email: claims.Subject,
			}
			toJson(ui, rw)
			return
		}

		info, err := s.provider.UserInfo(r.Context(), oauth2.StaticTokenSource(&oauth2.Token{
			AccessToken: c.Value,
		}))
		if err != nil {
			http.Error(rw, fmt.Sprintf("failed to query userinfo endpoint: %v", err), http.StatusUnauthorized)
			return
		}

		ui := UserInfo{
			Email: info.Email,
		}

		toJson(ui, rw)
	}
}

func toJson(ui UserInfo, rw http.ResponseWriter) {
	b, err := json.Marshal(ui)
	if err != nil {
		http.Error(rw, fmt.Sprintf("failed to marshal to JSON: %v", err), http.StatusInternalServerError)
		return
	}

	rw.Write(b)
}

func (c *AuthServer) startAuthFlow(rw http.ResponseWriter, r *http.Request) {
	nonce, err := generateNonce()
	if err != nil {
		http.Error(rw, fmt.Sprintf("failed to generate nonce: %v", err), http.StatusInternalServerError)
		return
	}

	returnUrl := r.URL.Query().Get("return_url")

	if returnUrl == "" {
		returnUrl = r.URL.String()
	}

	fmt.Printf("return url:%v\n", returnUrl)

	b, err := json.Marshal(SessionState{
		Nonce:     nonce,
		ReturnURL: returnUrl,
	})
	if err != nil {
		http.Error(rw, fmt.Sprintf("failed to marshal state to JSON: %v", err), http.StatusInternalServerError)
		return
	}

	state := base64.StdEncoding.EncodeToString(b)

	var scopes []string
	// "openid", "offline_access", "email" and "groups" scopes added by default
	scopes = append(scopes, scopeProfile)
	authCodeUrl := c.oauth2Config(scopes).AuthCodeURL(state)

	// Issue state cookie
	http.SetCookie(rw, c.createCookie(StateCookieName, state))

	http.Redirect(rw, r, authCodeUrl, http.StatusSeeOther)
}

func (c *AuthServer) createCookie(name, value string) *http.Cookie {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().UTC().Add(c.config.CookieDuration),
		HttpOnly: true,
	}

	if c.config.IssueSecureCookies {
		cookie.Secure = true
	}

	return cookie
}

func (c *AuthServer) clearCookie(name string) *http.Cookie {
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
