package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/pkg/server/middleware"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/authfakes"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"google.golang.org/grpc/metadata"
)

var (
	jwtClient      *authfakes.FakeJWTClient
	defaultHandler http.HandlerFunc
	log            logr.Logger
)

var _ = Describe("WithProviderToken", func() {
	_ = BeforeEach(func() {
		jwtClient = &authfakes.FakeJWTClient{
			VerifyJWTStub: func(s string) (*auth.Claims, error) {
				return &auth.Claims{
					ProviderToken: "provider-token",
				}, nil
			},
		}

		defaultHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
		log, _ = testutils.MakeFakeLogr()
	})

	It("does nothing when no token is passed", func() {
		midware := middleware.WithProviderToken(jwtClient, defaultHandler, log)

		req := httptest.NewRequest(http.MethodGet, "http://www.foo.com/", nil)
		res := httptest.NewRecorder()
		midware.ServeHTTP(res, req)

		Expect(res.Result().StatusCode).To(Equal(http.StatusOK))
		Expect(jwtClient.VerifyJWTCallCount()).To(Equal(0))
	})

	It("extracts JWT token from the header", func() {
		midware := middleware.WithProviderToken(jwtClient, defaultHandler, log)

		req := httptest.NewRequest(http.MethodGet, "http://www.foo.com", nil)
		req.Header.Add(middleware.GitProviderTokenHeader, "token my-jwt-token")

		res := httptest.NewRecorder()

		midware.ServeHTTP(res, req)

		Expect(jwtClient.VerifyJWTArgsForCall(0)).To(Equal("my-jwt-token"))
	})

	It("passes the request through when a token is invalid", func() {
		jwtClient.VerifyJWTStub = func(s string) (*auth.Claims, error) {
			return nil, auth.ErrUnauthorizedToken
		}

		midware := middleware.WithProviderToken(jwtClient, defaultHandler, log)
		req := httptest.NewRequest(http.MethodGet, "http://www.foo.com", nil)
		req.Header.Add(middleware.GitProviderTokenHeader, "token my-jwt-token")

		res := httptest.NewRecorder()
		// Ensure a 401 is not returned, since we pass invalid tokens through.
		nextEndpointRes := http.StatusInternalServerError
		res.WriteHeader(nextEndpointRes)

		midware.ServeHTTP(res, req)

		Expect(res.Result().StatusCode).To(Equal(nextEndpointRes))
	})

	It("passes the provider token into the context", func() {
		var request *http.Request

		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			request = r
		})

		midware := middleware.WithProviderToken(jwtClient, next, log)
		req := httptest.NewRequest(http.MethodGet, "http://www.foo.com", nil)
		req.Header.Add(middleware.GitProviderTokenHeader, "token my-jwt-token")

		res := httptest.NewRecorder()

		midware.ServeHTTP(res, req)

		providerToken, err := middleware.ExtractProviderToken(request.Context())
		Expect(err).ToNot(HaveOccurred())

		Expect(providerToken.AccessToken).To(Equal("provider-token"))
	})
})

var _ = Describe("ExtractProviderToken", func() {
	_ = BeforeEach(func() {
		jwtClient = &authfakes.FakeJWTClient{
			VerifyJWTStub: func(s string) (*auth.Claims, error) {
				return &auth.Claims{
					ProviderToken: "",
				}, nil
			},
		}
	})

	It("errors out when no provider token in context", func() {
		var request *http.Request
		next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			request = r
		})

		midware := middleware.WithProviderToken(jwtClient, next, log)
		req := httptest.NewRequest(http.MethodGet, "http://www.foo.com", nil)
		req.Header.Add(middleware.GitProviderTokenHeader, "token my-jwt-token")

		res := httptest.NewRecorder()

		midware.ServeHTTP(res, req)

		_, err := middleware.ExtractProviderToken(request.Context())
		Expect(err).To(MatchError("no token specified"))
	})
	It("extracts an auth token from grpc metadata", func() {
		tokenStr := "mytoken"
		md := metadata.New(map[string]string{middleware.GRPCAuthMetadataKey: tokenStr})
		ctx := metadata.NewIncomingContext(context.Background(), md)
		token, err := middleware.ExtractProviderToken(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(token.AccessToken).To(Equal(tokenStr))
	})
})
