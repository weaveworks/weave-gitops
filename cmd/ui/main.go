package main

import (
	"context"
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"os"
	"path/filepath"

	"github.com/coreos/go-oidc"
	"github.com/sirupsen/logrus"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"golang.org/x/oauth2"
)

func main() {
	log := logrus.New()
	mux := http.NewServeMux()

	provider, err := oidc.NewProvider(context.Background(), "http://127.0.0.1:5556/dex")
	if err != nil {
		panic(err)
	}

	idTokenVerifier := provider.Verifier(&oidc.Config{ClientID: "example-app"})

	var oauth2Config = oauth2.Config{
		// client_id and client_secret of the client.
		ClientID:     "example-app",
		ClientSecret: "ZXhhbXBsZS1hcHAtc2VjcmV0",

		// The redirectURL.
		RedirectURL: "http://127.0.0.1:9001/callback",

		// Discovery returns the OAuth2 endpoints.
		Endpoint: provider.Endpoint(),

		// "openid" is a required scope for OpenID Connect flows.
		//
		// Other scopes, such as "groups" can be requested.
		Scopes: []string{oidc.ScopeOpenID, "profile", "email", "groups"},
	}

	mux.Handle("/health/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	assetFS := getAssets()
	assetHandler := http.FileServer(http.FS(assetFS))
	redirector := createRedirector(assetFS, log)

	mux.Handle("/callback", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// state := r.URL.Query().Get("state")

		// Verify state.

		oauth2Token, err := oauth2Config.Exchange(r.Context(), r.URL.Query().Get("code"))
		if err != nil {
			log.Errorf("could not exchange token: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Extract the ID Token from OAuth2 token.
		rawIDToken, ok := oauth2Token.Extra("id_token").(string)
		if !ok {
			log.Errorf("could not get raw token: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Parse and verify ID Token payload.
		idToken, err := idTokenVerifier.Verify(r.Context(), rawIDToken)
		if err != nil {
			log.Errorf("could not parse token: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Extract custom claims.
		var claims struct {
			Email    string   `json:"email"`
			Verified bool     `json:"email_verified"`
			Groups   []string `json:"groups"`
		}
		if err := idToken.Claims(&claims); err != nil {
			log.Errorf("could not get get claims: %s", err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{Name: "token", Value: rawIDToken})

		redirector(w, r)

	}))

	mux.Handle("/api/authorize", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearerToken, err := r.Cookie("token")
		if err != nil {
			log.Errorf("could not get token from cookie: %w", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		idToken, err := idTokenVerifier.Verify(r.Context(), bearerToken.Value)
		if err != nil {
			log.Errorf("could verify bearer token: %w", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// Extract custom claims.
		var claims struct {
			Email    string   `json:"email"`
			Verified bool     `json:"email_verified"`
			Groups   []string `json:"groups"`
		}
		if err := idToken.Claims(&claims); err != nil {
			log.Errorf("failed to parse claims: %w", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !claims.Verified {
			log.Errorf("email (%q) in returned claims was not verified", claims.Email)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		response := struct {
			Email  string   `json:"email"`
			Groups []string `json:"groups"`
		}{
			Email:  claims.Email,
			Groups: claims.Groups,
		}

		b, err := json.Marshal(&response)

		if err != nil {
			log.Errorf("could not marshal json %s", err)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Write(b)

	}))

	gitopsServer := server.NewServer(oauth2Config)
	mux.Handle("/api/gitops/", http.StripPrefix("/api/gitops", gitopsServer))

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// Assume anything with a file extension in the name is a static asset.
		extension := filepath.Ext(req.URL.Path)
		// We use the golang http.FileServer for static file requests.
		// This will return a 404 on normal page requests, ie /some-page.
		// Redirect all non-file requests to index.html, where the JS routing will take over.
		if extension == "" {
			redirector(w, req)
			return
		}

		assetHandler.ServeHTTP(w, req)
	}))

	port := ":9001"

	log.Infof("Serving on port %s", port)

	if err := http.ListenAndServe(port, mux); err != nil {
		log.Error(err, "server exited")
		os.Exit(1)
	}
}

//go:embed dist/*
var static embed.FS

func getAssets() fs.FS {
	f, err := fs.Sub(static, "dist")

	if err != nil {
		panic(err)
	}
	return f
}

// A redirector ensures that index.html always gets served.
// The JS router will take care of actual navigation once the index.html page lands.
func createRedirector(fsys fs.FS, log logrus.FieldLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		indexPage, err := fsys.Open("index.html")

		if err != nil {
			log.Error(err, "could not open index.html page")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		stat, err := indexPage.Stat()
		if err != nil {
			log.Error(err, "could not get index.html stat")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		bt := make([]byte, stat.Size())
		_, err = indexPage.Read(bt)

		if err != nil {
			log.Error(err, "could not read index.html")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = w.Write(bt)

		if err != nil {
			log.Error(err, "error writing index.html")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
