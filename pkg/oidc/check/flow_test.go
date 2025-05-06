package check_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/weaveworks/weave-gitops/pkg/logger"
	"github.com/weaveworks/weave-gitops/pkg/oidc/check"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

type TestProvider struct {
	srv             *httptest.Server
	URL             string
	RequestedScopes []string
	GenClaims       func() jwt.Claims
}

func (tp TestProvider) genToken() string {
	claims := tp.GenClaims()
	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	ss, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		panic(err)
	}
	return ss
}

func (tp *TestProvider) Start() error {
	mux := &http.ServeMux{}

	mux.HandleFunc("/.well-known/openid-configuration", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		fmt.Fprintf(w, `{
"issuer": "%s",
"authorization_endpoint": "%s/auth",
"token_endpoint": "%s/token"
}`, tp.URL, tp.URL, tp.URL)
	})

	mux.HandleFunc("/auth", func(w http.ResponseWriter, r *http.Request) {
		tp.RequestedScopes = strings.Split(r.URL.Query().Get("scope"), " ")
		go http.Get(r.URL.Query().Get("redirect_uri"))
	})

	mux.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("content-type", "application/json")
		fmt.Fprintf(w, `{
"access_token": "token",
"id_token": "%s"
}`, tp.genToken())
	})

	tp.srv = httptest.NewServer(mux)
	tp.URL = tp.srv.URL

	return nil
}

func (tp TestProvider) Shutdown() {
	tp.srv.Close()
}

func (tp TestProvider) IssuerURL() string {
	return tp.URL
}

func TestGetClaimsWithSecret(t *testing.T) {
	var issuer string

	tests := []struct {
		name        string
		secret      func() *corev1.Secret
		expectedErr string
	}{
		{
			name: "empty secret",
			secret: func() *corev1.Secret {
				return &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "flux-system",
						Name:      "test-oidc",
					},
				}
			},
			expectedErr: "missing fields: clientID,clientSecret,issuerURL,redirectURL",
		},
		{
			name: "valid secret",
			secret: func() *corev1.Secret {
				return &corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Namespace: "flux-system",
						Name:      "test-oidc",
					},
					Data: map[string][]byte{
						"clientID":     []byte("client"),
						"clientSecret": []byte("csec"),
						"issuerURL":    []byte(issuer),
						"redirectURL":  []byte("something else"),
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			tp := TestProvider{
				GenClaims: func() jwt.Claims {
					return struct {
						jwt.RegisteredClaims
						Username string `json:"email"`
					}{
						RegisteredClaims: jwt.RegisteredClaims{
							Issuer:    issuer,
							Audience:  []string{"client"},
							ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
						},
						Username: "user@example.org",
					}
				},
			}
			g.Expect(tp.Start()).To(Succeed())

			t.Cleanup(func() {
				tp.Shutdown()
			})

			issuer = tp.URL
			c := fake.NewClientBuilder().
				WithObjects(tt.secret()).
				Build()

			var logBuf strings.Builder
			log := logger.NewCLILogger(&logBuf)

			_, err := check.GetPrincipal(t.Context(), check.Options{
				SecretName:      "test-oidc",
				SecretNamespace: "flux-system",
				OpenURL: func(u string) error {
					http.Get(u) // #nosec: G107
					return nil
				},
				InsecureSkipSignatureCheck: true,
			}, log, c)

			if tt.expectedErr != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tt.expectedErr)))
				return
			}

			g.Expect(err).NotTo(HaveOccurred())
		})
	}
}

func TestGetClaimsWithoutSecret(t *testing.T) {
	var issuer string

	tests := []struct {
		name             string
		opts             check.Options
		claims           func() jwt.Claims
		expectedScopes   []string
		expectedUsername string
		expectedGroups   []string
		expectedErr      string
	}{
		{
			name:           "requests default scopes",
			expectedScopes: auth.DefaultScopes,
			claims: func() jwt.Claims {
				return struct {
					jwt.RegisteredClaims
					Username string `json:"email"`
				}{
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    issuer,
						Audience:  []string{"client"},
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
					},
					Username: "user@example.org",
				}
			},
			expectedUsername: "user@example.org",
		},
		{
			name: "requests scopes from options",
			opts: check.Options{
				Scopes: []string{"foo", "bar"},
			},
			claims: func() jwt.Claims {
				return struct {
					jwt.RegisteredClaims
					Username string `json:"email"`
				}{
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    issuer,
						Audience:  []string{"client"},
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
					},
					Username: "user@example.org",
				}
			},
			expectedScopes: []string{"foo", "bar"},
		},
		{
			name: "respects username/groups claim name from options",
			opts: check.Options{
				ClaimUsername: "foo",
				ClaimGroups:   "bar",
			},
			claims: func() jwt.Claims {
				return struct {
					jwt.RegisteredClaims
					Username string   `json:"foo"`
					Groups   []string `json:"bar"`
				}{
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    issuer,
						Audience:  []string{"client"},
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
					},
					Username: "userfoo@example.org",
					Groups:   []string{"groups"},
				}
			},
			expectedUsername: "userfoo@example.org",
			expectedGroups:   []string{"groups"},
		},
		{
			name: "returns proper groups",
			claims: func() jwt.Claims {
				return struct {
					jwt.RegisteredClaims
					Username string   `json:"email"`
					Groups   []string `json:"groups"`
				}{
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    issuer,
						Audience:  []string{"client"},
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
					},
					Username: "user@example.org",
					Groups:   []string{"g1", "g2", "g3"},
				}
			},
			expectedGroups: []string{"g1", "g2", "g3"},
		},
		{
			name: "gracefully handles string groups claim (as opposed to list)",
			claims: func() jwt.Claims {
				return struct {
					jwt.RegisteredClaims
					Username string `json:"email"`
					Groups   string `json:"groups"`
				}{
					RegisteredClaims: jwt.RegisteredClaims{
						Issuer:    issuer,
						Audience:  []string{"client"},
						ExpiresAt: jwt.NewNumericDate(time.Now().Add(10 * time.Minute)),
					},
					Username: "user@example.org",
					Groups:   "g1",
				}
			},
			expectedGroups: []string{"g1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)

			tp := TestProvider{
				GenClaims: func() jwt.Claims {
					return tt.claims()
				},
			}
			g.Expect(tp.Start()).To(Succeed())

			t.Cleanup(func() {
				tp.Shutdown()
			})

			issuer = tp.URL

			// apply defaults

			if tt.opts.OpenURL == nil {
				tt.opts.OpenURL = func(u string) error {
					http.Get(u) // #nosec: G107
					return nil
				}
			}
			if tt.opts.IssuerURL == "" {
				tt.opts.IssuerURL = tp.IssuerURL()
			}
			tt.opts.InsecureSkipSignatureCheck = true
			if tt.opts.ClientID == "" {
				tt.opts.ClientID = "client"
			}

			var logBuf strings.Builder
			log := logger.NewCLILogger(&logBuf)

			c, err := check.GetPrincipal(t.Context(), tt.opts, log, nil)

			if tt.expectedErr != "" {
				g.Expect(err).To(MatchError(ContainSubstring(tt.expectedErr)))
				return
			}

			g.Expect(err).NotTo(HaveOccurred())

			if tt.expectedScopes != nil {
				g.Expect(tp.RequestedScopes).To(Equal(tt.expectedScopes))
			}
			if tt.expectedUsername != "" {
				g.Expect(c.ID).To(Equal(tt.expectedUsername))
			}
			if tt.expectedGroups != nil {
				g.Expect(c.Groups).To(ConsistOf(tt.expectedGroups))
			}
		})
	}
}
