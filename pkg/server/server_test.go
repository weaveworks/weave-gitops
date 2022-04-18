package server_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/go-logr/logr"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/stretchr/testify/assert"
	pb "github.com/weaveworks/weave-gitops/pkg/api/applications"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/server"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/authfakes"
	authtypes "github.com/weaveworks/weave-gitops/pkg/services/auth/types"
	"github.com/weaveworks/weave-gitops/pkg/services/servicesfakes"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"github.com/weaveworks/weave-gitops/pkg/vendorfakes/fakelogr"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

var _ = Describe("ApplicationsServer", func() {
	var (
		namespace *corev1.Namespace
		err       error
		ctx       context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		namespace = &corev1.Namespace{}
		namespace.Name = "kube-test-" + rand.String(5)
		err = k8sClient.Create(ctx, namespace)
		Expect(err).NotTo(HaveOccurred(), "failed to create test namespace")
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, namespace)).To(Succeed())
	})

	It("Authorize", func() {
		ctx := context.Background()
		provider := "github"
		token := "token"

		jwtClient := auth.NewJwtClient(secretKey)
		expectedToken, err := jwtClient.GenerateJWT(auth.ExpirationTime, gitproviders.GitProviderGitHub, token)
		Expect(err).NotTo(HaveOccurred())

		res, err := appsClient.Authenticate(ctx, &pb.AuthenticateRequest{
			ProviderName: provider,
			AccessToken:  token,
		})
		Expect(err).NotTo(HaveOccurred())

		Expect(res.Token).To(Equal(expectedToken))
	})
	It("Authorize fails on wrong provider", func() {
		ctx := context.Background()
		provider := "wrong_provider"
		token := "token"

		_, err := appsClient.Authenticate(ctx, &pb.AuthenticateRequest{
			ProviderName: provider,
			AccessToken:  token,
		})

		Expect(err.Error()).To(ContainSubstring(server.ErrBadProvider.Error()))
		Expect(err.Error()).To(ContainSubstring(codes.InvalidArgument.String()))

	})
	It("Authorize fails on empty provider token", func() {
		ctx := context.Background()
		provider := "github"

		_, err := appsClient.Authenticate(ctx, &pb.AuthenticateRequest{
			ProviderName: provider,
			AccessToken:  "",
		})

		Expect(err).Should(testutils.MatchGRPCError(codes.InvalidArgument, server.ErrEmptyAccessToken))
	})
	Describe("GetGithubDeviceCode", func() {
		It("returns a device code", func() {
			ctx := context.Background()
			code := "123-456"
			ghAuthClient.GetDeviceCodeStub = func() (*auth.GithubDeviceCodeResponse, error) {
				return &auth.GithubDeviceCodeResponse{DeviceCode: code}, nil
			}

			res, err := appsClient.GetGithubDeviceCode(ctx, &pb.GetGithubDeviceCodeRequest{})
			Expect(err).NotTo(HaveOccurred())

			Expect(res.DeviceCode).To(Equal(code))
		})
		It("returns an error when github returns an error", func() {
			ctx := context.Background()
			someError := errors.New("some gh error")
			ghAuthClient.GetDeviceCodeStub = func() (*auth.GithubDeviceCodeResponse, error) {
				return nil, someError
			}
			_, err := appsClient.GetGithubDeviceCode(ctx, &pb.GetGithubDeviceCodeRequest{})
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue(), "could not get grpc status from err")
			Expect(st.Message()).To(ContainSubstring(someError.Error()))
		})
	})

	Describe("GetGithubAuthStatus", func() {
		It("returns an ErrAuthPending when the user is not yet authenticated", func() {
			ctx := context.Background()
			ghAuthClient.GetDeviceCodeAuthStatusStub = func(s string) (string, error) {
				return "", auth.ErrAuthPending
			}
			res, err := appsClient.GetGithubAuthStatus(ctx, &pb.GetGithubAuthStatusRequest{DeviceCode: "somedevicecode"})
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue(), "could not get status from err")
			Expect(st.Message()).To(ContainSubstring(auth.ErrAuthPending.Error()))
			Expect(res).To(BeNil())
		})
		It("retuns a jwt if the user has authenticated", func() {
			ctx := context.Background()
			token := "abc123def456"
			ghAuthClient.GetDeviceCodeAuthStatusStub = func(s string) (string, error) {
				return token, nil
			}
			res, err := appsClient.GetGithubAuthStatus(ctx, &pb.GetGithubAuthStatusRequest{DeviceCode: "somedevicecode"})
			Expect(err).NotTo(HaveOccurred())

			verified, err := auth.NewJwtClient(secretKey).VerifyJWT(res.AccessToken)
			Expect(err).NotTo(HaveOccurred())
			Expect(verified.ProviderToken).To(Equal(token))
		})
		It("returns an error other than ErrAuthPending", func() {
			ctx := context.Background()
			someErr := errors.New("some other err")
			ghAuthClient.GetDeviceCodeAuthStatusStub = func(s string) (string, error) {
				return "", someErr
			}
			res, err := appsClient.GetGithubAuthStatus(ctx, &pb.GetGithubAuthStatusRequest{DeviceCode: "somedevicecode"})
			Expect(err).To(HaveOccurred())
			st, ok := status.FromError(err)
			Expect(ok).To(BeTrue(), "could not get status from err")
			Expect(st.Message()).To(ContainSubstring(someErr.Error()))
			Expect(res).To(BeNil())
		})
	})

	Describe("ParseRepoURL", func() {
		type expected struct {
			provider pb.GitProvider
			owner    string
			name     string
		}
		DescribeTable("parses a repo url", func(uri string, e expected) {
			res, err := appsClient.ParseRepoURL(context.Background(), &pb.ParseRepoURLRequest{
				Url: uri,
			})
			Expect(err).NotTo(HaveOccurred())
			Expect(res.Provider).To(Equal(e.provider))
			Expect(res.Owner).To(Equal(e.owner))
			Expect(res.Name).To(Equal(e.name))
		},
			Entry("github+ssh", "git@github.com:some-org/my-repo.git", expected{
				provider: pb.GitProvider_GitHub,
				owner:    "some-org",
				name:     "my-repo",
			}),
			Entry("gitlab+ssh", "git@gitlab.com:other-org/cool-repo.git", expected{
				provider: pb.GitProvider_GitLab,
				owner:    "other-org",
				name:     "cool-repo",
			}),
		)

		It("returns an error on an invalid URL", func() {
			_, err := appsClient.ParseRepoURL(context.Background(), &pb.ParseRepoURLRequest{
				Url: "not-a  -valid-url",
			})
			Expect(err).To(HaveOccurred(), "should have gotten an invalid arg error")
			s, ok := status.FromError(err)
			Expect(ok).To(BeTrue(), "could not get status from error")
			Expect(s.Code()).To(Equal(codes.InvalidArgument))
		})
	})
	Describe("GetGitlabAuthURL", func() {
		It("returns the gitlab url", func() {
			urlString := "http://gitlab.com/oauth/authorize"
			authUrl, err := url.Parse(urlString)
			Expect(err).NotTo(HaveOccurred())

			glAuthClient.AuthURLReturns(*authUrl, nil)

			res, err := appsClient.GetGitlabAuthURL(context.Background(), &pb.GetGitlabAuthURLRequest{
				RedirectUri: "http://example.com/oauth/fake",
			})
			Expect(err).NotTo(HaveOccurred())

			u, err := url.Parse(res.Url)
			Expect(err).NotTo(HaveOccurred())
			Expect(u.String()).To(Equal(urlString))
		})
	})

	Describe("AuthorizeGitlab", func() {
		It("exchanges a token", func() {
			token := "some-token"
			glAuthClient.ExchangeCodeReturns(&authtypes.TokenResponseState{AccessToken: token}, nil)

			res, err := appsClient.AuthorizeGitlab(context.Background(), &pb.AuthorizeGitlabRequest{
				RedirectUri: "http://example.com/oauth/callback",
				Code:        "some-challenge-code",
			})
			Expect(err).NotTo(HaveOccurred())

			claims, err := jwtClient.VerifyJWT(res.Token)
			Expect(err).NotTo(HaveOccurred())

			Expect(claims.ProviderToken).To(Equal(token))
		})
		It("returns an error if the exchange fails", func() {
			e := errors.New("some code exchange error")
			glAuthClient.ExchangeCodeReturns(nil, e)

			_, err := appsClient.AuthorizeGitlab(context.Background(), &pb.AuthorizeGitlabRequest{
				RedirectUri: "http://example.com/oauth/callback",
				Code:        "some-challenge-code",
			})
			status, ok := status.FromError(err)
			Expect(ok).To(BeTrue(), "could not get status from error response")
			Expect(status.Code()).To(Equal(codes.Unknown))
			Expect(err.Error()).To(ContainSubstring(e.Error()))
		})
	})
	DescribeTable("ValidateProviderToken", func(provider pb.GitProvider, ctx context.Context, errResponse error, expectedCode codes.Code, valid bool) {
		glAuthClient.ValidateTokenReturns(errResponse)
		ghAuthClient.ValidateTokenReturns(errResponse)

		res, err := appsClient.ValidateProviderToken(ctx, &pb.ValidateProviderTokenRequest{
			Provider: provider,
		})

		if errResponse != nil {
			Expect(err).To(HaveOccurred())
			s, ok := status.FromError(err)
			Expect(ok).To(BeTrue(), "could not get status from error")
			Expect(s.Code()).To(Equal(expectedCode))
			return
		}

		Expect(err).NotTo(HaveOccurred())

		if valid {
			// Note that res is nil when there is an error.
			// Only check a field in `res` when valid, else this will panic
			Expect(res.Valid).To(BeTrue())
		}
	},
		Entry("bad gitlab token", pb.GitProvider_GitLab, contextWithAuth(context.Background()), errors.New("this token is bad"), codes.InvalidArgument, false),
		Entry("good gitlab token", pb.GitProvider_GitLab, contextWithAuth(context.Background()), nil, codes.OK, true),
		Entry("bad github token", pb.GitProvider_GitHub, contextWithAuth(context.Background()), errors.New("this token is bad"), codes.InvalidArgument, false),
		Entry("good github token", pb.GitProvider_GitHub, contextWithAuth(context.Background()), nil, codes.OK, true),
		Entry("no gitops jwt", pb.GitProvider_GitHub, context.Background(), errors.New("unauth error"), codes.Unauthenticated, false),
	)

	Describe("middleware", func() {
		Describe("logging", func() {
			var sink *fakelogr.LogSink
			var log logr.Logger
			var appsSrv pb.ApplicationsServer
			var mux *runtime.ServeMux
			var httpHandler http.Handler
			var err error

			BeforeEach(func() {
				log, sink = testutils.MakeFakeLogr()
				k8s := fake.NewClientBuilder().WithScheme(kube.CreateScheme()).Build()

				rand.Seed(time.Now().UnixNano())
				secretKey := rand.String(20)

				fakeFactory := &servicesfakes.FakeFactory{}

				cfg := server.ApplicationsConfig{
					Logger:    log,
					JwtClient: auth.NewJwtClient(secretKey),
					Factory:   fakeFactory,
				}

				fakeClientGetter := kubefakes.NewFakeClientGetter(k8s)
				appsSrv = server.NewApplicationsServer(&cfg, server.WithClientGetter(fakeClientGetter))
				mux = runtime.NewServeMux(middleware.WithGrpcErrorLogging(log))
				httpHandler = middleware.WithLogging(log, mux)
				err = pb.RegisterApplicationsHandlerServer(context.Background(), mux, appsSrv)
				Expect(err).NotTo(HaveOccurred())
			})

			It("logs invalid requests", func() {
				ts := httptest.NewServer(httpHandler)
				defer ts.Close()

				// Test a 404 here
				path := "/foo"
				url := ts.URL + path

				res, err := http.Get(url)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusNotFound))

				Expect(sink.InfoCallCount()).To(BeNumerically(">", 0))
				vals := sink.WithValuesArgsForCall(0)

				expectedStatus := strconv.Itoa(res.StatusCode)

				list := formatLogVals(vals)
				Expect(list).To(ConsistOf("uri", path, "status", expectedStatus))
			})

			It("logs server errors", func() {
				err = pb.RegisterApplicationsHandlerServer(context.Background(), mux, pb.UnimplementedApplicationsServer{})
				Expect(err).NotTo(HaveOccurred())

				ts := httptest.NewServer(httpHandler)
				defer ts.Close()

				path := "/v1/featureflags"
				url := ts.URL + path

				res, err := http.Get(url)
				// err is still nil even if we get a 5XX.
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusNotImplemented))

				Expect(sink.ErrorCallCount()).To(BeNumerically(">", 0))
				vals := sink.WithValuesArgsForCall(0)
				list := formatLogVals(vals)

				expectedStatus := strconv.Itoa(res.StatusCode)
				Expect(list).To(ConsistOf("uri", path, "status", expectedStatus))

				err, msg, _ := sink.ErrorArgsForCall(0)
				// This is the meat of this test case.
				// Check that the same error passed by kubeClient is logged.
				Expect(msg).To(Equal(middleware.ServerErrorText))
				Expect(err.Error()).To(ContainSubstring("GetFeatureFlags not implemented"))
			})

			It("logs ok requests", func() {
				ts := httptest.NewServer(httpHandler)
				defer ts.Close()

				// A valid URL for our server
				path := "/v1/featureflags"
				url := ts.URL + path

				res, err := http.Get(url)
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusOK))

				Expect(sink.InfoCallCount()).To(BeNumerically(">", 0))
				_, msg, _ := sink.InfoArgsForCall(0)
				Expect(msg).To(ContainSubstring(middleware.RequestOkText))

				vals := sink.WithValuesArgsForCall(0)
				list := formatLogVals(vals)

				expectedStatus := strconv.Itoa(res.StatusCode)
				Expect(list).To(ConsistOf("uri", path, "status", expectedStatus))
			})

			It("Authorize fails generating jwt token", func() {
				fakeJWTToken := &authfakes.FakeJWTClient{}
				fakeJWTToken.GenerateJWTStub = func(duration time.Duration, name gitproviders.GitProviderName, s22 string) (string, error) {
					return "", fmt.Errorf("some error")
				}

				factory := &servicesfakes.FakeFactory{}
				fakeKubeGetter := kubefakes.NewFakeKubeGetter(&kubefakes.FakeKube{})

				appsSrv = server.NewApplicationsServer(
					&server.ApplicationsConfig{Factory: factory, JwtClient: fakeJWTToken},
					server.WithKubeGetter(fakeKubeGetter))
				mux = runtime.NewServeMux(middleware.WithGrpcErrorLogging(log))
				httpHandler = middleware.WithLogging(log, mux)
				err = pb.RegisterApplicationsHandlerServer(context.Background(), mux, appsSrv)
				Expect(err).NotTo(HaveOccurred())

				ts := httptest.NewServer(httpHandler)
				defer ts.Close()

				// A valid URL for our server
				path := "/v1/authenticate/github"
				url := ts.URL + path

				res, err := http.Post(url, "application/json", strings.NewReader(`{"accessToken":"sometoken"}`))
				Expect(err).NotTo(HaveOccurred())
				Expect(res.StatusCode).To(Equal(http.StatusInternalServerError))

				bts, err := ioutil.ReadAll(res.Body)
				Expect(err).NotTo(HaveOccurred())

				Expect(bts).To(MatchJSON(`{"code": 13,"message": "error generating jwt token. some error","details": []}`))

				Expect(sink.InfoCallCount()).To(BeNumerically(">", 0))
				_, msg, _ := sink.InfoArgsForCall(0)
				Expect(msg).To(ContainSubstring(middleware.ServerErrorText))

				vals := sink.WithValuesArgsForCall(0)
				list := formatLogVals(vals)

				expectedStatus := strconv.Itoa(res.StatusCode)
				Expect(list).To(ConsistOf("uri", path, "status", expectedStatus))
			})
		})
	})
})

