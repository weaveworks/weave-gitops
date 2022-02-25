package auth_test

import (
	"bytes"
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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
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
	fakeKubernetesClient := ctrlclient.NewClientBuilder().Build()

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	if err != nil {
		t.Errorf("failed to create HMAC signer: %v", err)
	}

	srv, err := auth.NewAuthServer(ctx, logr.Discard(), http.DefaultClient,
		auth.AuthConfig{
			auth.OIDCConfig{
				IssuerURL:     fake.Issuer,
				ClientID:      fake.ClientID,
				ClientSecret:  fake.ClientSecret,
				RedirectURL:   "",
				TokenDuration: 20 * time.Minute,
			},
		}, fakeKubernetesClient, tokenSignerVerifier)
	if err != nil {
		t.Error("failed to create auth config")
	}

	_ = auth.RegisterAuthServer(mux, "/oauth2", srv, 1)

	s := httptest.NewServer(mux)
	defer s.Close()

	// Set the correct redirect URL now that we have a server running
	srv.SetRedirectURL(s.URL + "/oauth2/callback")

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, s.URL, nil)
	auth.WithAPIAuth(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}), srv, nil).ServeHTTP(res, req)

	if res.Result().StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status of %d but got %d", http.StatusUnauthorized, res.Result().StatusCode)
	}

	// Test out the publicRoutes
	res = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, s.URL+"/v1/featureflags", nil)
	auth.WithAPIAuth(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}), srv, []string{"/v1/featureflags"}).ServeHTTP(res, req)

	if res.Result().StatusCode != http.StatusOK {
		t.Errorf("expected status of %d but got %d", http.StatusUnauthorized, res.Result().StatusCode)
	}
}

func TestOauth2FlowRedirectsToOIDCIssuerForUnauthenticatedRequests(t *testing.T) {
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
	fakeKubernetesClient := ctrlclient.NewClientBuilder().Build()

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	if err != nil {
		t.Errorf("failed to create HMAC signer: %v", err)
	}

	srv, err := auth.NewAuthServer(ctx, logr.Discard(), http.DefaultClient,
		auth.AuthConfig{
			auth.OIDCConfig{
				IssuerURL:     fake.Issuer,
				ClientID:      fake.ClientID,
				ClientSecret:  fake.ClientSecret,
				RedirectURL:   "",
				TokenDuration: 20 * time.Minute,
			},
		}, fakeKubernetesClient, tokenSignerVerifier)
	if err != nil {
		t.Error("failed to create auth config")
	}

	_ = auth.RegisterAuthServer(mux, "/oauth2", srv, 1)

	s := httptest.NewServer(mux)
	defer s.Close()

	// Set the correct redirect URL now that we have a server running
	redirectURL := s.URL + "/oauth2/callback"
	srv.SetRedirectURL(redirectURL)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, s.URL, nil)
	srv.OAuth2Flow().ServeHTTP(res, req)

	if res.Result().StatusCode != http.StatusSeeOther {
		t.Errorf("expected status of %d but got %d", http.StatusSeeOther, res.Result().StatusCode)
	}

	authCodeURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s", m.AuthorizationEndpoint(), fake.ClientID, url.QueryEscape(redirectURL), strings.Join([]string{"profile", oidc.ScopeOpenID, oidc.ScopeOfflineAccess, "email"}, "+"))
	if !strings.HasPrefix(res.Result().Header.Get("Location"), authCodeURL) {
		t.Errorf("expected Location header URL to include scopes %s but does not: %s", authCodeURL, res.Result().Header.Get("Location"))
	}
}

func TestIsPublicRoute(t *testing.T) {
	assert.True(t, auth.IsPublicRoute(&url.URL{Path: "/foo"}, []string{"/foo"}))
	assert.False(t, auth.IsPublicRoute(&url.URL{Path: "foo"}, []string{"/foo"}))
	assert.False(t, auth.IsPublicRoute(&url.URL{Path: "/foob"}, []string{"/foo"}))
}

func TestRateLimit(t *testing.T) {
	ctx := context.Background()
	mux := http.NewServeMux()
	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	require.NoError(t, err)

	password := "my-secret-password"
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	require.NoError(t, err)

	hashedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "admin-password-hash",
			Namespace: "wego-system",
		},
		Data: map[string][]byte{
			"password": hashed,
		},
	}
	fakeKubernetesClient := ctrlclient.NewClientBuilder().WithObjects(hashedSecret).Build()

	srv, err := auth.NewAuthServer(ctx, logr.Discard(), http.DefaultClient,
		auth.AuthConfig{
			auth.OIDCConfig{
				TokenDuration: 20 * time.Minute,
			},
		}, fakeKubernetesClient, tokenSignerVerifier)
	require.NoError(t, err)
	err = auth.RegisterAuthServer(mux, "/oauth2", srv, 1)
	require.NoError(t, err)

	s := httptest.NewServer(mux)
	defer s.Close()

	res1, err := http.Post(s.URL+"/oauth2/sign_in", "application/json", bytes.NewReader([]byte(`{"password":"my-secret-password"}`)))
	require.NoError(t, err)

	if res1.StatusCode != http.StatusOK {
		t.Errorf("expected 200 but got %d instead", res1.StatusCode)
	}

	res2, err := http.Post(s.URL+"/oauth2/sign_in", "application/json", bytes.NewReader([]byte(`{"password":"my-secret-password"}`)))
	require.NoError(t, err)

	if res2.StatusCode != http.StatusTooManyRequests {
		t.Errorf("expected 429 but got %d instead", res2.StatusCode)
	}

	time.Sleep(time.Second)

	res3, err := http.Post(s.URL+"/oauth2/sign_in", "application/json", bytes.NewReader([]byte(`{"password":"my-secret-password"}`)))
	require.NoError(t, err)

	if res3.StatusCode != http.StatusOK {
		t.Errorf("expected 200 but got %d instead", res3.StatusCode)
	}
}
