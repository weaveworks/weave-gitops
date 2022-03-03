package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/internal"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/types"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/types/typesfakes"
	fakehttp "github.com/weaveworks/weave-gitops/pkg/vendorfakes/http"
)

func roundTripperErrorStub(*http.Request) (*http.Response, error) {
	return &http.Response{}, errors.New("ka-boom")
}

var _ = Describe("Gitlab auth flow", func() {

	var redirectUrl string
	var client *http.Client
	var authFlow types.AuthFlow
	var mockRoundTripper *fakehttp.FakeRoundTripper
	var mockHandler *fakehttp.FakeHandler
	var tokenState types.TokenResponseState

	expectedTokenRequest := func() *http.Request {
		tokenUrl := internal.GitlabTokenUrl(redirectUrl, "12345", authFlow.(*gitlabAuthFlow).codeVerifier)
		expectedRequest, _ := http.NewRequest(http.MethodPost, tokenUrl.String(), strings.NewReader(""))
		return expectedRequest
	}

	assertRoundTripCall := func(index int) {
		Expect(mockRoundTripper.RoundTripCallCount()).To(Equal(1))

		expectedRequest := expectedTokenRequest()
		arg := mockRoundTripper.RoundTripArgsForCall(0)
		Expect(arg.Method).To(Equal(expectedRequest.Method))
		Expect(arg.URL).To(Equal(expectedRequest.URL))
	}

	assertServerHTTPCall := func(w http.ResponseWriter, r *http.Request) {
		Expect(mockHandler.ServeHTTPCallCount()).To(Equal(1))
		receivedWriter, receivedRequest := mockHandler.ServeHTTPArgsForCall(0)
		Expect(receivedWriter).To(Equal(w))
		Expect(receivedRequest).To(Equal(r))
	}

	BeforeEach(func() {
		mockRoundTripper = &fakehttp.FakeRoundTripper{}
		mockHandler = &fakehttp.FakeHandler{}
		client = &http.Client{
			Transport: mockRoundTripper,
		}

		redirectUrl = "https://weave.works/call-me-back"
		authFlow, _ = NewGitlabAuthFlow(redirectUrl, client)
		tokenState = types.TokenResponseState{}
	})

	It("client throws an error", func() {
		w := httptest.NewRecorder()
		mockRoundTripper.RoundTripStub = roundTripperErrorStub

		r, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?code=12345", redirectUrl), strings.NewReader(""))
		callbackHandler := authFlow.CallbackHandler(&tokenState, mockHandler)
		callbackHandler.ServeHTTP(w, r)

		assertRoundTripCall(0)
		assertServerHTTPCall(w, r)
		Expect(w.Code).To(Equal(http.StatusInternalServerError))
		Expect(w.Body.String()).To(Equal(http.StatusText(http.StatusInternalServerError)))
		Expect(tokenState.HttpStatusCode).To(Equal(http.StatusInternalServerError))
		_, expectedErr := client.Do(expectedTokenRequest())
		Expect(tokenState.Err).To(MatchError(fmt.Errorf("gitlab token requeset client issue: %w", expectedErr)))
	})

	It("client returns status code 401", func() {
		w := httptest.NewRecorder()
		mockRoundTripper.RoundTripStub = func(request *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusUnauthorized,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil
		}

		r, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?code=12345", redirectUrl), strings.NewReader(""))
		callbackHandler := authFlow.CallbackHandler(&tokenState, mockHandler)
		callbackHandler.ServeHTTP(w, r)

		assertRoundTripCall(0)
		assertServerHTTPCall(w, r)
		Expect(w.Code).To(Equal(http.StatusUnauthorized))
		Expect(w.Body.String()).To(Equal(http.StatusText(http.StatusUnauthorized)))

		expectedToken := types.TokenResponseState{
			HttpStatusCode: http.StatusUnauthorized,
			Err:            errors.New(http.StatusText(http.StatusUnauthorized)),
		}
		Expect(tokenState).To(Equal(expectedToken))
	})

	It("json decoding throws an error", func() {
		w := httptest.NewRecorder()

		mangledJson := `{"access_token": "abc-123", mangled-json": true,,,}`
		reader := io.NopCloser(strings.NewReader(mangledJson))
		mockRoundTripper.RoundTripStub = func(request *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       reader,
			}, nil
		}

		r, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?code=12345", redirectUrl), strings.NewReader(""))
		callbackHandler := authFlow.CallbackHandler(&tokenState, mockHandler)
		callbackHandler.ServeHTTP(w, r)

		assertRoundTripCall(0)
		assertServerHTTPCall(w, r)
		Expect(w.Code).To(Equal(http.StatusInternalServerError))
		Expect(w.Body.String()).To(Equal(http.StatusText(http.StatusInternalServerError)))
		Expect(tokenState.HttpStatusCode).To(Equal(http.StatusInternalServerError))
		expectedErr := json.NewDecoder(io.NopCloser(strings.NewReader(mangledJson))).Decode(&internal.GitlabTokenResponse{})
		Expect(tokenState.Err).To(MatchError(fmt.Errorf("gitlab token response json decode: %w", expectedErr)))
	})

	It("parses valid JSON returning a valid response", func() {
		w := httptest.NewRecorder()

		validJson := `{"access_token": "abc-123", "token_type": "shiny", "expires_in": 5, "refresh_token": "xyz-456", "created_at": 4}`
		reader := io.NopCloser(strings.NewReader(validJson))
		mockRoundTripper.RoundTripStub = func(request *http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       reader,
			}, nil
		}

		r, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?code=12345", redirectUrl), strings.NewReader(""))
		callbackHandler := authFlow.CallbackHandler(&tokenState, mockHandler)
		callbackHandler.ServeHTTP(w, r)

		assertRoundTripCall(0)
		assertServerHTTPCall(w, r)
		Expect(w.Code).To(Equal(http.StatusOK))
		Expect(w.Body.String()).To(Equal(""))

		expectedToken := types.TokenResponseState{
			AccessToken:      "abc-123",
			TokenType:        "shiny",
			ExpiresInSeconds: time.Second * 5,
			RefreshToken:     "xyz-456",
			CreatedAt:        4,
			HttpStatusCode:   http.StatusOK,
			Err:              nil,
		}
		Expect(tokenState).To(Equal(expectedToken))
	})

})

