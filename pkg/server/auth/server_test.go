package auth_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/go-logr/logr"
	"github.com/google/go-cmp/cmp"
	"github.com/oauth2-proxy/mockoidc"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/featureflags"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"golang.org/x/crypto/bcrypt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	ctrlclientfake "sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// A custom client that doesn't automatically follow redirects
var httpClient = &http.Client{
	CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	},
}

func TestCallbackAllowsGet(t *testing.T) {
	g := NewGomegaWithT(t)

	methods := []string{
		http.MethodPost,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodHead,
		http.MethodOptions,
	}

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC}, scs.New())

	for _, m := range methods {
		req := httptest.NewRequest(m, "https://example.com/callback", nil)
		w := httptest.NewRecorder()
		s.Callback(w, req)

		resp := w.Result()
		g.Expect(resp.StatusCode).To(Equal(http.StatusMethodNotAllowed))
		g.Expect(resp.Header.Get("Allow")).To(Equal(http.MethodGet))
	}
}

func TestCallbackErrorFromOIDC(t *testing.T) {
	g := NewGomegaWithT(t)
	sm := &fakeSessionManager{}

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC}, sm)

	req := httptest.NewRequest(http.MethodGet, "https://example.com/callback?error=invalid_request&error_description=Unsupported%20response_type%20value", nil).
		WithContext(contextWithValues(context.Background(), map[string]any{}))
	w := httptest.NewRecorder()
	s.Callback(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
}

func TestCallbackCodeIsEmpty(t *testing.T) {
	g := NewGomegaWithT(t)

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	req := httptest.NewRequest(http.MethodGet, "https://example.com/callback", nil)
	w := httptest.NewRecorder()
	s.Callback(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
}

func TestCallbackStateCookieNotSet(t *testing.T) {
	g := NewGomegaWithT(t)

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	req := httptest.NewRequest(http.MethodGet, "https://example.com/callback?code=123", nil)
	w := httptest.NewRecorder()
	s.Callback(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
}

func TestCallbackStateCookieNotValid(t *testing.T) {
	g := NewGomegaWithT(t)

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	req := httptest.NewRequest(http.MethodGet, "https://example.com/callback?code=123&state=some_state", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.StateCookieName,
		Value: "some_different_state",
	})

	w := httptest.NewRecorder()
	s.Callback(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
}

func TestCallbackStateCookieNotBase64Encoded(t *testing.T) {
	g := NewGomegaWithT(t)

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	req := httptest.NewRequest(http.MethodGet, "https://example.com/callback?code=123&state=some_state", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.StateCookieName,
		Value: "some_state",
	})

	w := httptest.NewRecorder()
	s.Callback(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
}

func TestCallbackStateCookieNotJSONPayload(t *testing.T) {
	g := NewGomegaWithT(t)

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	encState := base64.StdEncoding.EncodeToString([]byte("some_state"))

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("https://example.com/callback?code=123&state=%s", encState), nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.StateCookieName,
		Value: encState,
	})

	w := httptest.NewRecorder()
	s.Callback(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
}

func TestCallbackCodeExchangeError(t *testing.T) {
	g := NewGomegaWithT(t)

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	state, _ := json.Marshal(auth.SessionState{
		Nonce:     "abcde",
		ReturnURL: "https://example.com",
	})
	encState := base64.StdEncoding.EncodeToString(state)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("https://example.com/callback?code=123&state=%s", encState), nil).
		WithContext(contextWithValues(context.Background(), map[string]any{
			auth.StateCookieName: encState,
		}))

	w := httptest.NewRecorder()
	s.Callback(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusInternalServerError))
}

