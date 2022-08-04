package auth_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

func TestJWTCookiePrincipalGetter(t *testing.T) {
	const cookieName = "auth-token"

	privKey := testutils.MakeRSAPrivateKey(t)
	authTests := []struct {
		name   string
		cookie string
		want   *auth.UserPrincipal
	}{
		{"JWT ID Token", testutils.MakeJWToken(t, privKey, "example@example.com"), &auth.UserPrincipal{ID: "example@example.com", Groups: []string{"testing"}}},
		{"no cookie value", "", nil},
	}

	srv := testutils.MakeKeysetServer(t, privKey)
	keySet := oidc.NewRemoteKeySet(oidc.ClientContext(context.TODO(), srv.Client()), srv.URL)
	verifier := oidc.NewVerifier("http://127.0.0.1:5556/dex", keySet, &oidc.Config{ClientID: "test-service"})

	for _, tt := range authTests {
		t.Run(tt.name, func(t *testing.T) {
			principal, err := auth.NewJWTCookiePrincipalGetter(logr.Discard(), verifier, cookieName).Principal(makeCookieRequest(cookieName, tt.cookie))
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, principal); diff != "" {
				t.Fatalf("failed to get principal:\n%s", diff)
			}
		})
	}
}

func TestJWTAuthorizationHeaderPrincipalGetter(t *testing.T) {
	privKey := testutils.MakeRSAPrivateKey(t)
	authTests := []struct {
		name          string
		authorization string
		want          *auth.UserPrincipal
	}{
		{"JWT ID Token", "Bearer " + testutils.MakeJWToken(t, privKey, "example@example.com"), &auth.UserPrincipal{ID: "example@example.com", Groups: []string{"testing"}}},
		{"no auth header value", "", nil},
	}

	srv := testutils.MakeKeysetServer(t, privKey)
	keySet := oidc.NewRemoteKeySet(oidc.ClientContext(context.TODO(), srv.Client()), srv.URL)
	verifier := oidc.NewVerifier("http://127.0.0.1:5556/dex", keySet, &oidc.Config{ClientID: "test-service"})

	for _, tt := range authTests {
		t.Run(tt.name, func(t *testing.T) {
			principal, err := auth.NewJWTAuthorizationHeaderPrincipalGetter(logr.Discard(), verifier).Principal(makeAuthenticatedRequest(tt.authorization))
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, principal); diff != "" {
				t.Fatalf("failed to get principal:\n%s", diff)
			}
		})
	}
}

func makeCookieRequest(cookieName, token string) *http.Request {
	req := httptest.NewRequest("GET", "http://example.com/", nil)
	if token != "" {
		req.AddCookie(&http.Cookie{
			Name:  cookieName,
			Value: token,
		})
	}

	return req
}

func makeAuthenticatedRequest(token string) *http.Request {
	req := httptest.NewRequest("GET", "http://example.com/", nil)
	if token != "" {
		req.Header.Set("Authorization", token)
	}

	return req
}

func TestMultiAuth(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	err := errors.New("oops")
	noAuthError := errors.New("Could not find valid principal")

	multiAuthTests := []struct {
		name  string
		auths []auth.PrincipalGetter
		want  *auth.UserPrincipal
		err   error
	}{
		{
			name:  "no auths",
			auths: []auth.PrincipalGetter{},
			want:  nil,
			err:   noAuthError,
		},
		{
			name:  "no successful auths",
			auths: []auth.PrincipalGetter{stubPrincipalGetter{}},
			want:  nil,
			err:   noAuthError,
		},
		{
			name:  "one successful auth",
			auths: []auth.PrincipalGetter{stubPrincipalGetter{id: "testing"}},
			want:  &auth.UserPrincipal{ID: "testing"},
		},
		{
			name:  "two auths, one unsuccessful",
			auths: []auth.PrincipalGetter{stubPrincipalGetter{}, stubPrincipalGetter{id: "testing"}},
			want:  &auth.UserPrincipal{ID: "testing"},
		},
		{
			name:  "two auths, none successful",
			auths: []auth.PrincipalGetter{stubPrincipalGetter{}, stubPrincipalGetter{}},
			want:  nil,
			err:   noAuthError,
		},
		{
			name:  "error",
			auths: []auth.PrincipalGetter{errorPrincipalGetter{err: err}},
			want:  nil,
			err:   err,
		},
	}

	for _, tt := range multiAuthTests {
		t.Run(tt.name, func(t *testing.T) {
			mg := auth.MultiAuthPrincipal{Log: logr.Discard(), Getters: tt.auths}
			req := httptest.NewRequest("GET", "http://example.com/", nil)

			principal, err := mg.Principal(req)

			if tt.err != nil {
				g.Expect(err).To(MatchError(tt.err))
			}

			g.Expect(principal).To(Equal(tt.want))
		})
	}
}

type stubPrincipalGetter struct {
	id string
}

func (s stubPrincipalGetter) Principal(r *http.Request) (*auth.UserPrincipal, error) {
	if s.id != "" {
		return &auth.UserPrincipal{ID: s.id}, nil
	}

	return nil, nil
}

type errorPrincipalGetter struct {
	err error
}

func (s errorPrincipalGetter) Principal(r *http.Request) (*auth.UserPrincipal, error) {
	return nil, s.err
}