func cliOutputSuccess() string {
	var cliOutputSuccess = []string{
		"Starting authorization server:\n",
		"Visit this URL to authenticate with Gitlab:\n\n",
		"https://weave.works/unit/test\n\n",
		"Waiting for authentication flow completion...\n\n",
		"Shutting the server down...\n",
		"Server shutdown complete!\n",
	}

	return strings.Join(cliOutputSuccess, "")
}

var _ = Describe("Shutdown server handler", func() {
	It("Completes the wait group", func() {
		token := types.TokenResponseState{Err: errors.New("ka-boom"), HttpStatusCode: http.StatusInternalServerError}
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "https://weave.works/unit/test", strings.NewReader(""))
		wg := &sync.WaitGroup{}
		wg.Add(1)

		handler := shutdownServerForCLI(&token, wg)
		handler.ServeHTTP(w, r)
		wg.Wait()
		Expect(w.Body.String()).To(Equal(serverShutdownErrorMessage))
	})

	It("writes out a success message when it receives a 200", func() {
		token := types.TokenResponseState{HttpStatusCode: http.StatusOK}
		w := httptest.NewRecorder()
		r, _ := http.NewRequest(http.MethodGet, "https://weave.works/unit/test", strings.NewReader(""))
		wg := &sync.WaitGroup{}
		wg.Add(1)

		handler := shutdownServerForCLI(&token, wg)
		handler.ServeHTTP(w, r)
		wg.Wait()
		Expect(w.Body.String()).To(Equal(serverShutdownSuccessMessage))
	})
})

