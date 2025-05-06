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

	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-logr/logr"
	"github.com/oauth2-proxy/mockoidc"
	. "github.com/onsi/gomega"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

const testNamespace = "flux-system"

func TestWithAPIAuthReturns401ForUnauthenticatedRequests(t *testing.T) {
	sm := scs.New()
	g := NewGomegaWithT(t)

	m, err := mockoidc.Run()
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		_ = m.Shutdown()
	})

	fake := m.Config()
	mux := http.NewServeMux()
	fakeKubernetesClient := ctrlclient.NewClientBuilder().Build()

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	oidcCfg := auth.OIDCConfig{
		ClientID:     fake.ClientID,
		ClientSecret: fake.ClientSecret,
		IssuerURL:    fake.Issuer,
	}

	authMethods := map[auth.AuthMethod]bool{auth.OIDC: true}

	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), oidcCfg, fakeKubernetesClient, tokenSignerVerifier, testNamespace, authMethods, "", sm)
	g.Expect(err).NotTo(HaveOccurred())

	srv, err := auth.NewAuthServer(t.Context(), authCfg)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(auth.RegisterAuthServer(mux, "/oauth2", srv, 1)).To(Succeed())

	s := httptest.NewServer(mux)

	t.Cleanup(func() {
		s.Close()
	})

	// Set the correct redirect URL now that we have a server running
	srv.SetRedirectURL(s.URL + "/oauth2/callback")

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, s.URL, nil)
	handler := sm.LoadAndSave(auth.WithAPIAuth(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}), srv, nil, sm))

	handler.ServeHTTP(res, req)

	g.Expect(res).To(HaveHTTPStatus(http.StatusUnauthorized))

	// Test out the publicRoutes
	res = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, s.URL+"/v1/featureflags", nil)
	auth.WithAPIAuth(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}), srv, []string{"/v1/featureflags"}, sm).ServeHTTP(res, req)

	g.Expect(res).To(HaveHTTPStatus(http.StatusOK))
}

func TestAnonymousAuth(t *testing.T) {
	g := NewGomegaWithT(t)

	authMethods := map[auth.AuthMethod]bool{auth.Anonymous: true}
	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), auth.OIDCConfig{}, nil, nil, testNamespace, authMethods, "test-user", nil)
	g.Expect(err).NotTo(HaveOccurred())

	srv, err := auth.NewAuthServer(t.Context(), authCfg)
	g.Expect(err).NotTo(HaveOccurred())

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	auth.WithAPIAuth(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		// no cookie checking etc, principal is just there ready to go
		g.Expect(auth.Principal(r.Context()).ID).To(Equal("test-user"))
	}), srv, nil, nil).ServeHTTP(res, req)
}

func TestWithAPIAuthOnlyUsesValidMethods(t *testing.T) {
	// In theory all attempts to login in this should fail as, despite
	// the auth server having access to
	g := NewGomegaWithT(t)
	sm := scs.New()

	m, err := mockoidc.Run()
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		_ = m.Shutdown()
	})

	fake := m.Config()
	mux := http.NewServeMux()

	password := "my-secret-password"
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	g.Expect(err).NotTo(HaveOccurred())

	hashedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-user-auth",
			Namespace: "flux-system",
		},
		Data: map[string][]byte{
			"password": hashed,
		},
	}

	fakeKubernetesClient := ctrlclient.NewClientBuilder().WithObjects(hashedSecret).Build()

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	oidcCfg := auth.OIDCConfig{
		ClientID:     fake.ClientID,
		ClientSecret: fake.ClientSecret,
		IssuerURL:    fake.Issuer,
	}

	authMethods := map[auth.AuthMethod]bool{} // This is not a valid AuthMethod

	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), oidcCfg, fakeKubernetesClient, tokenSignerVerifier, testNamespace, authMethods, "", sm)
	g.Expect(err).NotTo(HaveOccurred())

	srv, err := auth.NewAuthServer(t.Context(), authCfg)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(auth.RegisterAuthServer(mux, "/oauth2", srv, 1)).To(Succeed())

	s := httptest.NewServer(mux)

	t.Cleanup(func() {
		s.Close()
	})

	// Set the correct redirect URL now that we have a server running
	srv.SetRedirectURL(s.URL + "/oauth2/callback")

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, s.URL, nil)
	auth.WithAPIAuth(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}), srv, nil, scs.New()).ServeHTTP(res, req)

	g.Expect(res).To(HaveHTTPStatus(http.StatusUnauthorized))

	// Try logging in via the static user
	res1, err := http.Post(s.URL+"/oauth2/sign_in", "application/json", bytes.NewReader([]byte(`{"password":"bad-password"}`)))

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res1).To(HaveHTTPStatus(http.StatusUnauthorized))

	// Test out the publicRoutes
	res = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, s.URL+"/v1/featureflags", nil)
	auth.WithAPIAuth(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}), srv, []string{"/v1/featureflags"}, scs.New()).ServeHTTP(res, req)

	g.Expect(res).To(HaveHTTPStatus(http.StatusOK))
}

