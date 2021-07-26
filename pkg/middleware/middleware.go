package middleware

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

type statusRecorder struct {
	http.ResponseWriter
	Status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.Status = status
	r.ResponseWriter.WriteHeader(status)
}

type grpcGatewayErrorMsg struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type key int

const grpcRequestInfoContextKey key = iota

type grpcRequestInfo struct {
	Err error
}

// InjectErrorIntoContext injects the gRPC error message into the request context.
// Our errors happen in gRPC land, so we cannot introspect into the content of
// the error message in the WithLogging http.Handler. We need to extract the
// error message and inject it into the request context to fetch it later in the logging middleware.
// We can't log here because it would create duplicate log messages when combined with the WithLogging middleware.
// https://github.com/grpc-ecosystem/grpc-gateway/issues/1043
func InjectErrorIntoContext() runtime.ServeMuxOption {
	return runtime.WithErrorHandler(func(ctx context.Context, mux *runtime.ServeMux, marshaler runtime.Marshaler, w http.ResponseWriter, r *http.Request, err error) {
		newCtx := context.WithValue(r.Context(), grpcRequestInfoContextKey, grpcRequestInfo{Err: err})
		r = r.WithContext(newCtx)
		// We don't want to change the behavior of error handling, just intercept for logging.
		runtime.DefaultHTTPErrorHandler(ctx, mux, marshaler, w, r, err)
	})
}

// Adds basic logging for HTTP requests.
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
			l.V(logger.LogLevelDebug).Info("request success")
		}

		if recorder.Status >= 400 && recorder.Status < 500 {
			l.V(logger.LogLevelWarn).Info("request error")
		}

		if recorder.Status >= 500 {
			c := r.Context().Value(grpcRequestInfoContextKey)

			vals, ok := c.(grpcRequestInfo)
			if !ok {
				l.Error(errors.New("could not get error from context"), "server error")
				return

			}
			l.Error(vals.Err, "server error")
		}
	})
}
