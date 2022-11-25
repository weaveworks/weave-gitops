package auth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/core/logger"
)

// PrincipalGetter implementations are responsible for extracting a named
// principal from an HTTP request.
type PrincipalGetter interface {
	// Principal extracts a principal from the http.Request.
	// It's not an error for there to be no principal in the request.
	Principal(r *http.Request) (*UserPrincipal, error)
}

type tokenVerifier interface {
	Verify(ctx context.Context, rawIDToken string) (*oidc.IDToken, error)
}

// JWTCookiePrincipalGetter inspects a cookie for a JWT token
// and returns a principal object.
type JWTCookiePrincipalGetter struct {
	log          logr.Logger
	verifier     tokenVerifier
	cookieName   string
	claimsConfig *ClaimsConfig
}

// NewJWTCookiePrincipalGetter looks for a cookie in the provided name and
// treats that as a JWT token that can be decoded to a Principal.
func NewJWTCookiePrincipalGetter(log logr.Logger, verifier tokenVerifier, cookieName string, config *ClaimsConfig) PrincipalGetter {
	return &JWTCookiePrincipalGetter{
		log:          log,
		verifier:     verifier,
		cookieName:   cookieName,
		claimsConfig: config,
	}
}

func (pg *JWTCookiePrincipalGetter) Principal(r *http.Request) (*UserPrincipal, error) {
	cookie, err := r.Cookie(pg.cookieName)
	if err == http.ErrNoCookie {
		return nil, nil
	}
	pg.log.V(logger.LogLevelDebug).Info("parsing cookie JWT token", "claimsConfig", pg.claimsConfig)

	return parseJWTToken(r.Context(), pg.verifier, cookie.Value, pg.claimsConfig)
}

// JWTAuthorizationHeaderPrincipalGetter inspects the Authorization
// header (bearer token) for a JWT token and returns a principal
// object.
type JWTAuthorizationHeaderPrincipalGetter struct {
	log          logr.Logger
	verifier     tokenVerifier
	claimsConfig *ClaimsConfig
}

func NewJWTAuthorizationHeaderPrincipalGetter(log logr.Logger, verifier tokenVerifier, config *ClaimsConfig) PrincipalGetter {
	return &JWTAuthorizationHeaderPrincipalGetter{
		log:          log,
		verifier:     verifier,
		claimsConfig: config,
	}
}

func (pg *JWTAuthorizationHeaderPrincipalGetter) Principal(r *http.Request) (*UserPrincipal, error) {
	pg.log.Info("attempt to read token from auth header")

	header := r.Header.Get("Authorization")
	if header == "" {
		return nil, nil
	}
	pg.log.V(logger.LogLevelDebug).Info("parsing authorization header JWT token", "claimsConfig", pg.claimsConfig)

	return parseJWTToken(r.Context(), pg.verifier, extractToken(header), pg.claimsConfig)
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

func parseJWTToken(ctx context.Context, verifier tokenVerifier, rawIDToken string, cc *ClaimsConfig) (*UserPrincipal, error) {
	token, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return nil, fmt.Errorf("failed to verify JWT token: %w", err)
	}

	return cc.PrincipalFromClaims(token)
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
type MultiAuthPrincipal struct {
	Log     logr.Logger
	Getters []PrincipalGetter
}

func (m MultiAuthPrincipal) Principal(r *http.Request) (*UserPrincipal, error) {
	for _, v := range m.Getters {
		p, err := v.Principal(r)
		if err != nil {
			return nil, err
		}

		if p != nil {
			m.Log.V(logger.LogLevelDebug).Info("Found principal", "user", p.ID, "groups", p.Groups, "tokenLength", len(p.Token()), "method", reflect.TypeOf(v))

			return p, nil
		}
	}

	return nil, errors.New("could not find valid principal")
}