func TestSignInAllowsPOST(t *testing.T) {
	g := NewGomegaWithT(t)

	methods := []string{
		http.MethodGet,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodHead,
		http.MethodOptions,
	}

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	s, _ := makeAuthServer(t, ctrlclientfake.NewClientBuilder().Build(), tokenSignerVerifier, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	for _, m := range methods {
		req := httptest.NewRequest(m, "https://example.com/signin", nil)
		w := httptest.NewRecorder()
		s.SignIn().ServeHTTP(w, req)

		resp := w.Result()
		g.Expect(resp.StatusCode).To(Equal(http.StatusMethodNotAllowed))
		g.Expect(resp.Header.Get("Allow")).To(Equal(http.MethodPost))
	}
}

func TestSignInNoPayloadReturnsBadRequest(t *testing.T) {
	g := NewGomegaWithT(t)

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	if err != nil {
		t.Errorf("failed to create HMAC signer: %v", err)
	}

	s, _ := makeAuthServer(t, ctrlclientfake.NewClientBuilder().Build(), tokenSignerVerifier, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	req := httptest.NewRequest(http.MethodPost, "https://example.com/signin", nil)
	w := httptest.NewRecorder()
	s.SignIn().ServeHTTP(w, req)

	resp := w.Result()
	g.Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

	b, err := io.ReadAll(resp.Body)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(string(b)).To(ContainSubstring("Failed to read request body."))
}

func TestSignInNoSecret(t *testing.T) {
	g := NewGomegaWithT(t)

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	s, _ := makeAuthServer(t, ctrlclientfake.NewClientBuilder().Build(), tokenSignerVerifier, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	j, err := json.Marshal(auth.LoginRequest{})
	g.Expect(err).NotTo(HaveOccurred())

	reader := bytes.NewReader(j)

	req := httptest.NewRequest(http.MethodPost, "https://example.com/signin", reader)
	w := httptest.NewRecorder()
	s.SignIn().ServeHTTP(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
}

func TestSignInWrongUsernameReturnsUnauthorized(t *testing.T) {
	g := NewGomegaWithT(t)

	username := "admin"
	password := "my-secret-password"

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	g.Expect(err).NotTo(HaveOccurred())

	hashedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-user-auth",
			Namespace: "flux-system",
		},
		Data: map[string][]byte{
			"username": []byte(base64.StdEncoding.EncodeToString([]byte(username))),
			"password": hashed,
		},
	}

	fakeKubernetesClient := ctrlclientfake.NewClientBuilder().WithObjects(hashedSecret).Build()

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	s, _ := makeAuthServer(t, fakeKubernetesClient, tokenSignerVerifier, []auth.AuthMethod{auth.UserAccount}, &fakeSessionManager{})

	login := auth.LoginRequest{
		Username: "wrong",
		Password: "my-secret-password",
	}

	j, err := json.Marshal(login)
	g.Expect(err).NotTo(HaveOccurred())

	reader := bytes.NewReader(j)

	req := httptest.NewRequest(http.MethodPost, "https://example.com/signin", reader)
	w := httptest.NewRecorder()
	s.SignIn().ServeHTTP(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusUnauthorized))
}

func TestSignInWrongPasswordReturnsUnauthorized(t *testing.T) {
	g := NewGomegaWithT(t)

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

	fakeKubernetesClient := ctrlclientfake.NewClientBuilder().WithObjects(hashedSecret).Build()

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	s, _ := makeAuthServer(t, fakeKubernetesClient, tokenSignerVerifier, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	login := auth.LoginRequest{
		Password: "wrong",
	}

	j, err := json.Marshal(login)
	g.Expect(err).NotTo(HaveOccurred())

	reader := bytes.NewReader(j)

	req := httptest.NewRequest(http.MethodPost, "https://example.com/signin", reader)
	w := httptest.NewRecorder()
	s.SignIn().ServeHTTP(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusUnauthorized))
}

func TestSignInCorrectPassword(t *testing.T) {
	g := NewGomegaWithT(t)
	sm := &fakeSessionManager{}
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

	fakeKubernetesClient := ctrlclientfake.NewClientBuilder().WithObjects(hashedSecret).Build()

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	s, _ := makeAuthServer(t, fakeKubernetesClient, tokenSignerVerifier, []auth.AuthMethod{auth.OIDC}, sm)

	login := auth.LoginRequest{
		Password: password,
	}

	j, err := json.Marshal(login)
	g.Expect(err).NotTo(HaveOccurred())

	reader := bytes.NewReader(j)

	req := httptest.NewRequest(http.MethodPost, "https://example.com/signin", reader)
	w := httptest.NewRecorder()
	s.SignIn().ServeHTTP(w, req)

	resp := w.Result()
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	idToken := sm.stringValue(auth.IDTokenCookieName)
	_, err = tokenSignerVerifier.Verify(idToken)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestUserInfoAllowsGET(t *testing.T) {
	g := NewGomegaWithT(t)

	methods := []string{
		http.MethodPost,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodHead,
		http.MethodOptions,
	}

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	for _, m := range methods {
		req := httptest.NewRequest(m, "https://example.com/userinfo", nil)
		w := httptest.NewRecorder()
		s.UserInfo(w, req)

		resp := w.Result()
		g.Expect(resp.StatusCode).To(Equal(http.StatusMethodNotAllowed))
		g.Expect(resp.Header.Get("Allow")).To(Equal(http.MethodGet))
	}
}

func TestUserInfoIDTokenCookieNotSet(t *testing.T) {
	g := NewGomegaWithT(t)

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	req := httptest.NewRequest(http.MethodGet, "https://example.com/userinfo", nil)
	w := httptest.NewRecorder()
	s.UserInfo(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
}

func TestUserInfoAnonymous(t *testing.T) {
	g := NewGomegaWithT(t)

	authMethods := map[auth.AuthMethod]bool{
		auth.Anonymous: true,
	}
	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), auth.OIDCConfig{}, nil, nil, testNamespace, authMethods, "test-user", nil)
	g.Expect(err).NotTo(HaveOccurred())

	s, err := auth.NewAuthServer(context.Background(), authCfg)
	g.Expect(err).NotTo(HaveOccurred())

	req := httptest.NewRequest(http.MethodGet, "/foo", nil)
	w := httptest.NewRecorder()
	s.UserInfo(w, req)

	// Anonymous auth should always return a 200
	resp := w.Result()
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	// Check we return the noAuthUser
	var info auth.UserInfo
	g.Expect(json.NewDecoder(resp.Body).Decode(&info)).To(Succeed())
	g.Expect(info.ID).To(Equal("test-user"))
}

func TestUserInfoAdminFlow(t *testing.T) {
	g := NewGomegaWithT(t)
	sm := &fakeSessionManager{}

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	hashedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-user-auth",
			Namespace: "flux-system",
		},
		Data: map[string][]byte{
			"username": []byte("anything"),
			"password": []byte("hash"),
		},
	}
	fakeKubernetesClient := ctrlclientfake.NewClientBuilder().WithObjects(hashedSecret).Build()
	s, _ := makeAuthServer(t, fakeKubernetesClient, tokenSignerVerifier, []auth.AuthMethod{auth.UserAccount}, sm)

	signed, err := tokenSignerVerifier.Sign("wego-admin")
	g.Expect(err).NotTo(HaveOccurred())

	req := httptest.NewRequest(http.MethodGet, "https://example.com/userinfo", nil).WithContext(
		contextWithSessionValues(map[string]any{
			auth.IDTokenCookieName: signed,
		}))

	w := httptest.NewRecorder()
	s.UserInfo(w, req)

	resp := w.Result()
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	var info auth.UserInfo

	g.Expect(json.NewDecoder(resp.Body).Decode(&info)).To(Succeed())
	g.Expect(info.Email).To(Equal("wego-admin"))
}

