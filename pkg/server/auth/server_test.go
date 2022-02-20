package auth_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlclientfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestCallbackSupportsGet(t *testing.T) {
	methods := []string{
		http.MethodPost,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodHead,
		http.MethodOptions,
	}

	s, _ := makeAuthServer(t, nil, nil)

	for _, m := range methods {
		req := httptest.NewRequest(m, "https://example.com/callback", nil)
		w := httptest.NewRecorder()
		s.Callback().ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected status to be 405 but got %v instead", resp.StatusCode)
		}

		if resp.Header.Get("Allow") != "GET" {
			t.Errorf("expected `Allow` header to be set to `GET` but was not")
		}
	}
}

func TestCallbackErrorFromOIDC(t *testing.T) {
	s, _ := makeAuthServer(t, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "https://example.com/callback?error=invalid_request&error_description=Unsupported%20response_type%20value", nil)
	w := httptest.NewRecorder()
	s.Callback().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status to be 400 but got %v instead", resp.StatusCode)
	}
}

func TestCallbackCodeIsEmpty(t *testing.T) {
	s, _ := makeAuthServer(t, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "https://example.com/callback", nil)
	w := httptest.NewRecorder()
	s.Callback().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status to be 400 but got %v instead", resp.StatusCode)
	}
}

func TestCallbackStateCookieNotSet(t *testing.T) {
	s, _ := makeAuthServer(t, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "https://example.com/callback?code=123", nil)
	w := httptest.NewRecorder()
	s.Callback().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status to be 400 but got %v instead", resp.StatusCode)
	}
}

func TestCallbackStateCookieNotValid(t *testing.T) {
	s, _ := makeAuthServer(t, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "https://example.com/callback?code=123&state=some_state", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.StateCookieName,
		Value: "some_different_state",
	})

	w := httptest.NewRecorder()
	s.Callback().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status to be 400 but got %v instead", resp.StatusCode)
	}
}

func TestCallbackStateCookieNotBase64Encoded(t *testing.T) {
	s, _ := makeAuthServer(t, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "https://example.com/callback?code=123&state=some_state", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.StateCookieName,
		Value: "some_state",
	})

	w := httptest.NewRecorder()
	s.Callback().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status to be 400 but got %v instead", resp.StatusCode)
	}
}

func TestCallbackStateCookieNotJSONPayload(t *testing.T) {
	s, _ := makeAuthServer(t, nil, nil)

	encState := base64.StdEncoding.EncodeToString([]byte("some_state"))

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("https://example.com/callback?code=123&state=%s", encState), nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.StateCookieName,
		Value: encState,
	})

	w := httptest.NewRecorder()
	s.Callback().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status to be 400 but got %v instead", resp.StatusCode)
	}
}

func TestCallbackCodeExchangeError(t *testing.T) {
	s, _ := makeAuthServer(t, nil, nil)

	state, _ := json.Marshal(auth.SessionState{
		Nonce:     "abcde",
		ReturnURL: "https://example.com",
	})
	encState := base64.StdEncoding.EncodeToString(state)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("https://example.com/callback?code=123&state=%s", encState), nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.StateCookieName,
		Value: encState,
	})

	w := httptest.NewRecorder()
	s.Callback().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected status to be 500 but got %v instead", resp.StatusCode)
	}
}

func TestSignInSupportsPOST(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodHead,
		http.MethodOptions,
	}

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	if err != nil {
		t.Errorf("failed to create HMAC signer: %v", err)
	}

	s, _ := makeAuthServer(t, ctrlclientfake.NewClientBuilder().Build(), tokenSignerVerifier)

	for _, m := range methods {
		req := httptest.NewRequest(m, "https://example.com/signin", nil)
		w := httptest.NewRecorder()
		s.SignIn().ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected status to be 405 but got %v instead", resp.StatusCode)
		}

		if resp.Header.Get("Allow") != "POST" {
			t.Errorf("expected `Allow` header to be set to `POST` but was not")
		}
	}
}

func TestSignInNoPayloadReturnsBadRequest(t *testing.T) {
	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	if err != nil {
		t.Errorf("failed to create HMAC signer: %v", err)
	}

	s, _ := makeAuthServer(t, ctrlclientfake.NewClientBuilder().Build(), tokenSignerVerifier)

	req := httptest.NewRequest(http.MethodPost, "https://example.com/signin", nil)
	w := httptest.NewRecorder()
	s.SignIn().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status to be 400 but got %v instead", resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("expected to read response body successfully but got error instead: %v", err)
	}

	respBody := string(b)
	if respBody != "Failed to read request body.\n" {
		t.Errorf("expected different response body but got instead: %q", respBody)
	}
}

