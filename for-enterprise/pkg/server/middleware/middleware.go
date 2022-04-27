package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/gitops/pkg/services/auth"
	"golang.org/x/oauth2"
	"google.golang.org/grpc/metadata"
)

type contextVals struct {
	ProviderToken *oauth2.Token
}

type key int

const (
	tokenKey               key = iota
	GRPCAuthMetadataKey        = "grpc-auth"
	GitProviderTokenHeader     = "Git-Provider-Token"
)

// Injects the token into the request context to be retrieved later.
// Use the ExtractToken func inside the server handler where appropriate.
func WithProviderToken(jwtClient auth.JWTClient, h http.Handler, log logr.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get(GitProviderTokenHeader)
		tokenSlice := strings.Split(tokenStr, "token ")

		if len(tokenSlice) < 2 {
			log.Info("missing or invalid token.")
			// No token specified. Nothing to be done.
			// We do NOT return 400 here because there may be some 'unauthenticated' routes (ie /login)
			h.ServeHTTP(w, r)
			return
		}

		// The actual token data
		token := tokenSlice[1]

		claims, err := jwtClient.VerifyJWT(token)
		if err != nil {
			log.Info("could not parse claims: " + err.Error())
			// Certain routes do not require a token, so pass the request through.
			// If the route requires a token and it isn't present,
			// the next handler will error and return that to the user.
			h.ServeHTTP(w, r)
			return
		}

		vals := contextVals{ProviderToken: &oauth2.Token{AccessToken: claims.ProviderToken}}

		c := context.WithValue(r.Context(), tokenKey, vals)
		r = r.WithContext(c)
		h.ServeHTTP(w, r)
	})
}

// Get the token from request context.
func ExtractProviderToken(ctx context.Context) (*oauth2.Token, error) {
	// Tests use straight GRPC connections instead of the http gateway.
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		val := md.Get(GRPCAuthMetadataKey)
		if val != nil {
			return &oauth2.Token{AccessToken: val[0]}, nil
		}
	}

	c := ctx.Value(tokenKey)

	vals, ok := c.(contextVals)
	if !ok {
		return nil, errors.New("could not get token from context")
	}

	if vals.ProviderToken == nil || vals.ProviderToken.AccessToken == "" {
		return nil, errors.New("no token specified")
	}

	return vals.ProviderToken, nil
}
