package auth_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"gotest.tools/v3/assert"
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

	assert.NilError(t, err)

	cookie := &http.Cookie{
		Name:     "id_token",
		Value:    "foo",
		Path:     "/",
		Expires:  time.Now().UTC().Add(5),
		HttpOnly: true,
	}

	w := httptest.NewRecorder()

	http.SetCookie(w, cookie)

	req := httptest.NewRequest(http.MethodPost, "https://example.com/logout", nil)
	s.Logout().ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status to be 200 but got %v instead", resp.StatusCode)
	}

	fmt.Println(resp.Cookies())

	for _, c := range resp.Cookies() {
		if c.Name == auth.IDTokenCookieName {
			cookie = c
			break
		}
	}

	assert.Equal(t, cookie.Value, "")
}