func TestOauth2FlowRedirectsToOIDCIssuerWithCustomScopes(t *testing.T) {
	g := NewGomegaWithT(t)
	sm := &fakeSessionManager{}

	m, err := mockoidc.Run()
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		_ = m.Shutdown()
	})

	fake := m.Config()
	mux := http.NewServeMux()
	fakeKubernetesClient := ctrlclient.NewClientBuilder().Build()

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	oidcCfg := auth.OIDCConfig{
		ClientID:     fake.ClientID,
		ClientSecret: fake.ClientSecret,
		IssuerURL:    fake.Issuer,
		ClaimsConfig: &auth.ClaimsConfig{Username: "email", Groups: "groups"},
		Scopes:       []string{"test1", "test2"},
	}

	authMethods := map[auth.AuthMethod]bool{auth.OIDC: true}

	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), oidcCfg, fakeKubernetesClient, tokenSignerVerifier, testNamespace, authMethods, "", sm)
	g.Expect(err).NotTo(HaveOccurred())

	srv, err := auth.NewAuthServer(t.Context(), authCfg)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(auth.RegisterAuthServer(mux, "/oauth2", srv, 1)).To(Succeed())

	s := httptest.NewServer(mux)

	t.Cleanup(s.Close)

	// Set the correct redirect URL now that we have a server running
	redirectURL := s.URL + "/oauth2/callback"
	srv.SetRedirectURL(redirectURL)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, s.URL, nil)
	srv.OAuth2Flow().ServeHTTP(res, req)

	g.Expect(res).To(HaveHTTPStatus(http.StatusSeeOther))

	authCodeURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s", m.AuthorizationEndpoint(), fake.ClientID, url.QueryEscape(redirectURL), strings.Join([]string{oidc.ScopeOpenID, "test1", "test2"}, "+"))
	g.Expect(res.Result().Header.Get("Location")).To(ContainSubstring(authCodeURL))
}

func TestOauth2FlowRedirectsToOIDCIssuerForUnauthenticatedRequests(t *testing.T) {
	g := NewGomegaWithT(t)
	sm := &fakeSessionManager{}

	m, err := mockoidc.Run()
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		_ = m.Shutdown()
	})

	fake := m.Config()
	mux := http.NewServeMux()
	fakeKubernetesClient := ctrlclient.NewClientBuilder().Build()

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	oidcCfg := auth.OIDCConfig{
		ClientID:     fake.ClientID,
		ClientSecret: fake.ClientSecret,
		IssuerURL:    fake.Issuer,
		ClaimsConfig: &auth.ClaimsConfig{Username: "email", Groups: "groups"},
		Scopes:       auth.DefaultScopes,
	}

	authMethods := map[auth.AuthMethod]bool{auth.OIDC: true}

	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), oidcCfg, fakeKubernetesClient, tokenSignerVerifier, testNamespace, authMethods, "", sm)
	g.Expect(err).NotTo(HaveOccurred())

	srv, err := auth.NewAuthServer(t.Context(), authCfg)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(auth.RegisterAuthServer(mux, "/oauth2", srv, 1)).To(Succeed())

	s := httptest.NewServer(mux)

	t.Cleanup(func() {
		s.Close()
	})

	// Set the correct redirect URL now that we have a server running
	redirectURL := s.URL + "/oauth2/callback"
	srv.SetRedirectURL(redirectURL)

	res := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, s.URL, nil)
	srv.OAuth2Flow().ServeHTTP(res, req)

	g.Expect(res).To(HaveHTTPStatus(http.StatusSeeOther))

	authCodeURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s", m.AuthorizationEndpoint(), fake.ClientID, url.QueryEscape(redirectURL), strings.Join([]string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, auth.ScopeEmail, auth.ScopeGroups}, "+"))
	g.Expect(res.Result().Header.Get("Location")).To(ContainSubstring(authCodeURL))
}

func TestIsPublicRoute(t *testing.T) {
	g := NewGomegaWithT(t)

	g.Expect(auth.IsPublicRoute(&url.URL{Path: "/foo"}, []string{"/foo"})).To(BeTrue())
	g.Expect(auth.IsPublicRoute(&url.URL{Path: "foo"}, []string{"/foo"})).To(BeFalse())
	g.Expect(auth.IsPublicRoute(&url.URL{Path: "/foob"}, []string{"/foo"})).To(BeFalse())
}