var _ = Describe("Gitlab cli auth flow", func() {
	var client *http.Client
	var mockAuthFlow *typesfakes.FakeAuthFlow
	var mockRoundTripper *fakehttp.FakeRoundTripper
	var mockHandler *fakehttp.FakeHandler
	var writer *bytes.Buffer

	var cliOutputWithError = func(err error) string {
		cliOutput := []string{
			cliOutputSuccess(),
			"There was an issue going through the Gitlab authentication flow:\n\n",
			err.Error(),
		}
		return strings.Join(cliOutput, "")
	}

	BeforeEach(func() {
		mockAuthFlow = &typesfakes.FakeAuthFlow{}
		mockRoundTripper = &fakehttp.FakeRoundTripper{}
		mockHandler = &fakehttp.FakeHandler{}
		client = &http.Client{
			Transport: mockRoundTripper,
		}
		writer = bytes.NewBufferString("")
	})

	AfterEach(func() {
		http.DefaultServeMux = new(http.ServeMux)
	})

	It("auth flow authorize throws an error", func() {
		mockAuthFlow.AuthorizeStub = func(ctx context.Context) (*http.Request, error) { return nil, errors.New("ka-boom") }
		cliHandler := NewGitlabAuthFlowHandler(client, mockAuthFlow)
		value, err := cliHandler(context.Background(), writer)

		Expect(err).To(MatchError(fmt.Errorf("could not do code request: %w", errors.New("ka-boom"))))
		Expect(value).To(BeEmpty())
	})

	It("client throws an error", func() {
		mockAuthFlow.AuthorizeStub = func(ctx context.Context) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, "https://weave.works/unit/test", strings.NewReader(""))
		}

		mockAuthFlow.CallbackHandlerStub = func(state *types.TokenResponseState, next http.Handler) http.Handler {
			mockHandler.ServeHTTPStub = func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				state.AccessToken = "keep-it-secret-keep-it-safe"
				next.ServeHTTP(w, r)
			}

			return http.HandlerFunc(mockHandler.ServeHTTPStub)
		}

		mockRoundTripper.RoundTripStub = roundTripperErrorStub

		cliHandler := NewGitlabAuthFlowHandler(client, mockAuthFlow)
		value, err := cliHandler(context.Background(), writer)

		Expect(err).ToNot(BeNil())
		Expect(value).To(BeEmpty())

		expectedRequest, _ := http.NewRequest(http.MethodGet, "https://weave.works/unit/test", strings.NewReader(""))
		_, expectedErr := client.Do(expectedRequest)
		wrappedErr := fmt.Errorf("gitlab auth client error: %w", expectedErr)
		Expect(err).To(MatchError(wrappedErr))
		Expect(writer.String()).To(Equal(cliOutputWithError(wrappedErr)))
	})
})