func TestUserInfoAdminFlow_differentUsername(t *testing.T) {
	g := NewGomegaWithT(t)

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	hashedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-user-auth",
			Namespace: "flux-system",
		},
		Data: map[string][]byte{
			"username": []byte("anything"),
			"password": []byte("hash"),
		},
	}
	fakeKubernetesClient := ctrlclientfake.NewClientBuilder().WithObjects(hashedSecret).Build()

	s, _ := makeAuthServer(t, fakeKubernetesClient, tokenSignerVerifier, []auth.AuthMethod{auth.UserAccount}, &fakeSessionManager{})

	signed, err := tokenSignerVerifier.Sign("dev")
	g.Expect(err).NotTo(HaveOccurred())

	req := httptest.NewRequest(http.MethodGet, "https://example.com/userinfo", nil).WithContext(
		contextWithSessionValues(map[string]any{
			auth.IDTokenCookieName: signed,
		}))

	w := httptest.NewRecorder()
	s.UserInfo(w, req)

	resp := w.Result()
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	var info auth.UserInfo

	g.Expect(json.NewDecoder(resp.Body).Decode(&info)).To(Succeed())
	g.Expect(info.Email).To(Equal("dev"))
}

func TestUserInfoAdminFlowBadCookie(t *testing.T) {
	g := NewGomegaWithT(t)

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	hashedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-user-auth",
			Namespace: "flux-system",
		},
		Data: map[string][]byte{
			"username": []byte("anything"),
			"password": []byte("hash"),
		},
	}
	fakeKubernetesClient := ctrlclientfake.NewClientBuilder().WithObjects(hashedSecret).Build()

	s, _ := makeAuthServer(t, fakeKubernetesClient, tokenSignerVerifier, []auth.AuthMethod{auth.UserAccount}, &fakeSessionManager{})

	req := httptest.NewRequest(http.MethodGet, "https://example.com/userinfo", nil).WithContext(
		contextWithSessionValues(map[string]any{
			auth.IDTokenCookieName: "bad-cookie",
		}))

	w := httptest.NewRecorder()
	s.UserInfo(w, req)

	resp := w.Result()
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	var info auth.UserInfo

	g.Expect(json.NewDecoder(resp.Body).Decode(&info)).To(Succeed())
	g.Expect(info.Email).To(Equal(""))
}

