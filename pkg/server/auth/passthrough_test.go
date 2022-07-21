package auth_test

import (
	"context"
	"testing"

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
		{"JWT ID Token", token, &auth.UserPrincipal{ID: "example@example.com", Groups: []string{"testing"}, Token: token}},
		{"no cookie value", "", nil},
	}

	srv := testutils.MakeKeysetServer(t, privKey)
	keySet := oidc.NewRemoteKeySet(oidc.ClientContext(context.TODO(), srv.Client()), srv.URL)
	verifier := oidc.NewVerifier("http://127.0.0.1:5556/dex", keySet, &oidc.Config{ClientID: "test-service"})

	for _, tt := range authTests {
		t.Run(tt.name, func(t *testing.T) {
			principal, err := auth.NewJWTPassthroughCookiePrincipalGetter(logr.Discard(), verifier, cookieName).Principal(makeCookieRequest(cookieName, tt.cookie))
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, principal); diff != "" {
				t.Fatalf("failed to get principal:\n%s", diff)
			}
		})
	}
}