var _ = Describe("Gitlab auth flow end-to-end", func() {
	var client *http.Client
	var mockAuthFlow *typesfakes.FakeAuthFlow
	var mockRoundTripper *fakehttp.FakeRoundTripper
	var mockHandler *fakehttp.FakeHandler
	var writer *bytes.Buffer

	BeforeEach(func() {
		mockAuthFlow = &typesfakes.FakeAuthFlow{}
		mockRoundTripper = &fakehttp.FakeRoundTripper{}
		mockHandler = &fakehttp.FakeHandler{}
		client = &http.Client{
			Transport: mockRoundTripper,
		}
		writer = bytes.NewBufferString("")
	})

	AfterEach(func() {
		http.DefaultServeMux = new(http.ServeMux)
	})

	It("access token is returned at the end of the flow", func() {
		mockAuthFlow.AuthorizeStub = func(ctx context.Context) (*http.Request, error) {
			return http.NewRequest(http.MethodGet, "https://weave.works/unit/test", strings.NewReader(""))
		}

		mockAuthFlow.CallbackHandlerStub = func(state *types.TokenResponseState, next http.Handler) http.Handler {
			mockHandler.ServeHTTPStub = func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				state.HttpStatusCode = http.StatusOK
				state.AccessToken = "keep-it-secret-keep-it-safe"
				next.ServeHTTP(w, r)
			}

			return http.HandlerFunc(mockHandler.ServeHTTPStub)
		}

		mockRoundTripper.RoundTripStub = func(request *http.Request) (*http.Response, error) {
			return &http.Response{}, nil
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		cliHandler := NewGitlabAuthFlowHandler(client, mockAuthFlow)

		var returnedValue string
		var returnedErr error
		go func(v *string, e *error) {
			value, err := cliHandler(context.Background(), writer)
			*v = value
			*e = err
			wg.Done()
		}(&returnedValue, &returnedErr)

		sampleReq, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s?code=12345", internal.GitlabRedirectUriCLI), strings.NewReader(""))
		dc := http.DefaultClient

		var sampleRes *http.Response
		var clientErr error
		// The server might not be ready when we reach here, we'll use a periodic tick that will wait for a successful
		// call to the local server set up in the go routine above.
		callServerTick := time.Tick(500 * time.Millisecond)
		for {
			<-callServerTick
			sampleRes, clientErr = dc.Do(sampleReq)
			if clientErr == nil {
				break
			}
		}

		wg.Wait()
		Expect(sampleRes.StatusCode).To(Equal(http.StatusOK))
		defer sampleRes.Body.Close()
		b, _ := ioutil.ReadAll(sampleRes.Body)
		Expect(string(b)).To(Equal(serverShutdownSuccessMessage))
		Expect(returnedErr).To(BeNil())
		Expect(returnedValue).To(Equal("keep-it-secret-keep-it-safe"))
		Expect(writer.String()).To(Equal(cliOutputSuccess()))
	})
})

var _ = Describe("GitlabAuthClient", func() {
	It("AuthURL", func() {
		rt := fakehttp.FakeRoundTripper{}
		rt.RoundTripReturns(&http.Response{}, nil)
		c := NewGitlabAuthClient(&http.Client{Transport: &rt})

		u, err := c.AuthURL(context.Background(), "http://example.com:9999/oauth/callback")
		Expect(err).NotTo(HaveOccurred())
		Expect(u.Hostname()).To(Equal("gitlab.com"))
		Expect(u.Scheme).To(Equal("https"))
	})
	It("ExchangeCode", func() {
		rt := fakehttp.FakeRoundTripper{}
		res := &http.Response{StatusCode: http.StatusOK}

		rs := &internal.GitlabTokenResponse{
			AccessToken: "this-is-a-secret",
			ExpiresIn:   1600,
		}
		b, err := json.Marshal(rs)
		Expect(err).NotTo(HaveOccurred())

		res.Body = ioutil.NopCloser(bytes.NewReader(b))

		rt.RoundTripReturns(res, nil)

		c := NewGitlabAuthClient(&http.Client{Transport: &rt})

		tokenState, err := c.ExchangeCode(context.Background(), "http://example.com/oauth/callback", "abc123def456")
		Expect(err).NotTo(HaveOccurred())

		Expect(tokenState.AccessToken).To(Equal(rs.AccessToken))
		Expect(tokenState.ExpiresInSeconds).To(Equal(time.Duration(rs.ExpiresIn) * time.Second))
	})
	Describe("ValidateToken", func() {
		It("returns an error when a 401 is returned", func() {
			rt := fakehttp.FakeRoundTripper{}
			rt.RoundTripReturns(&http.Response{StatusCode: http.StatusUnauthorized}, nil)
			c := NewGitlabAuthClient(&http.Client{Transport: &rt})

			Expect(c.ValidateToken(context.Background(), "sometoken")).To(HaveOccurred())
		})
		It("does not return an error when a token is valid", func() {
			rt := fakehttp.FakeRoundTripper{}
			rt.RoundTripReturns(&http.Response{StatusCode: http.StatusOK}, nil)
			c := NewGitlabAuthClient(&http.Client{Transport: &rt})

			Expect(c.ValidateToken(context.Background(), "sometoken")).NotTo(HaveOccurred())
		})
	})
})
