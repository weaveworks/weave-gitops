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
	scopeProfile = "profile"
	// ScopeEmail is the "email" scope
	scopeEmail = "email"
)

// RegisterAuthServer registers the /callback route under a specified prefix.
// This route is called by the OIDC Provider in order to pass back state after
// the authentication flow completes.
func RegisterAuthServer(mux *http.ServeMux, prefix string, srv *AuthServer) {
	mux.Handle(prefix+"/callback", srv)
}

type principalCtxKey struct{}

// Principal gets the principal from the context.
func Principal(ctx context.Context) *UserPrincipal {
	principal, ok := ctx.Value(principalCtxKey{}).(*UserPrincipal)
	if ok {
		return principal
	}

	return nil
}

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
func WithAPIAuth(next http.Handler, srv *AuthServer) http.Handler {
	cookieAuth := NewJWTCookiePrincipalGetter(srv.logger,
		srv.verifier(), IDTokenCookieName)
	headerAuth := NewJWTAuthorizationHeaderPrincipalGetter(srv.logger, srv.verifier())
	multi := MultiAuthPrincipal{cookieAuth, headerAuth}

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		principal, err := multi.Principal(r)
		if err != nil {
			srv.logger.Error(err, "failed to get principal")
		}

		if principal == nil || err != nil {
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
func WithWebAuth(next http.Handler, srv *AuthServer) http.Handler {
	cookieAuth := NewJWTCookiePrincipalGetter(srv.logger,
		srv.verifier(), IDTokenCookieName)
	headerAuth := NewJWTAuthorizationHeaderPrincipalGetter(srv.logger, srv.verifier())
	multi := MultiAuthPrincipal{cookieAuth, headerAuth}

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		principal, err := multi.Principal(r)
		if err != nil {
			srv.logger.Error(err, "failed to get principal")
		}

		if principal == nil || err != nil {
			startAuthFlow(rw, r, srv)
			return
		}

		next.ServeHTTP(rw, r.Clone(WithPrincipal(r.Context(), principal)))
	})
}

func startAuthFlow(rw http.ResponseWriter, r *http.Request, srv *AuthServer) {
	nonce, err := generateNonce()
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
	scopes = append(scopes, scopeProfile)
	authCodeUrl := srv.oauth2Config(scopes).AuthCodeURL(state)

	// Issue state cookie
	http.SetCookie(rw, srv.createCookie(StateCookieName, state))

	http.Redirect(rw, r, authCodeUrl, http.StatusSeeOther)
}

func generateNonce() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}
