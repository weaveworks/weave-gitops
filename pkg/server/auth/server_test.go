package auth_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-logr/logr"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func TestLogoutSuccess(t *testing.T) {
	m, err := mockoidc.Run()
	if err != nil {
		t.Errorf("failed to create mock OIDC server: %v", err)
	}

	s, err := auth.NewAuthServer(context.Background(), logr.Discard(), http.DefaultClient, auth.AuthConfig{
		OIDCConfig: auth.OIDCConfig{
			IssuerURL: m.Config().Issuer,
		},
		CookieConfig: auth.CookieConfig{
			CookieDuration: 5,
		},
	})

	assert.Nil(t, err)

	w := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodPost, "https://example.com/logout", nil)
	s.Logout().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status to be 200 but got %v instead", resp.StatusCode)
	}

	cookie := &http.Cookie{}

	for _, c := range resp.Cookies() {
		if c.Name == auth.IDTokenCookieName {
			cookie = c
			break
		}
	}

	assert.Equal(t, cookie.Value, "")
}

func TestLogoutWithWrongMethod(t *testing.T) {
	m, err := mockoidc.Run()
	if err != nil {
		t.Errorf("failed to create mock OIDC server: %v", err)
	}

	s, err := auth.NewAuthServer(context.Background(), logr.Discard(), http.DefaultClient, auth.AuthConfig{
		OIDCConfig: auth.OIDCConfig{
			IssuerURL: m.Config().Issuer,
		},
		CookieConfig: auth.CookieConfig{
			CookieDuration: 5,
		},
	})

	assert.Nil(t, err)

	w := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "https://example.com/logout", nil)
	s.Logout().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("expected status to be 405 but got %v instead", resp.StatusCode)
	}
}