func TestRateLimit(t *testing.T) {
	g := NewGomegaWithT(t)
	sm := &fakeSessionManager{}

	mux := http.NewServeMux()
	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	password := "my-secret-password"
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	g.Expect(err).NotTo(HaveOccurred())

	hashedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-user-auth",
			Namespace: "flux-system",
		},
		Data: map[string][]byte{
			"password": hashed,
		},
	}
	fakeKubernetesClient := ctrlclient.NewClientBuilder().WithObjects(hashedSecret).Build()

	oidcCfg := auth.OIDCConfig{}

	authMethods := map[auth.AuthMethod]bool{auth.UserAccount: true}

	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), oidcCfg, fakeKubernetesClient, tokenSignerVerifier, testNamespace, authMethods, "", sm)
	g.Expect(err).NotTo(HaveOccurred())

	srv, err := auth.NewAuthServer(t.Context(), authCfg)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(auth.RegisterAuthServer(mux, "/oauth2", srv, 1)).To(Succeed())

	s := httptest.NewServer(mux)

	t.Cleanup(func() {
		s.Close()
	})

	resp, err := http.Post(s.URL+"/oauth2/sign_in", "application/json", bytes.NewReader([]byte(`{"password":"bad-password"}`)))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(resp).To(HaveHTTPStatus(http.StatusUnauthorized))

	goodSignIn := func() (*http.Response, error) {
		return http.Post(s.URL+"/oauth2/sign_in", "application/json", bytes.NewReader([]byte(`{"password":"my-secret-password"}`)))
	}
	g.Eventually(goodSignIn).Should(HaveHTTPStatus(http.StatusTooManyRequests))
	time.Sleep(time.Second)
	g.Expect(goodSignIn()).To(HaveHTTPStatus(http.StatusOK))
}

func TestUserPrincipalValid(t *testing.T) {
	g := NewGomegaWithT(t)

	tests := []struct {
		name    string
		user    *auth.UserPrincipal
		isValid bool
	}{
		{
			name:    "Full valid OIDC user",
			user:    auth.NewUserPrincipal(auth.ID("Jane"), auth.Groups([]string{"team-a", "qa"})),
			isValid: true,
		},
		{
			name:    "any token",
			user:    auth.NewUserPrincipal(auth.Token("abcdefghi09123")),
			isValid: true,
		},
		{
			name:    "Just a user id",
			user:    auth.NewUserPrincipal(auth.ID("Samir")),
			isValid: true,
		},
		{
			name:    "Empty",
			user:    auth.NewUserPrincipal(),
			isValid: false,
		},
		{
			name:    "Group only",
			user:    auth.NewUserPrincipal(auth.Groups([]string{"team-b"})),
			isValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := tt.user.Valid()

			g.Expect(res).To(Equal(tt.isValid))
		})
	}
}

func TestUserPrincipal_String(t *testing.T) {
	// This is primarily to guard against leaking the auth token if the
	// principal is logged out.
	p := auth.NewUserPrincipal(auth.ID("testing"), auth.Groups([]string{"group1", "group2"}), auth.Token("test-token"))

	if s := p.String(); s != `id="testing" groups=[group1 group2]` {
		t.Fatalf("principal.String() got %s, want %s", s, `id="testing" groups=[group1 group2]`)
	}
}

type sessionsCtxKey struct{}

// Use the fakeSessionManager for cases where you want to pass in an
// *http.Request to a handler rather than routing through a Mux.
var _ auth.SessionManager = (*fakeSessionManager)(nil)

func contextWithSessionValues(values map[string]any) context.Context {
	return contextWithValues(context.TODO(), values)
}

func contextWithValues(ctx context.Context, values map[string]any) context.Context {
	return context.WithValue(ctx, sessionsCtxKey{}, values)
}

type fakeSessionManager struct {
	// Things that are put into the context are actually stored here
	// They would normally be output in the `LoadAndSave` middleware by the
	// user's session cookie
	PutValues map[string]any
	// Record the IDs of destroyed sessions (taken from the sessionid in the
	// context).
	Destroyed []string
}

func (sm *fakeSessionManager) stringValue(name string) string {
	v, ok := sm.PutValues[name]
	if !ok {
		return ""
	}
	return v.(string)
}

func (sm *fakeSessionManager) LoadAndSave(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "Not Implemented", http.StatusNotImplemented)
	})
}

func (sm *fakeSessionManager) GetString(ctx context.Context, key string) string {
	values, ok := ctx.Value(sessionsCtxKey{}).(map[string]any)
	if !ok {
		values = map[string]any{}
	}

	v, ok := values[key]
	if ok {
		return v.(string)
	}

	return ""
}

func (sm *fakeSessionManager) Remove(context.Context, string) {
	panic("not implemented")
}

func (sm *fakeSessionManager) Put(ctx context.Context, key string, val interface{}) {
	if sm.PutValues == nil {
		sm.PutValues = map[string]any{}
	}

	sm.PutValues[key] = val
}

func (sm *fakeSessionManager) Destroy(ctx context.Context) error {
	sid := sm.GetString(ctx, "sessionid")
	if sid != "" {
		sm.Destroyed = append(sm.Destroyed, sid)
	}

	return nil
}