func getVerifyTokens(t *testing.T, m *mockoidc.MockOIDC) map[string]interface{} {
	const (
		state = "abcdef"
		nonce = "ghijkl"
		code  = "mnopqr"
	)

	g := NewGomegaWithT(t)

	authorizeQuery := valuesFromMap(map[string]string{
		"client_id":     m.Config().ClientID,
		"scope":         "openid email profile groups",
		"response_type": "code",
		"redirect_uri":  "https://example.com/oauth2/callback",
		"state":         state,
		"nonce":         nonce,
	})

	authorizeURL, err := url.Parse(m.AuthorizationEndpoint())
	g.Expect(err).NotTo(HaveOccurred())

	authorizeURL.RawQuery = authorizeQuery.Encode()

	authorizeReq, err := http.NewRequest(http.MethodGet, authorizeURL.String(), nil)
	g.Expect(err).NotTo(HaveOccurred())

	m.QueueCode(code)

	authorizeResp, err := httpClient.Do(authorizeReq)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(authorizeResp.StatusCode).To(Equal(http.StatusFound))

	appRedirect, err := url.Parse(authorizeResp.Header.Get("Location"))
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(appRedirect.Query().Get("code")).To(Equal(code))
	g.Expect(appRedirect.Query().Get("state")).To(Equal(state))

	tokenForm := valuesFromMap(map[string]string{
		"client_id":     m.Config().ClientID,
		"client_secret": m.Config().ClientSecret,
		"grant_type":    "authorization_code",
		"code":          code,
	})

	tokenReq, err := http.NewRequest(
		http.MethodPost, m.TokenEndpoint(), bytes.NewBufferString(tokenForm.Encode()))
	g.Expect(err).NotTo(HaveOccurred())
	tokenReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	tokenResp, err := httpClient.Do(tokenReq)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(tokenResp.StatusCode).To(Equal(http.StatusOK))

	defer tokenResp.Body.Close()
	body, err := io.ReadAll(tokenResp.Body)
	g.Expect(err).NotTo(HaveOccurred())

	tokens := make(map[string]interface{})
	g.Expect(json.Unmarshal(body, &tokens)).To(Succeed())

	return tokens
}

