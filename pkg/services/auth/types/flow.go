package types

import (
	"context"
	"net/http"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . AuthFlow

// AuthFlow is an interface for OAuth authorization flows
type AuthFlow interface {
	Authorize(ctx context.Context) (*http.Request, error)
	CallbackHandler(*TokenResponseState, http.Handler) http.Handler
}
