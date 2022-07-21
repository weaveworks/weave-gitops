package auth

import (
	"context"
	"fmt"
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-logr/logr"
	authv1 "k8s.io/api/authentication/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// BearerTokenPassthroughPrincipalGetter inspects the Authorization
// header (bearer token) and returns it within a principal object.
type BearerTokenPassthroughPrincipalGetter struct {
	log              logr.Logger
	verifier         *oidc.IDTokenVerifier
	headerName       string
	kubernetesClient client.Client
}

// NewBearerTokenPassthroughPrincipalGetter creates a new implementation of the PrincipalGetter
// interface that can decode and verify OIDC Bearer tokens from a named request header.
func NewBearerTokenPassthroughPrincipalGetter(log logr.Logger, verifier *oidc.IDTokenVerifier, headerName string, kubernetesClient client.Client) PrincipalGetter {
	return &BearerTokenPassthroughPrincipalGetter{
		log:              log,
		verifier:         verifier,
		headerName:       headerName,
		kubernetesClient: kubernetesClient,
	}
}

// Principal is an implementation of the PrincipalGetter interface.
//
// Headers of the form Authorization: Bearer <token> are stored within a UserPrincipal.
// The token is not verified, and no ID or Group information will be available.
func (pg *BearerTokenPassthroughPrincipalGetter) Principal(r *http.Request) (*UserPrincipal, error) {
	token := r.Header.Get(pg.headerName)
	if len(token) == 0 {
		return nil, nil
	}

	token = extractToken(token)

	tr := authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{
			Token: token,
		},
	}

	err := pg.kubernetesClient.Create(context.Background(), &tr)
	if err != nil {
		return nil, err
	}

	if !tr.Status.Authenticated {
		return nil, fmt.Errorf("error: user token authentication failed")
	}

	return &UserPrincipal{Token: token}, nil
}

// NewJWTPassthroughCookiePrincipalGetter creates and returns a new
// JWTPassthroughCookiePrincipalGetter.
func NewJWTPassthroughCookiePrincipalGetter(log logr.Logger, verifier *oidc.IDTokenVerifier, cookieName string) PrincipalGetter {
	return &JWTPassthroughCookiePrincipalGetter{
		log:        log,
		verifier:   verifier,
		cookieName: cookieName,
	}
}

// JWTPassthroughCookiePrincipalGetter inspects a cookie for a JWT token and returns a
// principal value.
//
// The JWT Token is parsed, and the token and user/groups are available.
type JWTPassthroughCookiePrincipalGetter struct {
	log        logr.Logger
	verifier   *oidc.IDTokenVerifier
	cookieName string
}

func (pg *JWTPassthroughCookiePrincipalGetter) Principal(r *http.Request) (*UserPrincipal, error) {
	cookie, err := r.Cookie(pg.cookieName)
	if err == http.ErrNoCookie {
		return nil, nil
	}

	principal, err := parseJWTToken(r.Context(), pg.verifier, cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to parse for passthrough: %w", err)
	}
	principal.Token = cookie.Value

	return principal, nil
}