func TestUserInfoOIDCFlow(t *testing.T) {
	g := NewGomegaWithT(t)
	sm := &fakeSessionManager{}

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	s, m := makeAuthServer(t, nil, tokenSignerVerifier, []auth.AuthMethod{auth.OIDC}, sm)

	tokens := getVerifyTokens(t, m)

	_, err = m.Keypair.VerifyJWT(tokens["access_token"].(string))
	g.Expect(err).NotTo(HaveOccurred())
	_, err = m.Keypair.VerifyJWT(tokens["refresh_token"].(string))
	g.Expect(err).NotTo(HaveOccurred())
	idToken, err := m.Keypair.VerifyJWT(tokens["id_token"].(string))
	g.Expect(err).NotTo(HaveOccurred())

	req := httptest.NewRequest(http.MethodGet, "https://example.com/userinfo", nil).WithContext(
		contextWithSessionValues(map[string]any{
			auth.IDTokenCookieName: idToken.Raw,
		}))

	w := httptest.NewRecorder()
	s.UserInfo(w, req)

	resp := w.Result()
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	var info auth.UserInfo

	g.Expect(json.NewDecoder(resp.Body).Decode(&info)).To(Succeed())
	g.Expect(info.Email).To(Equal("jane.doe@example.com"))
}

func TestUserInfoOIDCFlow_with_custom_claims(t *testing.T) {
	const (
		state = "abcdef"
		nonce = "ghijkl"
		code  = "mnopqr"
	)

	g := NewGomegaWithT(t)

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	authServer, m := makeAuthServer(t, nil, tokenSignerVerifier, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	authorizeQuery := valuesFromMap(map[string]string{
		"client_id":     m.Config().ClientID,
		"scope":         "openid email profile groups",
		"response_type": "code",
		"redirect_uri":  "https://example.com/oauth2/callback",
		"state":         state,
		"nonce":         nonce,
	})

	authorizeURL, err := url.Parse(m.AuthorizationEndpoint())
	g.Expect(err).NotTo(HaveOccurred())

	authorizeURL.RawQuery = authorizeQuery.Encode()
	authorizeReq, err := http.NewRequest(http.MethodGet, authorizeURL.String(), nil)
	g.Expect(err).NotTo(HaveOccurred())

	m.QueueCode(code)

	authorizeResp, err := httpClient.Do(authorizeReq)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(authorizeResp.StatusCode).To(Equal(http.StatusFound))

	tokenForm := valuesFromMap(map[string]string{
		"client_id":     m.Config().ClientID,
		"client_secret": m.Config().ClientSecret,
		"grant_type":    "authorization_code",
		"code":          code,
	})

	tokenReq, err := http.NewRequest(
		http.MethodPost, m.TokenEndpoint(), bytes.NewBufferString(tokenForm.Encode()))
	g.Expect(err).NotTo(HaveOccurred())
	tokenReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	tokenResp, err := httpClient.Do(tokenReq)
	g.Expect(err).NotTo(HaveOccurred())

	defer tokenResp.Body.Close()

	body, err := io.ReadAll(tokenResp.Body)
	g.Expect(err).NotTo(HaveOccurred())

	tokens := make(map[string]interface{})
	g.Expect(json.Unmarshal(body, &tokens)).To(Succeed())

	idToken, err := m.Keypair.VerifyJWT(tokens["id_token"].(string))
	g.Expect(err).NotTo(HaveOccurred())

	req := httptest.NewRequest(http.MethodGet, "https://example.com/userinfo", nil).WithContext(
		contextWithSessionValues(map[string]any{
			auth.IDTokenCookieName: idToken.Raw,
		}))

	w := httptest.NewRecorder()
	authServer.OIDCConfig.ClaimsConfig = &auth.ClaimsConfig{
		Username: "preferred_username",
		Groups:   "groups",
	}

	authServer.UserInfo(w, req)

	resp := w.Result()
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	var info auth.UserInfo

	g.Expect(json.NewDecoder(resp.Body).Decode(&info)).To(Succeed())
	g.Expect(info.Email).To(Equal("jane.doe"))
	g.Expect(info.ID).To(Equal("jane.doe"))
}

// Given the user only has a valid refresh_token
// we should be able to refresh it and get an id_token and an access_token
func TestRefresh(t *testing.T) {
	g := NewGomegaWithT(t)
	sm := &fakeSessionManager{}

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	s, m := makeAuthServer(t, nil, tokenSignerVerifier, []auth.AuthMethod{auth.OIDC}, sm)

	tokens := getVerifyTokens(t, m)

	tf := tokens["refresh_token"].(string)

	req := httptest.NewRequest(http.MethodGet, "https://example.com/test", nil).WithContext(
		contextWithSessionValues(map[string]any{
			auth.RefreshTokenCookieName: tf,
		}))

	w := httptest.NewRecorder()

	user, err := s.Refresh(w, req)
	g.Expect(err).To(Succeed())
	g.Expect(user.ID).To(Equal("jane.doe@example.com"))

	// We should have the 3 cookie set.
	// Technically the system doesn't have to set the refresh_token again
	g.Expect(sm.PutValues).To(HaveKey(auth.IDTokenCookieName))
	g.Expect(sm.PutValues).To(HaveKey(auth.AccessTokenCookieName))
	g.Expect(sm.PutValues).To(HaveKey(auth.RefreshTokenCookieName))

	// And they should all be valid!
	_, err = m.Keypair.VerifyJWT(sm.stringValue(auth.IDTokenCookieName))
	g.Expect(err).NotTo(HaveOccurred())
	_, err = m.Keypair.VerifyJWT(sm.stringValue(auth.AccessTokenCookieName))
	g.Expect(err).NotTo(HaveOccurred())
	_, err = m.Keypair.VerifyJWT(sm.stringValue(auth.RefreshTokenCookieName))
	g.Expect(err).NotTo(HaveOccurred())
}

func TestRefreshSucceedsReturns200(t *testing.T) {
	g := NewGomegaWithT(t)

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	s, m := makeAuthServer(t, nil, tokenSignerVerifier, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	tokens := getVerifyTokens(t, m)

	tf := tokens["refresh_token"].(string)

	req := httptest.NewRequest(http.MethodPost, "https://example.com/test", nil).
		WithContext(
			contextWithSessionValues(map[string]any{
				auth.RefreshTokenCookieName: tf,
			}))

	w := httptest.NewRecorder()
	s.RefreshHandler(w, req)
	g.Expect(w.Result().StatusCode).To(Equal(200))
}

func TestRefreshHandlerRejectsMethodGet(t *testing.T) {
	g := NewGomegaWithT(t)
	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "https://example.com/test", nil)

	s.RefreshHandler(w, req)
	g.Expect(w.Result().StatusCode).To(Equal(http.StatusMethodNotAllowed))
}

func TestRefreshFailsReturns401(t *testing.T) {
	g := NewGomegaWithT(t)

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "https://example.com/test", nil)
	s.RefreshHandler(w, req)
	g.Expect(w.Result().StatusCode).To(Equal(http.StatusUnauthorized))
}

