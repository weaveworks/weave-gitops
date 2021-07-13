package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
	"golang.org/x/oauth2"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type statusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

// Adds basic logging for HTTP requests.
// Note that this accepts a grpc-gateway ServeMux instead of a "normal" handler.
func WithLogging(h *runtime.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{
			ResponseWriter: w,
			Status:         200,
		}
		h.ServeHTTP(recorder, r)
		log.WithFields(log.Fields{
			"uri":    r.RequestURI,
			"status": recorder.Status,
		}).Info()
	})
}

type contextVals struct {
	Token *oauth2.Token
}

type key int

const tokenKey key = iota

// Injects the token into the request context to be retrieved later.
// Use the ExtractToken func inside the server handler where appropriate.
func WithToken(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		// Token gets specified with the work "token", then the data.
		// "token abc123def..."
		tokenSlice := strings.Split(tokenStr, "token ")

		if len(tokenSlice) < 2 {
			// No token specified. Nothing to be done.
			// We do NOT return 400 here because there may be some 'unauthenticated' routes (ie /login)
			h.ServeHTTP(w, r)
			return
		}

		// The actual token data
		token := tokenSlice[1]

		vals := contextVals{Token: &oauth2.Token{AccessToken: token}}

		c := context.WithValue(r.Context(), tokenKey, vals)
		r = r.WithContext(c)
		h.ServeHTTP(w, r)
	})
}

// Get the token from request context.
func ExtractToken(ctx context.Context) (*oauth2.Token, error) {
	c := ctx.Value(tokenKey)

	vals, ok := c.(contextVals)
	if !ok {
		return nil, errors.New("could not get token from context")
	}

	if vals.Token == nil || vals.Token.AccessToken == "" {
		return nil, errors.New("no token specified")
	}

	return vals.Token, nil
}
