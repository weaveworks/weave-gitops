package auth_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-logr/logr"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func TestWithAPIAuthReturns401ForUnauthenticatedRequests(t *testing.T) {
	ctx := context.Background()

	m, err := mockoidc.Run()
	if err != nil {
		t.Error("failed to create fake OIDC server")
	}

	defer func() {
		_ = m.Shutdown()
	}()

	fake := m.Config()
	mux := http.NewServeMux()

	c, err := auth.NewAuthConfig(ctx, fake.Issuer, fake.ClientID, fake.ClientSecret, "", false, 20*time.Minute, http.DefaultClient, logr.Discard())
	if err != nil {
		t.Error("failed to create auth config")
	}

	auth.RegisterAuthHandler(mux, "/oauth2", c)

	s := httptest.NewServer(mux)
	defer s.Close()

	// Set the correct redirect URL now that we have a server running
	c.SetRedirectURL(s.URL + "/oauth2/callback")

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, s.URL, nil)
	auth.WithAPIAuth(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}), c).ServeHTTP(res, req)

	if res.Result().StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status of %d but got %d", http.StatusUnauthorized, res.Result().StatusCode)
	}
}

func TestWithWebAuthRedirectsToOIDCIssuerForUnauthenticatedRequests(t *testing.T) {
	ctx := context.Background()

	m, err := mockoidc.Run()
	if err != nil {
		t.Error("failed to create fake OIDC server")
	}

	defer func() {
		_ = m.Shutdown()
	}()

	fake := m.Config()
	mux := http.NewServeMux()

	c, err := auth.NewAuthConfig(ctx, fake.Issuer, fake.ClientID, fake.ClientSecret, "", false, 20*time.Minute, http.DefaultClient, logr.Discard())
	if err != nil {
		t.Error("failed to create auth config")
	}

	auth.RegisterAuthHandler(mux, "/oauth2", c)

	s := httptest.NewServer(mux)
	defer s.Close()

	// Set the correct redirect URL now that we have a server running
	redirectURL := s.URL + "/oauth2/callback"
	c.SetRedirectURL(redirectURL)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, s.URL, nil)
	auth.WithWebAuth(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}), c).ServeHTTP(res, req)

	if res.Result().StatusCode != http.StatusSeeOther {
		t.Errorf("expected status of %d but got %d", http.StatusSeeOther, res.Result().StatusCode)
	}

	authCodeURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s", m.AuthorizationEndpoint(), fake.ClientID, url.QueryEscape(redirectURL), strings.Join([]string{auth.ScopeProfile, oidc.ScopeOpenID, oidc.ScopeOfflineAccess, auth.ScopeEmail}, "+"))
	if !strings.HasPrefix(res.Result().Header.Get("Location"), authCodeURL) {
		t.Errorf("expected Location header URL to include scopes %s but does not: %s", authCodeURL, res.Result().Header.Get("Location"))
	}
}
