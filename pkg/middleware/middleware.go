package middleware

import (
	"context"
	"errors"
	"fmt"
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
	AccessToken *oauth2.Token
}

const tokenKey = "wego_auth_token"

func WithToken(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")

		tokenSlice := strings.Split(tokenStr, "token ")

		if len(tokenSlice) < 2 {
			h.ServeHTTP(w, r)
			return
		}

		fmt.Println(len(tokenSlice))
		token := tokenSlice[1]

		vals := contextVals{AccessToken: &oauth2.Token{AccessToken: token}}

		c := context.WithValue(r.Context(), tokenKey, vals)
		r = r.WithContext(c)
		h.ServeHTTP(w, r)
	})
}

func ExtractToken(ctx context.Context) (*oauth2.Token, error) {
	c := ctx.Value(tokenKey)

	vals, ok := c.(contextVals)
	if !ok {
		return nil, errors.New("could not get token from context")
	}

	return vals.AccessToken, nil
}