func TestRefreshNoToken(t *testing.T) {
	g := NewGomegaWithT(t)
	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "https://example.com/test", nil)
	user, err := s.Refresh(w, req)
	g.Expect(err).To(MatchError("couldn't fetch refresh token from cookie"))
	g.Expect(user).To(BeNil())
}

func TestRefreshInvalidToken(t *testing.T) {
	g := NewGomegaWithT(t)
	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC}, &fakeSessionManager{})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "https://example.com/test", nil).WithContext(
		contextWithSessionValues(map[string]any{
			auth.RefreshTokenCookieName: "invalid",
		}))

	user, err := s.Refresh(w, req)

	g.Expect(err).To(MatchError(MatchRegexp("failed to refresh token: oauth2: \"invalid_request\"")))
	g.Expect(user).To(BeNil())
}

func TestLogoutSuccess(t *testing.T) {
	g := NewGomegaWithT(t)

	hashedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-user-auth",
			Namespace: "flux-system",
		},
		Data: map[string][]byte{
			"username": []byte("anything"),
			"password": []byte("hash"),
		},
	}
	fakeKubernetesClient := ctrlclientfake.NewClientBuilder().WithObjects(hashedSecret).Build()
	sm := &fakeSessionManager{}
	s, _ := makeAuthServer(t, fakeKubernetesClient, nil, []auth.AuthMethod{auth.UserAccount}, sm)

	w := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodPost, "https://example.com/logout", nil).WithContext(
		contextWithValues(context.Background(), map[string]any{
			"sessionid": "test-session",
		}))

	s.Logout(w, req)

	resp := w.Result()
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	g.Expect(sm.Destroyed).To(ContainElement("test-session"))
}

