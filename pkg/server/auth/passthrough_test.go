package auth_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"

	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

func TestPassthroughPrincipalGetter(t *testing.T) {
	const cookieName = "auth-token"

	privKey := testutils.MakeRSAPrivateKey(t)
	token := testutils.MakeJWToken(t, privKey, "example@example.com")
	authTests := []struct {
		name   string
		cookie string
		want   *auth.UserPrincipal
	}{
		{"JWT ID Token", token, auth.NewUserPrincipal(auth.ID("example@example.com"), auth.Groups([]string{"testing"}), auth.Token(token))},
		{"no cookie value", "", nil},
	}

	srv := testutils.MakeKeysetServer(t, privKey)
	keySet := oidc.NewRemoteKeySet(oidc.ClientContext(t.Context(), srv.Client()), srv.URL)
	verifier := oidc.NewVerifier("http://127.0.0.1:5556/dex", keySet, &oidc.Config{ClientID: "test-service"})

	for _, tt := range authTests {
		t.Run(tt.name, func(t *testing.T) {
			sessionManager := scs.New()
			sessionManager.Lifetime = 500 * time.Millisecond
			getter := auth.NewJWTPassthroughCookiePrincipalGetter(logr.Discard(), verifier, cookieName, sessionManager)
			ts := newTestJWTServerServer(t, cookieName, sessionManager, getter)

			header, _ := ts.execute(t, "/put", &http.Cookie{Name: cookieName, Value: tt.cookie})
			sessionToken := extractTokenFromCookie(header.Get("Set-Cookie"))

			_, body := ts.execute(t, "/get", &http.Cookie{Name: "session", Value: sessionToken})

			principal := parseBodyAsPrincipal(t, body)
			if diff := cmp.Diff(tt.want, principal, allowUnexportedPrincipal()); diff != "" {
				t.Fatalf("failed to get principal:\n%s", diff)
			}
		})
	}
}
