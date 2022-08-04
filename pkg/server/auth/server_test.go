package auth_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/oauth2-proxy/mockoidc"
	"github.com/onsi/gomega"
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
	g := gomega.NewGomegaWithT(t)

	methods := []string{
		http.MethodPost,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodHead,
		http.MethodOptions,
	}

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC})

	for _, m := range methods {
		req := httptest.NewRequest(m, "https://example.com/callback", nil)
		w := httptest.NewRecorder()
		s.Callback().ServeHTTP(w, req)

		resp := w.Result()
		g.Expect(resp.StatusCode).To(Equal(http.StatusMethodNotAllowed))
		g.Expect(resp.Header.Get("Allow")).To(Equal(http.MethodGet))
	}
}

func TestCallbackErrorFromOIDC(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC})

	req := httptest.NewRequest(http.MethodGet, "https://example.com/callback?error=invalid_request&error_description=Unsupported%20response_type%20value", nil)
	w := httptest.NewRecorder()
	s.Callback().ServeHTTP(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
}

func TestCallbackCodeIsEmpty(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC})

	req := httptest.NewRequest(http.MethodGet, "https://example.com/callback", nil)
	w := httptest.NewRecorder()
	s.Callback().ServeHTTP(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
}

func TestCallbackStateCookieNotSet(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC})

	req := httptest.NewRequest(http.MethodGet, "https://example.com/callback?code=123", nil)
	w := httptest.NewRecorder()
	s.Callback().ServeHTTP(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
}

func TestCallbackStateCookieNotValid(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC})

	req := httptest.NewRequest(http.MethodGet, "https://example.com/callback?code=123&state=some_state", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.StateCookieName,
		Value: "some_different_state",
	})

	w := httptest.NewRecorder()
	s.Callback().ServeHTTP(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
}

func TestCallbackStateCookieNotBase64Encoded(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC})

	req := httptest.NewRequest(http.MethodGet, "https://example.com/callback?code=123&state=some_state", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.StateCookieName,
		Value: "some_state",
	})

	w := httptest.NewRecorder()
	s.Callback().ServeHTTP(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
}

func TestCallbackStateCookieNotJSONPayload(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC})

	encState := base64.StdEncoding.EncodeToString([]byte("some_state"))

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("https://example.com/callback?code=123&state=%s", encState), nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.StateCookieName,
		Value: encState,
	})

	w := httptest.NewRecorder()
	s.Callback().ServeHTTP(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
}

func TestCallbackCodeExchangeError(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC})

	state, _ := json.Marshal(auth.SessionState{
		Nonce:     "abcde",
		ReturnURL: "https://example.com",
	})
	encState := base64.StdEncoding.EncodeToString(state)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("https://example.com/callback?code=123&state=%s", encState), nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.StateCookieName,
		Value: encState,
	})

	w := httptest.NewRecorder()
	s.Callback().ServeHTTP(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusInternalServerError))
}

func TestSignInAllowsPOST(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

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

	s, _ := makeAuthServer(t, ctrlclientfake.NewClientBuilder().Build(), tokenSignerVerifier, []auth.AuthMethod{auth.OIDC})

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
	g := gomega.NewGomegaWithT(t)

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	if err != nil {
		t.Errorf("failed to create HMAC signer: %v", err)
	}

	s, _ := makeAuthServer(t, ctrlclientfake.NewClientBuilder().Build(), tokenSignerVerifier, []auth.AuthMethod{auth.OIDC})

	req := httptest.NewRequest(http.MethodPost, "https://example.com/signin", nil)
	w := httptest.NewRecorder()
	s.SignIn().ServeHTTP(w, req)

	resp := w.Result()
	g.Expect(resp.StatusCode).To(Equal(http.StatusBadRequest))

	b, err := ioutil.ReadAll(resp.Body)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(string(b)).To(ContainSubstring("Failed to read request body."))
}

func TestSignInNoSecret(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	s, _ := makeAuthServer(t, ctrlclientfake.NewClientBuilder().Build(), tokenSignerVerifier, []auth.AuthMethod{auth.OIDC})

	j, err := json.Marshal(auth.LoginRequest{})
	g.Expect(err).NotTo(HaveOccurred())

	reader := bytes.NewReader(j)

	req := httptest.NewRequest(http.MethodPost, "https://example.com/signin", reader)
	w := httptest.NewRecorder()
	s.SignIn().ServeHTTP(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
}

func TestSignInWrongUsernameReturnsUnauthorized(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

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

	s, _ := makeAuthServer(t, fakeKubernetesClient, tokenSignerVerifier, []auth.AuthMethod{auth.UserAccount})

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
	g := gomega.NewGomegaWithT(t)

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

	s, _ := makeAuthServer(t, fakeKubernetesClient, tokenSignerVerifier, []auth.AuthMethod{auth.OIDC})

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

func TestSingInCorrectPassword(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

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

	s, _ := makeAuthServer(t, fakeKubernetesClient, tokenSignerVerifier, []auth.AuthMethod{auth.OIDC})

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

	var cookie *http.Cookie

	for _, c := range resp.Cookies() {
		if c.Name == auth.IDTokenCookieName {
			cookie = c
			break
		}
	}

	g.Expect(cookie).ToNot(BeNil())
	_, err = tokenSignerVerifier.Verify(cookie.Value)
	g.Expect(err).NotTo(HaveOccurred())
}

func TestUserInfoAllowsGET(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	methods := []string{
		http.MethodPost,
		http.MethodPatch,
		http.MethodDelete,
		http.MethodConnect,
		http.MethodHead,
		http.MethodOptions,
	}

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC})

	for _, m := range methods {
		req := httptest.NewRequest(m, "https://example.com/userinfo", nil)
		w := httptest.NewRecorder()
		s.UserInfo().ServeHTTP(w, req)

		resp := w.Result()
		g.Expect(resp.StatusCode).To(Equal(http.StatusMethodNotAllowed))
		g.Expect(resp.Header.Get("Allow")).To(Equal(http.MethodGet))
	}
}

