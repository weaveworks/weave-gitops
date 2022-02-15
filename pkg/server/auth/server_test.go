package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlclientfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestSignInOnlySupportsPOST(t *testing.T) {
	methods := []string{
		http.MethodGet,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodHead,
		http.MethodOptions,
	}

	s := auth.AuthServer{}

	for _, m := range methods {
		req := httptest.NewRequest(m, "https://example.com/signin", nil)
		w := httptest.NewRecorder()
		s.SignIn().ServeHTTP(w, req)

		resp := w.Result()
		if resp.StatusCode != http.StatusMethodNotAllowed {
			t.Errorf("expected status to be 405 but got %v instead", resp.StatusCode)
		}
	}
}

func TestSignInBadRequest(t *testing.T) {
	s := auth.AuthServer{}

	req := httptest.NewRequest(http.MethodPost, "https://example.com/signin", nil)
	w := httptest.NewRecorder()
	s.SignIn().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status to be 400 but got %v instead", resp.StatusCode)
	}
}

// TODO: Update this test to check what happens when a password is not set
func TestSignInWrongPassword(t *testing.T) {
	s := auth.AuthServer{}

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

	m, err := mockoidc.Run()
	if err != nil {
		t.Errorf("failed to create mock OIDC server: %v", err)
	}

	defer m.Shutdown()

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

	s, err := auth.NewAuthServer(context.Background(), logr.Discard(), http.DefaultClient, auth.AuthConfig{
		OIDCConfig: auth.OIDCConfig{
			IssuerURL: m.Config().Issuer,
		},
	}, fakeKubernetesClient)
	if err != nil {
		t.Errorf("failed to create a new AuthServer instance: %v", err)
	}

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

	// ensure that a JWT token is issues in an id_token cookie
}

// Add tests for verifying the token on the Userinfo handler

// Create middleware to validate token (whether OIDC or superuser)

// Return a { "email": "admin", "groups": [] } object from Userinfo
