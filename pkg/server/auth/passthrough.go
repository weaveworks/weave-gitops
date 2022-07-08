package auth

import (
	"net/http"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-logr/logr"
)

type BearerTokenPassthroughPrincipalGetter struct {
	log        logr.Logger
	verifier   *oidc.IDTokenVerifier
	headerName string
}

func NewBearerTokenPassthroughPrincipalGetter(log logr.Logger, verifier *oidc.IDTokenVerifier, headerName string) PrincipalGetter {
	return &BearerTokenPassthroughPrincipalGetter{
		log:        log,
		verifier:   verifier,
		headerName: headerName,
	}
}

func (pg *BearerTokenPassthroughPrincipalGetter) Principal(r *http.Request) (*UserPrincipal, error) {
	token := r.Header.Get(pg.headerName)
	if len(token) == 0 {
		return nil, nil
	}

	return &UserPrincipal{Token: token}, nil
}