func TestUserInfoIDTokenCookieNotSet(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	s, _ := makeAuthServer(t, nil, nil, []auth.AuthMethod{auth.OIDC})

	req := httptest.NewRequest(http.MethodGet, "https://example.com/userinfo", nil)
	w := httptest.NewRecorder()
	s.UserInfo().ServeHTTP(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusBadRequest))
}

func TestUserInfoAdminFlow(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

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
	s, _ := makeAuthServer(t, fakeKubernetesClient, tokenSignerVerifier, []auth.AuthMethod{auth.UserAccount})

	signed, err := tokenSignerVerifier.Sign("wego-admin")
	g.Expect(err).NotTo(HaveOccurred())

	req := httptest.NewRequest(http.MethodGet, "https://example.com/userinfo", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.IDTokenCookieName,
		Value: signed,
	})

	w := httptest.NewRecorder()
	s.UserInfo().ServeHTTP(w, req)

	resp := w.Result()
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	var info auth.UserInfo

	g.Expect(json.NewDecoder(resp.Body).Decode(&info)).To(Succeed())
	g.Expect(info.Email).To(Equal("wego-admin"))
}

func TestUserInfoAdminFlow_differentUsername(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

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

	s, _ := makeAuthServer(t, fakeKubernetesClient, tokenSignerVerifier, []auth.AuthMethod{auth.UserAccount})

	signed, err := tokenSignerVerifier.Sign("dev")
	g.Expect(err).NotTo(HaveOccurred())

	req := httptest.NewRequest(http.MethodGet, "https://example.com/userinfo", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.IDTokenCookieName,
		Value: signed,
	})

	w := httptest.NewRecorder()
	s.UserInfo().ServeHTTP(w, req)

	resp := w.Result()
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	var info auth.UserInfo

	g.Expect(json.NewDecoder(resp.Body).Decode(&info)).To(Succeed())
	g.Expect(info.Email).To(Equal("dev"))
}

func TestUserInfoAdminFlowBadCookie(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

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

	s, _ := makeAuthServer(t, fakeKubernetesClient, tokenSignerVerifier, []auth.AuthMethod{auth.UserAccount})

	req := httptest.NewRequest(http.MethodGet, "https://example.com/userinfo", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.IDTokenCookieName,
		Value: "",
	})

	w := httptest.NewRecorder()
	s.UserInfo().ServeHTTP(w, req)

	resp := w.Result()
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	var info auth.UserInfo

	g.Expect(json.NewDecoder(resp.Body).Decode(&info)).To(Succeed())
	g.Expect(info.Email).To(Equal(""))
}

