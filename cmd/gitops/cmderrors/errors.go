package cmderrors

import "errors"

var (
	ErrNoWGEEndpoint  = errors.New("the Weave GitOps Enterprise HTTP API endpoint flag (--endpoint) has not been set")
	ErrNoURL          = errors.New("the URL flag (--url) has not been set")
	ErrNoTLSCertOrKey = errors.New("both tls private key and cert must be specified")
)
