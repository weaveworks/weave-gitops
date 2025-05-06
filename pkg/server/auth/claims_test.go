package auth_test

import (
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

func TestPrincipalFromClaims(t *testing.T) {
	privKey := testutils.MakeRSAPrivateKey(t)
	principalTests := []struct {
		name   string
		token  string
		config *auth.ClaimsConfig
		want   *auth.UserPrincipal
	}{
		{
			name:   "simple user with groups",
			token:  testutils.MakeJWToken(t, privKey, "example@example.com"),
			config: &auth.ClaimsConfig{},
			want: &auth.UserPrincipal{
				ID:     "example@example.com",
				Groups: []string{"testing"},
			},
		},
		{
			name:   "custom user claim",
			token:  testutils.MakeJWToken(t, privKey, "example@example.com"),
			config: &auth.ClaimsConfig{Username: "sub"},
			want:   &auth.UserPrincipal{ID: "testing", Groups: []string{"testing"}},
		},
		{
			name: "custom groups claim",
			token: testutils.MakeJWToken(t, privKey, "example@example.com", func(m map[string]any) {
				m["test_groups"] = []string{"new-group1", "new-group2"}
			}),
			config: &auth.ClaimsConfig{Groups: "test_groups"},
			want:   &auth.UserPrincipal{ID: "example@example.com", Groups: []string{"new-group1", "new-group2"}},
		},
		{
			name: "single string custom group claim",
			token: testutils.MakeJWToken(t, privKey, "example@example.com", func(m map[string]any) {
				m["test_groups"] = "single-group"
			}),
			config: &auth.ClaimsConfig{Groups: "test_groups"},
			want:   &auth.UserPrincipal{ID: "example@example.com", Groups: []string{"single-group"}},
		},
	}

	srv := testutils.MakeKeysetServer(t, privKey)
	keySet := oidc.NewRemoteKeySet(oidc.ClientContext(t.Context(), srv.Client()), srv.URL)
	verifier := oidc.NewVerifier("http://127.0.0.1:5556/dex", keySet, &oidc.Config{ClientID: "test-service"})

	for _, tt := range principalTests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := verifier.Verify(t.Context(), tt.token)
			if err != nil {
				t.Fatal(err)
			}

			principal, err := tt.config.PrincipalFromClaims(token)
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, principal, cmpopts.IgnoreUnexported(auth.UserPrincipal{})); diff != "" {
				t.Fatalf("failed to generate principal:\n%s", diff)
			}
		})
	}
}
