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
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestWithAPIAuthReturns401ForUnauthenticatedRequests(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

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

	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), oidcCfg, fakeKubernetesClient, tokenSignerVerifier)
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

func TestOauth2FlowRedirectsToOIDCIssuerForUnauthenticatedRequests(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

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

	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), oidcCfg, fakeKubernetesClient, tokenSignerVerifier)
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
	g := gomega.NewGomegaWithT(t)

	g.Expect(auth.IsPublicRoute(&url.URL{Path: "/foo"}, []string{"/foo"})).To(BeTrue())
	g.Expect(auth.IsPublicRoute(&url.URL{Path: "foo"}, []string{"/foo"})).To(BeFalse())
	g.Expect(auth.IsPublicRoute(&url.URL{Path: "/foob"}, []string{"/foo"})).To(BeFalse())
}

func TestRateLimit(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

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

	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), oidcCfg, fakeKubernetesClient, tokenSignerVerifier)
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
}

func TestInitAuthServer(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m, err := mockoidc.Run()
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		_ = m.Shutdown()
	})

	initTests := []struct{
		name string
		secrets []*corev1.Secret
		cliOIDCConfig auth.OIDCConfig
		oidcSecretName string
		expectErr bool
		clusterUserFlag string
		oidcEnabledFlag string
	}{
		{
			name: "basic test",
			secrets: []*corev1.Secret{
				makeOIDCSecret(m.Config(), auth.DefaultOIDCAuthSecretName),
				makeClusterUserSecret("my-secret-password", auth.ClusterUserAuthSecretName),
			},
			cliOIDCConfig: auth.OIDCConfig{},
			oidcSecretName: auth.DefaultOIDCAuthSecretName,
			expectErr: false,
			clusterUserFlag: "true",
			oidcEnabledFlag: "true",
		},
		{
			name: "OIDC Only",
			secrets: []*corev1.Secret{
				makeOIDCSecret(m.Config(), auth.DefaultOIDCAuthSecretName),
			},
			cliOIDCConfig: auth.OIDCConfig{},
			oidcSecretName: auth.DefaultOIDCAuthSecretName,
			expectErr: false,
			clusterUserFlag: "false",
			oidcEnabledFlag: "true",
		},
		{
			name: "OIDC alt-secret",
			secrets: []*corev1.Secret{
				makeOIDCSecret(m.Config(), "alternate-oidc-secret"),
			},
			cliOIDCConfig: auth.OIDCConfig{},
			oidcSecretName: "alternate-oidc-secret",
			expectErr: false,
			clusterUserFlag: "false",
			oidcEnabledFlag: "true",
		},
		{
			name: "OIDC via CLI",
			secrets: []*corev1.Secret{},
			cliOIDCConfig: auth.OIDCConfig{
				IssuerURL: m.Config().Issuer,
				ClientID: m.Config().ClientID,
				ClientSecret: m.Config().ClientSecret,
				RedirectURL: "example.invalid/oauth2/callback",
			},
			oidcSecretName: auth.DefaultOIDCAuthSecretName,
			expectErr: false,
			clusterUserFlag: "false",
			oidcEnabledFlag: "true",
		},
		{
			name: "User only",
			secrets: []*corev1.Secret{
				makeClusterUserSecret("my-secret-password", auth.ClusterUserAuthSecretName),
			},
			cliOIDCConfig: auth.OIDCConfig{},
			oidcSecretName: auth.DefaultOIDCAuthSecretName,
			expectErr: false,
			clusterUserFlag: "true",
			oidcEnabledFlag: "false",
		},
	}

	for _, tt := range initTests {
		t.Run( tt.name, func(t *testing.T){
			partialKubernetesClient := ctrlclient.NewClientBuilder()

			// This is because I can't (be bothered to) figure out how to []*secret -> []*client.Object
			for _, obj := range tt.secrets {
				partialKubernetesClient.WithObjects(obj)
			}

			fakeKubernetesClient := partialKubernetesClient.Build()

			srv, err := auth.InitAuthServer(context.Background(), logr.Discard(), fakeKubernetesClient, tt.cliOIDCConfig, tt.oidcSecretName, "", false)

			if tt.expectErr {
				g.Expect(err).To(HaveOccurred())
				g.Expect(srv).To(BeNil())
			} else {
				g.Expect(err).NotTo(HaveOccurred())
				g.Expect(srv).NotTo(BeNil())
			}

			// Can check (somewhat) what's happened by making sure flags are set
			g.Expect(featureflags.Get(auth.FeatureFlagClusterUser)).To(Equal(tt.clusterUserFlag))
			g.Expect(featureflags.Get(auth.FeatureFlagOIDCAuth)).To(Equal(tt.oidcEnabledFlag))
		})
	}
}

func makeOIDCSecret(oidcConfig *mockoidc.Config, secretName string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
			Namespace: "flux-system",
		},
		Data: map[string][]byte{
			"issuerURL": []byte(oidcConfig.Issuer),
			"clientID": []byte(oidcConfig.ClientID),
			"clientSecret": []byte(oidcConfig.ClientSecret),
			"redirectURL": []byte("test.invalid/oauth2/callback"),
		},
	}
}

func makeClusterUserSecret(password, secretName string) *corev1.Secret {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name: secretName,
			Namespace: "flux-system",
		},
		Data: map[string][]byte{
			"password": hashed,
		},
	}
}
