package types

import (
	"context"
	"net/http"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

// AuthFlow is an interface for OAuth authorization flows
//counterfeiter:generate . AuthFlow
type AuthFlow interface {
	Authorize(ctx context.Context) (*http.Request, error)
	CallbackHandler(*TokenResponseState, http.Handler) http.Handler
}