func TestLogoutWithWrongMethod(t *testing.T) {
	g := NewGomegaWithT(t)

	hashedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-user-auth",
			Namespace: "flux-system",
		},
		Data: map[string][]byte{
			"username": []byte("anything"),
			"password": []byte("hash"),
		},
	}
	fakeKubernetesClient := ctrlclientfake.NewClientBuilder().WithObjects(hashedSecret).Build()

	s, _ := makeAuthServer(t, fakeKubernetesClient, nil, []auth.AuthMethod{auth.UserAccount}, &fakeSessionManager{})

	w := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "https://example.com/logout", nil)
	s.Logout(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusMethodNotAllowed))
}

func makeAuthServer(t *testing.T, client ctrlclient.Client, tsv auth.TokenSignerVerifier, authMethods []auth.AuthMethod, sm auth.SessionManager) (*auth.AuthServer, *mockoidc.MockOIDC) {
	t.Helper()
	g := NewGomegaWithT(t)

	featureflags.SetBoolean("OIDC_AUTH", false) // Reset this

	m, err := mockoidc.Run()
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		_ = m.Shutdown()
	})

	cfg := m.Config()

	if client == nil {
		client = ctrlclientfake.NewClientBuilder().Build()
	}

	oidcCfg := auth.OIDCConfig{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		IssuerURL:    cfg.Issuer,
	}

	authMethodsMap := map[auth.AuthMethod]bool{}
	for _, mthd := range authMethods {
		authMethodsMap[mthd] = true
	}

	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), oidcCfg, client, tsv, testNamespace, authMethodsMap, "", sm)
	g.Expect(err).NotTo(HaveOccurred())

	s, err := auth.NewAuthServer(context.Background(), authCfg)
	g.Expect(err).NotTo(HaveOccurred())

	return s, m
}

func TestAuthMethods(t *testing.T) {
	g := NewGomegaWithT(t)

	featureflags.SetBoolean("OIDC_AUTH", false)
	featureflags.SetBoolean("CLUSTER_USER_AUTH", false)
	featureflags.SetBoolean("ANONYMOUS_AUTH", false)

	authMethods := map[auth.AuthMethod]bool{}

	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), auth.OIDCConfig{}, ctrlclientfake.NewClientBuilder().Build(), nil, testNamespace, authMethods, "", scs.New())
	g.Expect(err).NotTo(HaveOccurred())

	_, err = auth.NewAuthServer(context.Background(), authCfg)
	g.Expect(err).To(MatchError(MatchRegexp("OIDC auth, local auth or anonymous mode must be enabled")))

	g.Expect(featureflags.Get("OIDC_AUTH")).To(Equal(""))
	g.Expect(featureflags.Get("CLUSTER_USER_AUTH")).To(Equal(""))

	hashedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster-user-auth",
			Namespace: "flux-system",
		},
		Data: map[string][]byte{
			"username": []byte("anything"),
			"password": []byte("hash"),
		},
	}
	fakeKubernetesClient := ctrlclientfake.NewClientBuilder().WithObjects(hashedSecret).Build()

	authMethods = map[auth.AuthMethod]bool{auth.UserAccount: true}

	authCfg, err = auth.NewAuthServerConfig(logr.Discard(), auth.OIDCConfig{}, fakeKubernetesClient, nil, testNamespace, authMethods, "", scs.New())
	g.Expect(err).NotTo(HaveOccurred())
	_, err = auth.NewAuthServer(context.Background(), authCfg)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(featureflags.Get("OIDC_AUTH")).To(Equal(""))
	g.Expect(featureflags.Get("CLUSTER_USER_AUTH")).To(Equal("true"))

	m, err := mockoidc.Run()
	g.Expect(err).NotTo(HaveOccurred())

	t.Cleanup(func() {
		_ = m.Shutdown()
	})

	cfg := m.Config()
	oidcCfg := auth.OIDCConfig{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		IssuerURL:    cfg.Issuer,
	}
	authMethods = map[auth.AuthMethod]bool{auth.OIDC: true}

	authCfg, err = auth.NewAuthServerConfig(logr.Discard(), oidcCfg, ctrlclientfake.NewClientBuilder().Build(), nil, testNamespace, authMethods, "", scs.New())
	g.Expect(err).NotTo(HaveOccurred())
	_, err = auth.NewAuthServer(context.Background(), authCfg)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(featureflags.Get("OIDC_AUTH")).To(Equal("true"))
	g.Expect(featureflags.Get("CLUSTER_USER_AUTH")).To(Equal(""))
}

