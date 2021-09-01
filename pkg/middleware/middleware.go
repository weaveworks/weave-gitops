package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/pkg/logger"
	"golang.org/x/oauth2"
)

type statusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

var RequestOkText = "request success"
var RequestErrorText = "request error"
var ServerErrorText = "server error"

// WithGrpcErrorLogging logs errors returned from server RPC handlers.
// Our errors happen in gRPC land, so we cannot introspect into the content of
// the error message in the WithLogging http.Handler.
// Normal gRPC middleware was not working for this:
// https://github.com/grpc-ecosystem/grpc-gateway/issues/1043
func WithGrpcErrorLogging(log logr.Logger) runtime.ServeMuxOption {
	return runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
		log.Error(err, ServerErrorText)
		// We don't want to change the behavior of error handling, just intercept for logging.
		runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
	})
}

// WithLogging adds basic logging for HTTP requests.
// Note that this accepts a grpc-gateway ServeMux instead of an http.Handler.
func WithLogging(log logr.Logger, mux *runtime.ServeMux) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{
			ResponseWriter: w,
			Status:         200,
		}
		mux.ServeHTTP(recorder, r)

		l := log.WithValues("uri", r.RequestURI, "status", recorder.Status)

		if recorder.Status < 400 {
			l.V(logger.LogLevelDebug).Info(RequestOkText)
		}

		if recorder.Status >= 400 && recorder.Status < 500 {
			l.V(logger.LogLevelWarn).Info(RequestErrorText)
		}

		if recorder.Status >= 500 {
			l.V(logger.LogLevelError).Info(ServerErrorText)
		}
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
