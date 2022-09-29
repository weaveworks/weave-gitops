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

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-logr/logr"
	"github.com/oauth2-proxy/mockoidc"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const testNamespace = "flux-system"

func TestWithAPIAuthReturns401ForUnauthenticatedRequests(t *testing.T) {
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

	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), oidcCfg, fakeKubernetesClient, tokenSignerVerifier, testNamespace, authMethods)
	g.Expect(err).NotTo(HaveOccurred())

	srv, err := auth.NewAuthServer(context.Background(), authCfg)
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
	auth.WithAPIAuth(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}), srv, nil).ServeHTTP(res, req)

	g.Expect(res).To(HaveHTTPStatus(http.StatusUnauthorized))

	// Test out the publicRoutes
	res = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, s.URL+"/v1/featureflags", nil)
	auth.WithAPIAuth(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}), srv, []string{"/v1/featureflags"}).ServeHTTP(res, req)

	g.Expect(res).To(HaveHTTPStatus(http.StatusOK))
}

func TestWithAPIAuthOnlyUsesValidMethods(t *testing.T) {
	// In theory all attempts to login in this should fail as, despite
	// the auth server having access to
	g := NewGomegaWithT(t)

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

	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), oidcCfg, fakeKubernetesClient, tokenSignerVerifier, testNamespace, authMethods)
	g.Expect(err).NotTo(HaveOccurred())

	srv, err := auth.NewAuthServer(context.Background(), authCfg)
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
	auth.WithAPIAuth(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}), srv, nil).ServeHTTP(res, req)

	g.Expect(res).To(HaveHTTPStatus(http.StatusUnauthorized))

	// Try logging in via the static user
	// res1, err := http.Post(s.URL+"/oauth2/sign_in", "application/json", bytes.NewReader([]byte(`{"password":"my-secret-password"}`)))
	res1, err := http.Post(s.URL+"/oauth2/sign_in", "application/json", bytes.NewReader([]byte(`{"password":"bad-password"}`)))

	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res1).To(HaveHTTPStatus(http.StatusUnauthorized))

	// Test out the publicRoutes
	res = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, s.URL+"/v1/featureflags", nil)
	auth.WithAPIAuth(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {}), srv, []string{"/v1/featureflags"}).ServeHTTP(res, req)

	g.Expect(res).To(HaveHTTPStatus(http.StatusOK))
}

func TestOauth2FlowRedirectsToOIDCIssuerForUnauthenticatedRequests(t *testing.T) {
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

	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), oidcCfg, fakeKubernetesClient, tokenSignerVerifier, testNamespace, authMethods)
	g.Expect(err).NotTo(HaveOccurred())

	srv, err := auth.NewAuthServer(context.Background(), authCfg)
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

	authCodeURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s", m.AuthorizationEndpoint(), fake.ClientID, url.QueryEscape(redirectURL), strings.Join([]string{"profile", oidc.ScopeOpenID, "email"}, "+"))
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

	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), oidcCfg, fakeKubernetesClient, tokenSignerVerifier, testNamespace, authMethods)
	g.Expect(err).NotTo(HaveOccurred())

	srv, err := auth.NewAuthServer(context.Background(), authCfg)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(auth.RegisterAuthServer(mux, "/oauth2", srv, 1)).To(Succeed())

	s := httptest.NewServer(mux)

	t.Cleanup(func() {
		s.Close()
	})

	res1, err := http.Post(s.URL+"/oauth2/sign_in", "application/json", bytes.NewReader([]byte(`{"password":"my-secret-password"}`)))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res1).To(HaveHTTPStatus(http.StatusOK))

	res2, err := http.Post(s.URL+"/oauth2/sign_in", "application/json", bytes.NewReader([]byte(`{"password":"my-secret-password"}`)))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res2).To(HaveHTTPStatus(http.StatusTooManyRequests))

	time.Sleep(time.Second)

	res3, err := http.Post(s.URL+"/oauth2/sign_in", "application/json", bytes.NewReader([]byte(`{"password":"my-secret-password"}`)))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res3).To(HaveHTTPStatus(http.StatusOK))

	time.Sleep(time.Second)

	res4, err := http.Post(s.URL+"/oauth2/sign_in", "application/json", bytes.NewReader([]byte(`{"password":"bad-password"}`)))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(res4).To(HaveHTTPStatus(http.StatusUnauthorized))
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

	want := `id="testing" groups=[group1 group2]`
	if s := p.String(); s != want {
		t.Fatalf("principal.String() got %s, want %s", s, want)
	}
}

func TestUserPrincipal_Hash(t *testing.T) {
	p := auth.NewUserPrincipal(auth.ID("testing"), auth.Groups([]string{"group1", "group2"}), auth.Token("test-token"))

	want := "f6b28168aaeae03685db1e9151a397a8"
	if s := p.Hash(); s != want {
		t.Fatalf("principal.String() got %s, want %s", s, want)
	}
}
