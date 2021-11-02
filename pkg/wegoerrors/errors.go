package wegoerrors

import "errors"

var ErrWGEHTTPApiEndpointNotSet = errors.New("the Weave GitOps Enterprise HTTP API endpoint flag (--endpoint) has not been set")