func TestNewOIDCConfigFromSecret(t *testing.T) {
	configTests := []struct {
		name string
		data map[string][]byte
		want auth.OIDCConfig
	}{
		{
			name: "basic fields",
			data: map[string][]byte{
				"issuerURL":     []byte("https://example.com/test"),
				"clientID":      []byte("test-client-id"),
				"clientSecret":  []byte("test-client-secret"),
				"redirectURL":   []byte("https://example.com/redirect"),
				"tokenDuration": []byte("10m"),
			},
			want: auth.OIDCConfig{
				IssuerURL:     "https://example.com/test",
				ClientID:      "test-client-id",
				ClientSecret:  "test-client-secret",
				RedirectURL:   "https://example.com/redirect",
				TokenDuration: time.Minute * 10,
				ClaimsConfig:  &auth.ClaimsConfig{Username: "email", Groups: "groups"},
				Scopes:        []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, auth.ScopeEmail, auth.ScopeGroups},
			},
		},
		{
			name: "bad duration defaults to 1 hour",
			data: map[string][]byte{
				"tokenDuration": []byte("10x"),
			},
			want: auth.OIDCConfig{
				TokenDuration: time.Hour * 1,
				ClaimsConfig:  &auth.ClaimsConfig{Username: "email", Groups: "groups"},
				Scopes:        []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, auth.ScopeEmail, auth.ScopeGroups},
			},
		},
		{
			name: "overridden claims",
			data: map[string][]byte{
				"claimUsername": []byte("test-user"),
				"claimGroups":   []byte("test-groups"),
			},
			want: auth.OIDCConfig{
				TokenDuration: time.Hour * 1,
				Scopes:        []string{oidc.ScopeOpenID, oidc.ScopeOfflineAccess, auth.ScopeEmail, auth.ScopeGroups},
				ClaimsConfig: &auth.ClaimsConfig{
					Username: "test-user", Groups: "test-groups",
				},
			},
		},
		{
			name: "overridden scopes",
			data: map[string][]byte{
				"claimUsername": []byte("test-user"),
				"customScopes":  []byte("other-groups,new-user-id"),
			},
			want: auth.OIDCConfig{
				TokenDuration: time.Hour * 1,
				Scopes:        []string{"other-groups", "new-user-id"},
				ClaimsConfig: &auth.ClaimsConfig{
					Username: "test-user", Groups: "groups",
				},
			},
		},
	}

	for _, tt := range configTests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := auth.NewOIDCConfigFromSecret(corev1.Secret{Data: tt.data})

			if diff := cmp.Diff(tt.want, cfg); diff != "" {
				t.Fatalf("failed to parse config from secret:\n%s", diff)
			}
		})
	}
}

func valuesFromMap(data map[string]string) url.Values {
	vals := url.Values{}
	for k, v := range data {
		vals.Set(k, v)
	}

	return vals
}
