package auth

import (
	"context"
	"net/http"
)

// SessionManager implementations provide session storage for requests.
type SessionManager interface {
	LoadAndSave(next http.Handler) http.Handler
	GetString(context.Context, string) string
	Remove(context.Context, string)
	Put(ctx context.Context, key string, val interface{})
	Destroy(ctx context.Context) error
}
