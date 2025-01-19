package check

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"

	"github.com/weaveworks/weave-gitops/pkg/logger"
)

//go:embed success.html
var successHTML string

//go:embed error.html
var errorHTML string

// retrieveIDToken starts an HTTP server that handles the response to an OIDC authentication request,
// issues a token request and returns the ID token. The HTTP server is always shut down before this function returns.
func retrieveIDToken(log logger.Logger, oauth2Config oauth2.Config, verifier *oidc.IDTokenVerifier) (*oidc.IDToken, error) {
	mux := http.ServeMux{}
	srv := http.Server{
		Handler:           &mux,
		ReadHeaderTimeout: 5 * time.Second,
	}
	var idToken *oidc.IDToken
	var handleErr error
	quitCh := make(chan struct{})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if handleErr != nil {
				fmt.Fprint(w, errorHTML)
			}
			close(quitCh)
		}()

		if err := handleServerError(r); err != nil {
			handleErr = err
			return
		}

		log.Successf("received response from OIDC provider")

		log.Actionf("exchanging code for token")
		ctx := context.Background()
		oauth2Token, err := oauth2Config.Exchange(ctx, r.URL.Query().Get("code"))
		if err != nil {
			handleErr = fmt.Errorf("error exchanging code: %w", err)
			return
		}

		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			handleErr = fmt.Errorf("no id_token field in OAuth 2 token")
			return
		}
		idToken, err = verifier.Verify(ctx, rawIDToken)
		if err != nil {
			handleErr = fmt.Errorf("ID token verification failed: %w", err)
			return
		}

		fmt.Fprint(w, successHTML)
	})

	listener, err := net.Listen("tcp", ":9876") // #nosec G102
	if err != nil {
		return nil, fmt.Errorf("failed starting listener: %w", err)
	}

	shutdownCompleteCh := make(chan struct{})
	go func() {
		<-quitCh
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := srv.Shutdown(ctx); err != nil {
			log.Warningf("local HTTP server shutdown failed: %s", err)
		}
		close(shutdownCompleteCh)
	}()

	if err := srv.Serve(listener); err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			return nil, fmt.Errorf("failed starting server: %w", err)
		}
	}

	<-shutdownCompleteCh

	return idToken, handleErr
}

// handleServerError parses the query parameters from an authentication error response
// and constructs an error variable from them.
func handleServerError(r *http.Request) error {
	q := r.URL.Query()
	err := q.Get("error")
	if err != "" {
		var buf strings.Builder
		fmt.Fprintf(&buf, "received error from identity provider: %s", q.Get("error"))
		desc := q.Get("error_description")
		if desc != "" {
			fmt.Fprintf(&buf, " (%s)", desc)
		}
		return errors.New(buf.String())
	}
	return nil
}
