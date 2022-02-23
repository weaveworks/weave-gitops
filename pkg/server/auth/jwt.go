package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-logr/logr"
)

// PrincipalGetter implementations are responsible for extracting a named
// principal from an HTTP request.
type PrincipalGetter interface {
	// Principal extracts a principal from the http.Request.
	// It's not an error for there to be no principal in the request.
	Principal(r *http.Request) (*UserPrincipal, error)
}

// JWTCookiePrincipalGetter inspects a cookie for a JWT token
// and returns a principal object.
type JWTCookiePrincipalGetter struct {
	log        logr.Logger
	verifier   *oidc.IDTokenVerifier
	cookieName string
}

func NewJWTCookiePrincipalGetter(log logr.Logger, verifier *oidc.IDTokenVerifier, cookieName string) PrincipalGetter {
	return &JWTCookiePrincipalGetter{
		log:        log,
		verifier:   verifier,
		cookieName: cookieName,
	}
}

func (pg *JWTCookiePrincipalGetter) Principal(r *http.Request) (*UserPrincipal, error) {
	pg.log.Info("attempt to read token from cookie")

	cookie, err := r.Cookie(pg.cookieName)
	if err == http.ErrNoCookie {
		return nil, nil
	}

	return parseJWTToken(r.Context(), pg.verifier, cookie.Value)
}

// JWTAuthorizationHeaderPrincipalGetter inspects the Authorization
// header (bearer token) for a JWT token and returns a principal
// object.
type JWTAuthorizationHeaderPrincipalGetter struct {
	log      logr.Logger
	verifier *oidc.IDTokenVerifier
}

func NewJWTAuthorizationHeaderPrincipalGetter(log logr.Logger, verifier *oidc.IDTokenVerifier) PrincipalGetter {
	return &JWTAuthorizationHeaderPrincipalGetter{
		log:      log,
		verifier: verifier,
	}
}

func (pg *JWTAuthorizationHeaderPrincipalGetter) Principal(r *http.Request) (*UserPrincipal, error) {
	pg.log.Info("attempt to read token from auth header")

	header := r.Header.Get("Authorization")
	if header == "" {
		return nil, nil
	}

	return parseJWTToken(r.Context(), pg.verifier, extractToken(header))
}

func extractToken(s string) string {
	parts := strings.Split(s, " ")
	if len(parts) != 2 {
		return ""
	}

	if strings.TrimSpace(parts[0]) != "Bearer" {
		return ""
	}

	return strings.TrimSpace(parts[1])
}

func parseJWTToken(ctx context.Context, verifier *oidc.IDTokenVerifier, rawIDToken string) (*UserPrincipal, error) {
	token, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify JWT token: %w", err)
	}

	var claims struct {
		Email  string   `json:"email"`
		Groups []string `json:"groups"`
	}

	if err := token.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to parse claims from the JWT token: %w", err)
	}

	return &UserPrincipal{ID: claims.Email, Groups: claims.Groups}, nil
}

type JWTAdminCookiePrincipalGetter struct {
	log        logr.Logger
	verifier   TokenSignerVerifier
	cookieName string
}

func NewJWTAdminCookiePrincipalGetter(log logr.Logger, verifier TokenSignerVerifier, cookieName string) PrincipalGetter {
	return &JWTAdminCookiePrincipalGetter{
		log:        log,
		verifier:   verifier,
		cookieName: cookieName,
	}
}

func (pg *JWTAdminCookiePrincipalGetter) Principal(r *http.Request) (*UserPrincipal, error) {
	pg.log.Info("attempt to read token from cookie")

	cookie, err := r.Cookie(pg.cookieName)
	if err == http.ErrNoCookie {
		return nil, nil
	}

	return parseJWTAdminToken(pg.verifier, cookie.Value)
}

func parseJWTAdminToken(verifier TokenSignerVerifier, rawIDToken string) (*UserPrincipal, error) {
	claims, err := verifier.Verify(rawIDToken)
	if err != nil {
		// FIXME: do some better handling here
		// return nil, fmt.Errorf("failed to verify JWT token: %w", err)
		// ANYWAY:, its probably not our token? e.g. an OIDC one
		return nil, nil
	}

	return &UserPrincipal{ID: claims.Subject, Groups: []string{}}, nil
}

// MultiAuthPrincipal looks for a principal in an array of principal getters and
// if it finds an error or a principal it returns, otherwise it returns (nil,nil).
type MultiAuthPrincipal []PrincipalGetter

func (m MultiAuthPrincipal) Principal(r *http.Request) (*UserPrincipal, error) {
	for _, v := range m {
		p, err := v.Principal(r)
		if err != nil {
			return nil, err
		}

		if p != nil {
			return p, nil
		}
	}

	return nil, nil
}
