package auth_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	. "github.com/onsi/gomega"

	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
)

func TestJWTCookiePrincipalGetter(t *testing.T) {
	const cookieName = "auth-token"

	privKey := testutils.MakeRSAPrivateKey(t)
	authTests := []struct {
		name         string
		cookie       string
		claimsConfig *auth.ClaimsConfig
		want         *auth.UserPrincipal
	}{
		{
			name:   "JWT ID Token",
			cookie: testutils.MakeJWToken(t, privKey, "example@example.com"),
			want:   &auth.UserPrincipal{ID: "example@example.com", Groups: []string{"testing"}},
		},
		{
			name: "Custom user and groups claim",
			cookie: testutils.MakeJWToken(t, privKey, "example@example.com", func(m map[string]any) {
				m["demo_groups"] = []string{"group1", "group2"}
			}),
			claimsConfig: &auth.ClaimsConfig{Username: "sub", Groups: "demo_groups"},
			want:         &auth.UserPrincipal{ID: "testing", Groups: []string{"group1", "group2"}},
		},
		{"no cookie value", "", nil, nil},
	}

	srv := testutils.MakeKeysetServer(t, privKey)
	keySet := oidc.NewRemoteKeySet(oidc.ClientContext(t.Context(), srv.Client()), srv.URL)
	verifier := oidc.NewVerifier("http://127.0.0.1:5556/dex", keySet, &oidc.Config{ClientID: "test-service"})

	for _, tt := range authTests {
		t.Run(tt.name, func(t *testing.T) {
			sessionManager := scs.New()
			sessionManager.Lifetime = 500 * time.Millisecond
			getter := auth.NewJWTCookiePrincipalGetter(logr.Discard(), verifier, tt.claimsConfig, cookieName, sessionManager)

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

func TestJWTAuthorizationHeaderPrincipalGetter(t *testing.T) {
	privKey := testutils.MakeRSAPrivateKey(t)
	authTests := []struct {
		name          string
		authorization string
		claimsConfig  *auth.ClaimsConfig
		want          *auth.UserPrincipal
	}{
		{name: "JWT ID Token", authorization: "Bearer " + testutils.MakeJWToken(t, privKey, "example@example.com"), want: &auth.UserPrincipal{ID: "example@example.com", Groups: []string{"testing"}}},
		{
			name:          "Custom user claim",
			authorization: "Bearer " + testutils.MakeJWToken(t, privKey, "example@example.com"),
			claimsConfig:  &auth.ClaimsConfig{Username: "sub"},
			want:          &auth.UserPrincipal{ID: "testing", Groups: []string{"testing"}},
		},
		{
			name: "Custom groups claim",
			authorization: "Bearer " + testutils.MakeJWToken(t, privKey, "example@example.com", func(m map[string]any) {
				m["test_groups"] = []string{"test-group1", "test-group2"}
			}),
			claimsConfig: &auth.ClaimsConfig{Groups: "test_groups"},
			want:         &auth.UserPrincipal{ID: "example@example.com", Groups: []string{"test-group1", "test-group2"}},
		},

		{"no auth header value", "", nil, nil},
	}

	srv := testutils.MakeKeysetServer(t, privKey)
	keySet := oidc.NewRemoteKeySet(oidc.ClientContext(t.Context(), srv.Client()), srv.URL)
	verifier := oidc.NewVerifier("http://127.0.0.1:5556/dex", keySet, &oidc.Config{ClientID: "test-service"})

	for _, tt := range authTests {
		t.Run(tt.name, func(t *testing.T) {
			principal, err := auth.NewJWTAuthorizationHeaderPrincipalGetter(logr.Discard(), verifier, tt.claimsConfig).Principal(makeAuthenticatedRequest(tt.authorization))
			if err != nil {
				t.Fatal(err)
			}
			if diff := cmp.Diff(tt.want, principal, allowUnexportedPrincipal()); diff != "" {
				t.Fatalf("failed to get principal:\n%s", diff)
			}
		})
	}
}

func makeAuthenticatedRequest(token string) *http.Request {
	req := httptest.NewRequest("GET", "http://example.com/", nil)
	if token != "" {
		req.Header.Set("Authorization", token)
	}

	return req
}

func TestMultiAuth(t *testing.T) {
	g := NewGomegaWithT(t)

	err := errors.New("oops")
	noAuthError := errors.New("could not find valid principal")

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
			auths: []auth.PrincipalGetter{fakePrincipalGetter{}},
			want:  nil,
			err:   noAuthError,
		},
		{
			name:  "one successful auth",
			auths: []auth.PrincipalGetter{fakePrincipalGetter{id: "testing"}},
			want:  &auth.UserPrincipal{ID: "testing"},
		},
		{
			name:  "two auths, one unsuccessful",
			auths: []auth.PrincipalGetter{fakePrincipalGetter{}, fakePrincipalGetter{id: "testing"}},
			want:  &auth.UserPrincipal{ID: "testing"},
		},
		{
			name:  "two auths, none successful",
			auths: []auth.PrincipalGetter{fakePrincipalGetter{}, fakePrincipalGetter{}},
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

type fakePrincipalGetter struct {
	id string
}

func (s fakePrincipalGetter) Principal(r *http.Request) (*auth.UserPrincipal, error) {
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

func allowUnexportedPrincipal() cmp.Option {
	return cmp.AllowUnexported(auth.UserPrincipal{})
}

func extractTokenFromCookie(c string) string {
	parts := strings.Split(c, ";")
	return strings.SplitN(parts[0], "=", 2)[1]
}

func parseBodyAsPrincipal(t *testing.T, s string) *auth.UserPrincipal {
	t.Helper()

	if s == "" {
		return nil
	}

	var raw map[string]interface{}
	if err := json.Unmarshal([]byte(s), &raw); err != nil {
		t.Fatal(err)
	}

	groups := func(g []any) []string {
		r := []string{}
		for _, v := range g {
			r = append(r, v.(string))
		}

		return r
	}(raw["groups"].([]any))

	p := auth.UserPrincipal{
		ID:     raw["id"].(string),
		Groups: groups,
	}
	token := raw["token"].(string)
	if token != "" {
		p.SetToken(token)
	}

	return &p
}

type testServer struct {
	*httptest.Server
}

func newTestServer(t *testing.T, handler http.Handler) *testServer {
	t.Helper()
	ts := httptest.NewTLSServer(handler)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	ts.Client().Jar = jar

	ts.Client().CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}

	return &testServer{ts}
}

func (ts *testServer) execute(t *testing.T, urlPath string, cookie *http.Cookie) (http.Header, string) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, ts.URL+urlPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}

	resp, err := ts.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	return resp.Header, string(body)
}

func newTestJWTMux(t *testing.T, cookieName string, sessionManager auth.SessionManager, principalGetter auth.PrincipalGetter) *http.ServeMux {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/put", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(cookieName)
		if err != nil {
			t.Fatalf("request to /put failed because failed to get cookie: %s", err)
		}

		sessionManager.Put(r.Context(), cookieName, cookie.Value)
	}))

	mux.HandleFunc("/get", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		principal, err := principalGetter.Principal(r)
		if err != nil {
			t.Fatal(err)
		}

		if principal == nil {
			return
		}

		b, err := json.Marshal(map[string]any{
			"id":     principal.ID,
			"groups": principal.Groups,
			"token":  principal.Token(),
		})
		if err != nil {
			t.Fatal(err)
		}

		w.Write(b)
	}))

	return mux
}

func newTestJWTServerServer(t *testing.T, cookieName string, sessionManager auth.SessionManager, principalGetter auth.PrincipalGetter) *testServer {
	t.Helper()
	mux := newTestJWTMux(t, cookieName, sessionManager, principalGetter)

	ts := newTestServer(t, sessionManager.LoadAndSave(mux))
	t.Cleanup(ts.Close)

	return ts
}
