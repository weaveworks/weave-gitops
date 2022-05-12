package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"net/url"

	"github.com/sethvargo/go-limiter/httplimit"
	"github.com/sethvargo/go-limiter/memorystore"
)

const (
	// StateCookieName is the name of the cookie that holds state during auth flow.
	StateCookieName = "state"
	// IDTokenCookieName is the name of the cookie that holds the ID Token once
	// the user has authenticated successfully with the OIDC Provider.
	IDTokenCookieName = "id_token"
	// ScopeProfile is the "profile" scope
	scopeProfile = "profile"
	// ScopeEmail is the "email" scope
	scopeEmail = "email"
	// ScopeGroups is the "groups" scope
	scopeGroups = "groups"
)

// RegisterAuthServer registers the /callback route under a specified prefix.
// This route is called by the OIDC Provider in order to pass back state after
// the authentication flow completes.
func RegisterAuthServer(mux *http.ServeMux, prefix string, srv *AuthServer, loginRequestRateLimit uint64) error {
	store, err := memorystore.New(&memorystore.Config{
		Tokens: loginRequestRateLimit,
	})
	if err != nil {
		return err
	}

	middleware, err := httplimit.NewMiddleware(store, httplimit.IPKeyFunc())
	if err != nil {
		return err
	}

	mux.Handle(prefix, srv.OAuth2Flow())
	mux.Handle(prefix+"/callback", srv.Callback())
	mux.Handle(prefix+"/sign_in", middleware.Handle(srv.SignIn()))
	mux.Handle(prefix+"/userinfo", srv.UserInfo())
	mux.Handle(prefix+"/logout", srv.Logout())

	return nil
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
func WithAPIAuth(next http.Handler, srv *AuthServer, publicRoutes []string) http.Handler {
	adminAuth := NewJWTAdminCookiePrincipalGetter(srv.Log, srv.tokenSignerVerifier, IDTokenCookieName)
	multi := MultiAuthPrincipal{adminAuth}

	if srv.oidcEnabled() {
		headerAuth := NewJWTAuthorizationHeaderPrincipalGetter(srv.Log, srv.verifier())
		cookieAuth := NewJWTCookiePrincipalGetter(srv.Log, srv.verifier(), IDTokenCookieName)
		multi = append(multi, headerAuth, cookieAuth)
	}

	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if IsPublicRoute(r.URL, publicRoutes) {
			next.ServeHTTP(rw, r)
			return
		}

		principal, err := multi.Principal(r)
		if err != nil {
			srv.Log.Error(err, "failed to get principal")
		}

		if principal == nil || err != nil {
			JSONError(srv.Log, rw, "Authentication required", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(rw, r.Clone(WithPrincipal(r.Context(), principal)))
	})
}

func generateNonce() (string, error) {
	b := make([]byte, 32)

	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b), nil
}

func IsPublicRoute(u *url.URL, publicRoutes []string) bool {
	for _, pr := range publicRoutes {
		if u.Path == pr {
			return true
		}
	}

	return false
}
