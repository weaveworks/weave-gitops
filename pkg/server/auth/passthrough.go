package auth

import (
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-logr/logr"
)

// BearerTokenPassthroughPrincipalGetter inspects the Authorization
// header (bearer token) and returns it within a principal object.
type BearerTokenPassthroughPrincipalGetter struct {
	log        logr.Logger
	verifier   *oidc.IDTokenVerifier
	headerName string
}

// NewBearerTokenPassthroughPrincipalGetter creates a new implementation of the PrincipalGetter
// interface that can decode and verify OIDC Bearer tokens from a named request header.
func NewBearerTokenPassthroughPrincipalGetter(log logr.Logger, verifier *oidc.IDTokenVerifier, headerName string) PrincipalGetter {
	return &BearerTokenPassthroughPrincipalGetter{
		log:        log,
		verifier:   verifier,
		headerName: headerName,
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

	return &UserPrincipal{Token: extractToken(token)}, nil
}
