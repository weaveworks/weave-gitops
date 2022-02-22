package cmderrors

import "errors"

var (
	ErrNoWGEEndpoint  = errors.New("the Weave GitOps Enterprise HTTP API endpoint flag (--endpoint) has not been set")
	ErrNoURL          = errors.New("the URL flag (--url) has not been set")
	ErrNoIssuerURL    = errors.New("the OIDC issuer URL flag (--oidc-issuer-url) has not been set")
	ErrNoClientID     = errors.New("the OIDC client ID flag (--oidc-client-id) has not been set")
	ErrNoClientSecret = errors.New("the OIDC client secret flag (--oidc-client-secret) has not been set")
	ErrNoRedirectURL  = errors.New("the OIDC redirect URL flag (--oidc-redirect-url) has not been set")
	ErrNoTLSCertOrKey = errors.New("both tls private key and cert must be specified")
)