func TestUserInfoOIDCFlow(t *testing.T) {
	const (
		state = "abcdef"
		nonce = "ghijkl"
		code  = "mnopqr"
	)

	g := gomega.NewGomegaWithT(t)

	tokenSignerVerifier, err := auth.NewHMACTokenSignerVerifier(5 * time.Minute)
	g.Expect(err).NotTo(HaveOccurred())

	s, m := makeAuthServer(t, nil, tokenSignerVerifier, []auth.AuthMethod{auth.OIDC})

	authorizeQuery := url.Values{}
	authorizeQuery.Set("client_id", m.Config().ClientID)
	authorizeQuery.Set("scope", "openid email profile groups")
	authorizeQuery.Set("response_type", "code")
	authorizeQuery.Set("redirect_uri", "https://example.com/oauth2/callback")
	authorizeQuery.Set("state", state)
	authorizeQuery.Set("nonce", nonce)

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

	tokenForm := url.Values{}
	tokenForm.Set("client_id", m.Config().ClientID)
	tokenForm.Set("client_secret", m.Config().ClientSecret)
	tokenForm.Set("grant_type", "authorization_code")
	tokenForm.Set("code", code)

	tokenReq, err := http.NewRequest(
		http.MethodPost, m.TokenEndpoint(), bytes.NewBufferString(tokenForm.Encode()))
	g.Expect(err).NotTo(HaveOccurred())
	tokenReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	tokenResp, err := httpClient.Do(tokenReq)
	g.Expect(err).NotTo(HaveOccurred())
	g.Expect(tokenResp.StatusCode).To(Equal(http.StatusOK))

	defer tokenResp.Body.Close()
	body, err := ioutil.ReadAll(tokenResp.Body)
	g.Expect(err).NotTo(HaveOccurred())

	tokens := make(map[string]interface{})
	g.Expect(json.Unmarshal(body, &tokens)).To(Succeed())

	_, err = m.Keypair.VerifyJWT(tokens["access_token"].(string))
	g.Expect(err).NotTo(HaveOccurred())
	_, err = m.Keypair.VerifyJWT(tokens["refresh_token"].(string))
	g.Expect(err).NotTo(HaveOccurred())
	idToken, err := m.Keypair.VerifyJWT(tokens["id_token"].(string))
	g.Expect(err).NotTo(HaveOccurred())

	req := httptest.NewRequest(http.MethodGet, "https://example.com/userinfo", nil)
	req.AddCookie(&http.Cookie{
		Name:  auth.IDTokenCookieName,
		Value: idToken.Raw,
	})

	w := httptest.NewRecorder()
	s.UserInfo().ServeHTTP(w, req)

	resp := w.Result()
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	var info auth.UserInfo

	g.Expect(json.NewDecoder(resp.Body).Decode(&info)).To(Succeed())
	g.Expect(info.Email).To(Equal("jane.doe@example.com"))
}

func TestLogoutSuccess(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

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

	s, _ := makeAuthServer(t, fakeKubernetesClient, nil, []auth.AuthMethod{auth.UserAccount})

	w := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodPost, "https://example.com/logout", nil)
	s.Logout().ServeHTTP(w, req)

	resp := w.Result()
	g.Expect(resp.StatusCode).To(Equal(http.StatusOK))

	cookies := make(map[string]*http.Cookie)

	for _, c := range resp.Cookies() {
		if c.Name == auth.IDTokenCookieName || c.Name == auth.AccessTokenCookieName {
			cookies[c.Name] = c
		}
	}

	g.Expect(cookies).To(HaveKey(auth.IDTokenCookieName))
	g.Expect(cookies).To(HaveKey(auth.AccessTokenCookieName))

	for _, c := range cookies {
		g.Expect(c).ToNot(BeNil())
		g.Expect(c.Value).To(Equal(""))
	}
}

func TestLogoutWithWrongMethod(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

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

	s, _ := makeAuthServer(t, fakeKubernetesClient, nil, []auth.AuthMethod{auth.UserAccount})

	w := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "https://example.com/logout", nil)
	s.Logout().ServeHTTP(w, req)

	g.Expect(w.Result().StatusCode).To(Equal(http.StatusMethodNotAllowed))
}

func makeAuthServer(t *testing.T, client ctrlclient.Client, tsv auth.TokenSignerVerifier, authMethods []auth.AuthMethod) (*auth.AuthServer, *mockoidc.MockOIDC) {
	t.Helper()
	g := gomega.NewGomegaWithT(t)

	featureflags.Set("OIDC_AUTH", "") // Reset this

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

	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), oidcCfg, client, tsv, authMethodsMap)
	g.Expect(err).NotTo(HaveOccurred())

	s, err := auth.NewAuthServer(context.Background(), authCfg)
	g.Expect(err).NotTo(HaveOccurred())

	return s, m
}

func TestAuthMethods(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	featureflags.Set("OIDC_AUTH", "")
	featureflags.Set("CLUSTER_USER_AUTH", "")

	authMethods := map[auth.AuthMethod]bool{}

	authCfg, err := auth.NewAuthServerConfig(logr.Discard(), auth.OIDCConfig{}, ctrlclientfake.NewClientBuilder().Build(), nil, authMethods)
	g.Expect(err).NotTo(HaveOccurred())

	_, err = auth.NewAuthServer(context.Background(), authCfg)
	g.Expect(err).To(HaveOccurred())

	g.Expect(featureflags.Get("OIDC_AUTH")).To(Equal("false"))
	g.Expect(featureflags.Get("CLUSTER_USER_AUTH")).To(Equal("false"))

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

	authCfg, err = auth.NewAuthServerConfig(logr.Discard(), auth.OIDCConfig{}, fakeKubernetesClient, nil, authMethods)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = auth.NewAuthServer(context.Background(), authCfg)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(featureflags.Get("OIDC_AUTH")).To(Equal("false"))
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

	authCfg, err = auth.NewAuthServerConfig(logr.Discard(), oidcCfg, ctrlclientfake.NewClientBuilder().Build(), nil, authMethods)
	g.Expect(err).NotTo(HaveOccurred())
	_, err = auth.NewAuthServer(context.Background(), authCfg)
	g.Expect(err).NotTo(HaveOccurred())

	g.Expect(featureflags.Get("OIDC_AUTH")).To(Equal("true"))
	g.Expect(featureflags.Get("CLUSTER_USER_AUTH")).To(Equal("false"))
}
