package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	// StateCookieName is the name of the cookie that holds state during auth flow.
	StateCookieName = "state"
	// IDTokenCookieName is the name of the cookie that holds the ID Token once
	// the user has authenticated successfully with the OIDC Provider.
	IDTokenCookieName = "id_token"
	// RefreshTokenCookieName is the name of the cookie that holds the refresh
	// token.
	RefreshTokenCookieName = "refresh_token"
	// ScopeProfile is the "profile" scope
	ScopeProfile = "profile"
	// ScopeEmail is the "email" scope
	ScopeEmail = "email"
)

// RegisterAuthHandler registers the /callback route under a specified prefix.
// This route is called by the OIDC Provider in order to pass back state after
// the authentication flow completes.
func RegisterAuthHandler(mux *http.ServeMux, prefix string, cfg *AuthConfig) {
	mux.Handle(prefix+"/callback", cfg.callback())
}

type principalCtxKey struct{}

// UserPrincipal is a simple model for the user, including their ID and Groups.
type UserPrincipal struct {
	ID     string   `json:"id"`
	Groups []string `json:"groups"`
}

// WithPrincipal sets the principal into the context.
func WithPrincipal(ctx context.Context, p *UserPrincipal) context.Context {
	return context.WithValue(ctx, principalCtxKey{}, p)
}

// WithAPIAuth middleware adds auth validation to API handlers.
//
// Unauthorized requests will be denied with a 401 status code.
func WithAPIAuth(next http.Handler, cfg *AuthConfig) http.Handler {
	cookieAuth := NewJWTCookiePrincipalGetter(cfg.logger,
		cfg.verifier(), IDTokenCookieName)
	headerAuth := NewJWTAuthorizationHeaderPrincipalGetter(cfg.logger, cfg.verifier())

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		principal, err := MultiAuthPrincipal{cookieAuth, headerAuth}.Principal(r)
		if err != nil || principal == nil {
			cfg.logger.Error(err, "failed to get principal")

			rw.Header().Set("WWW-Authenticate", `Bearer realm="Weave GitOps"`)
			http.Error(rw, "Authentication required", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(rw, r.Clone(WithPrincipal(r.Context(), principal)))
	})
}

// WithWebAuth middleware adds auth validation to HTML handlers.
//
// Unauthorized requests will be redirected to the OIDC Provider.
// It is meant to be used with routes that serve HTML content,
// not API routes.
func WithWebAuth(next http.Handler, cfg *AuthConfig) http.Handler {
	cookieAuth := NewJWTCookiePrincipalGetter(cfg.logger,
		cfg.verifier(), IDTokenCookieName)
	headerAuth := NewJWTAuthorizationHeaderPrincipalGetter(cfg.logger, cfg.verifier())

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		principal, err := MultiAuthPrincipal{cookieAuth, headerAuth}.Principal(r)
		if err != nil || principal == nil {
			cfg.logger.Error(err, "failed to get principal")

			startAuthFlow(rw, r, cfg)
			return
		}

		next.ServeHTTP(rw, r.Clone(WithPrincipal(r.Context(), principal)))
	})
}

func startAuthFlow(rw http.ResponseWriter, r *http.Request, cfg *AuthConfig) {
	nonce, err := generateNonce(32)
	if err != nil {
		http.Error(rw, fmt.Sprintf("failed to generate nonce: %v", err), http.StatusInternalServerError)
		return
	}

	b, err := json.Marshal(SessionState{
		Nonce:     nonce,
		ReturnURL: r.URL.String(),
	})
	if err != nil {
		http.Error(rw, fmt.Sprintf("failed to marshal state to JSON: %v", err), http.StatusInternalServerError)
		return
	}

	state := base64.StdEncoding.EncodeToString(b)

	var scopes []string
	// "openid", "offline_access" and "email" scopes added by default
	scopes = append(scopes, ScopeProfile)
	authCodeUrl := cfg.oauth2Config(scopes).AuthCodeURL(state)

	// Issue state cookie
	http.SetCookie(rw, cfg.createCookie(StateCookieName, state))

	http.Redirect(rw, r, authCodeUrl, http.StatusSeeOther)
}

func generateNonce(n int) (string, error) {
	b := make([]byte, n)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}