func TestSignInNoSecret(t *testing.T) {
	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	if err != nil {
		t.Errorf("failed to create HMAC signer: %v", err)
	}

	s, _ := makeAuthServer(t, ctrlclientfake.NewClientBuilder().Build(), tokenSignerVerifier)

	j, err := json.Marshal(auth.LoginRequest{})
	if err != nil {
		t.Errorf("failed to marshal to JSON: %v", err)
	}

	reader := bytes.NewReader(j)

	req := httptest.NewRequest(http.MethodPost, "https://example.com/signin", reader)
	w := httptest.NewRecorder()
	s.SignIn().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status to be 400 but got %v instead", resp.StatusCode)
	}
}

func TestSignInWrongPasswordReturnsUnauthorized(t *testing.T) {
	password := "my-secret-password"

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		t.Errorf("failed to generate a hash from password: %v", err)
	}

	hashedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "admin-password-hash",
			Namespace: "wego-system",
		},
		Data: map[string][]byte{
			"password": hashed,
		},
	}

	fakeKubernetesClient := ctrlclientfake.NewClientBuilder().WithObjects(hashedSecret).Build()

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	if err != nil {
		t.Errorf("failed to create HMAC signer: %v", err)
	}

	s, _ := makeAuthServer(t, fakeKubernetesClient, tokenSignerVerifier)

	login := auth.LoginRequest{
		Password: "wrong",
	}

	j, err := json.Marshal(login)
	if err != nil {
		t.Errorf("failed to marshal to JSON: %v", err)
	}

	reader := bytes.NewReader(j)

	req := httptest.NewRequest(http.MethodPost, "https://example.com/signin", reader)
	w := httptest.NewRecorder()
	s.SignIn().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status to be 401 but got %v instead", resp.StatusCode)
	}
}

func TestSingInCorrectPassword(t *testing.T) {
	password := "my-secret-password"

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Errorf("failed to generate a hash from password: %v", err)
	}

	hashedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "admin-password-hash",
			Namespace: "wego-system",
		},
		Data: map[string][]byte{
			"password": hashed,
		},
	}

	fakeKubernetesClient := ctrlclientfake.NewClientBuilder().WithObjects(hashedSecret).Build()

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	if err != nil {
		t.Errorf("failed to create HMAC signer: %v", err)
	}

	s, _ := makeAuthServer(t, fakeKubernetesClient, tokenSignerVerifier)

	login := auth.LoginRequest{
		Password: password,
	}

	j, err := json.Marshal(login)
	if err != nil {
		t.Errorf("failed to marshal to JSON: %v", err)
	}

	reader := bytes.NewReader(j)

	req := httptest.NewRequest(http.MethodPost, "https://example.com/signin", reader)
	w := httptest.NewRecorder()
	s.SignIn().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status to be 200 but got %v instead", resp.StatusCode)
	}

	var cookie *http.Cookie

	for _, c := range resp.Cookies() {
		if c.Name == auth.IDTokenCookieName {
			cookie = c
			break
		}
	}

	if cookie == nil {
		t.Errorf("expected to find cookie %q but did not", auth.IDTokenCookieName)
	}

	if _, err := tokenSignerVerifier.Verify(cookie.Value); err != nil {
		t.Errorf("expected to verify the issued token but got an error instead: %v", err)
	}
}

func makeAuthServer(t *testing.T, client ctrlclient.Client, tokenSignerVerifier auth.TokenSignerVerifier) (*auth.AuthServer, *mockoidc.MockOIDC) {
	t.Helper()

	m, err := mockoidc.Run()
	if err != nil {
		t.Errorf("failed to create mock OIDC server: %v", err)
	}

	t.Cleanup(func() {
		_ = m.Shutdown()
	})

	s, err := auth.NewAuthServer(context.Background(), logr.Discard(), http.DefaultClient, auth.AuthConfig{
		OIDCConfig: auth.OIDCConfig{
			ClientID:     m.Config().ClientID,
			ClientSecret: m.Config().ClientSecret,
			IssuerURL:    m.Config().Issuer,
		},
	}, client, tokenSignerVerifier)
	if err != nil {
		t.Errorf("failed to create a new AuthServer instance: %v", err)
	}

	return s, m
}

// Add tests for verifying the token on the Userinfo handler

// Create middleware to validate token (whether OIDC or superuser)
