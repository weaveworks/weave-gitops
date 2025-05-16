package auth_test

import (
	"testing"

	"github.com/alexedwards/scs/v2"
	"github.com/go-logr/logr"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/onsi/gomega"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/weaveworks/weave-gitops/pkg/featureflags"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func TestInitAuthServer(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	m, err := mockoidc.Run()
	g.Expect(err).NotTo(gomega.HaveOccurred())

	t.Cleanup(func() {
		_ = m.Shutdown()
	})

	initTests := []struct {
		name            string
		authMethods     []string
		secrets         []*corev1.Secret
		cliOIDCConfig   auth.OIDCConfig
		oidcSecretName  string
		expectErr       bool
		clusterUserFlag string
		oidcEnabledFlag string
	}{
		{
			name:        "basic test",
			authMethods: []string{"user-account", "oidc"},
			secrets: []*corev1.Secret{
				makeOIDCSecret(m.Config(), auth.DefaultOIDCAuthSecretName),
				makeClusterUserSecret("my-secret-password", auth.ClusterUserAuthSecretName),
			},
			cliOIDCConfig:   auth.OIDCConfig{},
			oidcSecretName:  auth.DefaultOIDCAuthSecretName,
			expectErr:       false,
			clusterUserFlag: "true",
			oidcEnabledFlag: "true",
		},
		{
			name:        "OIDC Only",
			authMethods: []string{"oidc"},
			secrets: []*corev1.Secret{
				makeOIDCSecret(m.Config(), auth.DefaultOIDCAuthSecretName),
			},
			cliOIDCConfig:   auth.OIDCConfig{},
			oidcSecretName:  auth.DefaultOIDCAuthSecretName,
			expectErr:       false,
			clusterUserFlag: "",
			oidcEnabledFlag: "true",
		},
		{
			name:        "OIDC alt-secret",
			authMethods: []string{"oidc"},
			secrets: []*corev1.Secret{
				makeOIDCSecret(m.Config(), "alternate-oidc-secret"),
			},
			cliOIDCConfig:   auth.OIDCConfig{},
			oidcSecretName:  "alternate-oidc-secret",
			expectErr:       false,
			clusterUserFlag: "",
			oidcEnabledFlag: "true",
		},
		{
			name:        "OIDC via CLI",
			authMethods: []string{"oidc"},
			secrets:     []*corev1.Secret{},
			cliOIDCConfig: auth.OIDCConfig{
				IssuerURL:    m.Config().Issuer,
				ClientID:     m.Config().ClientID,
				ClientSecret: m.Config().ClientSecret,
				RedirectURL:  "example.invalid/oauth2/callback",
			},
			oidcSecretName:  auth.DefaultOIDCAuthSecretName,
			expectErr:       false,
			clusterUserFlag: "",
			oidcEnabledFlag: "true",
		},
		{
			name:        "User only",
			authMethods: []string{"user-account"},
			secrets: []*corev1.Secret{
				makeClusterUserSecret("my-secret-password", auth.ClusterUserAuthSecretName),
			},
			cliOIDCConfig:   auth.OIDCConfig{},
			oidcSecretName:  auth.DefaultOIDCAuthSecretName,
			expectErr:       false,
			clusterUserFlag: "true",
			oidcEnabledFlag: "",
		},
		{
			name:            "No auth methods",
			authMethods:     []string{},
			secrets:         []*corev1.Secret{},
			cliOIDCConfig:   auth.OIDCConfig{},
			oidcSecretName:  "",
			expectErr:       true,
			clusterUserFlag: "",
			oidcEnabledFlag: "",
		},
	}

	for _, tt := range initTests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset feature flags for each run
			featureflags.SetBoolean(auth.FeatureFlagClusterUser, false)
			featureflags.SetBoolean(auth.FeatureFlagOIDCAuth, false)
			featureflags.SetBoolean(auth.FeatureFlagAnonymousAuth, false)

			partialKubernetesClient := ctrlclient.NewClientBuilder()

			// This is because I can't (be bothered to) figure out how to []*secret -> []*client.Object
			for _, obj := range tt.secrets {
				partialKubernetesClient.WithObjects(obj)
			}

			fakeKubernetesClient := partialKubernetesClient.Build()

			srv, err := auth.InitAuthServer(t.Context(), logr.Discard(), fakeKubernetesClient, auth.AuthParams{
				AuthMethodStrings: tt.authMethods,
				OIDCConfig:        tt.cliOIDCConfig,
				Namespace:         "test-namespace",
				OIDCSecretName:    tt.oidcSecretName,
				SessionManager:    scs.New(),
			})

			if tt.expectErr {
				g.Expect(err).To(gomega.HaveOccurred())
				g.Expect(srv).To(gomega.BeNil())
			} else {
				g.Expect(err).NotTo(gomega.HaveOccurred())
				g.Expect(srv).NotTo(gomega.BeNil())
			}

			// Can check (somewhat) what's happened by making sure flags are set
			g.Expect(featureflags.Get(auth.FeatureFlagClusterUser)).To(gomega.Equal(tt.clusterUserFlag))
			g.Expect(featureflags.Get(auth.FeatureFlagOIDCAuth)).To(gomega.Equal(tt.oidcEnabledFlag))
		})
	}
}

func makeOIDCSecret(oidcConfig *mockoidc.Config, secretName string) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: "test-namespace",
		},
		Data: map[string][]byte{
			"issuerURL":    []byte(oidcConfig.Issuer),
			"clientID":     []byte(oidcConfig.ClientID),
			"clientSecret": []byte(oidcConfig.ClientSecret),
			"redirectURL":  []byte("test.invalid/oauth2/callback"),
		},
	}
}

func makeClusterUserSecret(password, secretName string) *corev1.Secret {
	hashed, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: "test-namespace",
		},
		Data: map[string][]byte{
			"password": hashed,
		},
	}
}