func formatLogVals(vals []interface{}) []string {
	list := []string{}

	for _, v := range vals {
		// vals is a slice of empty interfaces. convert them.
		s, ok := v.(string)
		if !ok {
			// Last value is a status code represented as an int
			n := v.(int)
			s = strconv.Itoa(n)
		}

		list = append(list, s)
	}

	return list
}

func contextWithAuth(ctx context.Context) context.Context {
	md := metadata.New(map[string]string{middleware.GRPCAuthMetadataKey: "mytoken"})
	ctx = metadata.NewOutgoingContext(ctx, md)

	return ctx
}

func TestGetFeatureFlags(t *testing.T) {
	type Data struct {
		Flags map[string]string
	}

	tests := []struct {
		name     string
		envSet   func()
		envUnset func()
		state    []client.Object
		result   map[string]string
	}{
		{
			name: "Auth enabled",
			envSet: func() {
				os.Setenv("WEAVE_GITOPS_AUTH_ENABLED", "true")
			},
			envUnset: func() {
				os.Unsetenv("WEAVE_GITOPS_AUTH_ENABLED")
			},
			state: []client.Object{},
			result: map[string]string{
				"WEAVE_GITOPS_AUTH_ENABLED": "true",
				"CLUSTER_USER_AUTH":         "false",
				"OIDC_AUTH":                 "false",
			},
		},
		{
			name: "Auth disabled",
			envSet: func() {
				os.Setenv("WEAVE_GITOPS_AUTH_ENABLED", "false")
			},
			envUnset: func() {
				os.Unsetenv("WEAVE_GITOPS_AUTH_ENABLED")
			},
			state: []client.Object{},
			result: map[string]string{
				"WEAVE_GITOPS_AUTH_ENABLED": "false",
				"CLUSTER_USER_AUTH":         "false",
				"OIDC_AUTH":                 "false",
			},
		},
		{
			name:     "Auth not set",
			envSet:   func() {},
			envUnset: func() {},
			state:    []client.Object{},
			result: map[string]string{
				"WEAVE_GITOPS_AUTH_ENABLED": "",
				"CLUSTER_USER_AUTH":         "false",
				"OIDC_AUTH":                 "false",
			},
		},
		{
			name:     "Cluster auth secret set",
			envSet:   func() {},
			envUnset: func() {},
			state:    []client.Object{&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: "flux-system", Name: "cluster-user-auth"}}},
			result: map[string]string{
				"WEAVE_GITOPS_AUTH_ENABLED": "",
				"CLUSTER_USER_AUTH":         "true",
				"OIDC_AUTH":                 "false",
			},
		},
		{
			name:     "Cluster auth secret not set",
			envSet:   func() {},
			envUnset: func() {},
			state:    []client.Object{},
			result: map[string]string{
				"WEAVE_GITOPS_AUTH_ENABLED": "",
				"CLUSTER_USER_AUTH":         "false",
				"OIDC_AUTH":                 "false",
			},
		},
		{
			name:     "OIDC secret set",
			envSet:   func() {},
			envUnset: func() {},
			state:    []client.Object{&corev1.Secret{ObjectMeta: metav1.ObjectMeta{Namespace: "flux-system", Name: "oidc-auth"}}},
			result: map[string]string{
				"WEAVE_GITOPS_AUTH_ENABLED": "",
				"CLUSTER_USER_AUTH":         "false",
				"OIDC_AUTH":                 "true",
			},
		},
		{
			name:     "OIDC secret not set",
			envSet:   func() {},
			envUnset: func() {},
			state:    []client.Object{},
			result: map[string]string{
				"WEAVE_GITOPS_AUTH_ENABLED": "",
				"CLUSTER_USER_AUTH":         "false",
				"OIDC_AUTH":                 "false",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			log, _ := testutils.MakeFakeLogr()
			mux := runtime.NewServeMux(middleware.WithGrpcErrorLogging(log))

			cfg := server.ApplicationsConfig{
				Logger: logr.Discard(),
			}

			k8s := fake.NewClientBuilder().WithScheme(kube.CreateScheme()).WithObjects(tt.state...).Build()
			fakeClientGetter := kubefakes.NewFakeClientGetter(k8s)
			appSrv := server.NewApplicationsServer(&cfg, server.WithClientGetter(fakeClientGetter))
			err = pb.RegisterApplicationsHandlerServer(context.Background(), mux, appSrv)

			httpHandler := middleware.WithLogging(log, mux)

			ts := httptest.NewServer(httpHandler)
			defer ts.Close()

			path := "/v1/featureflags"
			url := ts.URL + path

			tt.envSet()
			defer tt.envUnset()

			res, err := http.Get(url)
			assert.NoError(t, err)
			assert.Equal(t, http.StatusOK, res.StatusCode)

			var data Data
			err = json.NewDecoder(res.Body).Decode(&data)
			assert.NoError(t, err)
			assert.Equal(t, data.Flags, tt.result)
		})
	}
}
