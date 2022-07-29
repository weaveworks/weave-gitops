package middleware

import (
	"context"
	"net/http"

	"google.golang.org/grpc/metadata"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/core/logger"
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
func WithLogging(log logr.Logger, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := &statusRecorder{
			ResponseWriter: w,
			Status:         200,
		}
		h.ServeHTTP(recorder, r)

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
	ProviderToken *oauth2.Token
}

type key int

const (
	tokenKey               key = iota
	GRPCAuthMetadataKey        = "grpc-auth"
	GitProviderTokenHeader     = "Git-Provider-Token"
)

func ContextWithGRPCAuth(ctx context.Context, token string) context.Context {
	md := metadata.New(map[string]string{GRPCAuthMetadataKey: token})
	ctx = metadata.NewOutgoingContext(ctx, md)

	return ctx
}
