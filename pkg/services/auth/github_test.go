package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	fakehttp "github.com/weaveworks/weave-gitops/pkg/vendorfakes/http"
)

type testServerTransport struct {
	testServeUrl string
	roundTripper http.RoundTripper
}

func (t *testServerTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	// Fake out the client but preserve the URL, as the URLs are key to validating that
	// the authHandler is working.
	tsUrl, err := url.Parse(t.testServeUrl)
	if err != nil {
		return nil, err
	}

	tsUrl.Path = r.URL.Path

	r.URL = tsUrl

	return t.roundTripper.RoundTrip(r)
}

// sleeper is a very lightweight fake sleep timer. Instead of faking out the system
// clock, we can accept `sleep` calls and keep track of how long we've slept.
type sleeper struct {
	mutex sync.Mutex
	time  time.Time
}

func (t *sleeper) sleep(d time.Duration) {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.time = t.time.Add(d)
}

func (t *sleeper) now() time.Time {
	t.mutex.Lock()
	defer t.mutex.Unlock()

	return t.time
}

var _ = Describe("Github Device Flow", func() {
	var ts *httptest.Server
	var client *http.Client
	token := "gho_sUpErSecRetToKeN"
	userCode := "ABC-123"
	verificationUri := "http://somegithuburl.com"

	var _ = BeforeEach(func() {
		ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			// Quick and dirty router to simulate the Github API
			if strings.Contains(r.URL.Path, "/device/code") {
				err := json.NewEncoder(w).Encode(&GithubDeviceCodeResponse{
					DeviceCode:      "123456789",
					UserCode:        userCode,
					VerificationURI: verificationUri,
					Interval:        1,
				})
				Expect(err).NotTo(HaveOccurred())

			}

			if strings.Contains(r.URL.Path, "/oauth/access_token") {
				err := json.NewEncoder(w).Encode(&githubAuthResponse{
					AccessToken: token,
					Error:       "",
				})
				Expect(err).NotTo(HaveOccurred())
			}
		}))

		client = ts.Client()
		client.Transport = &testServerTransport{testServeUrl: ts.URL, roundTripper: client.Transport}
	})

	var _ = AfterEach(func() {
		ts.Close()
	})

	It("does the auth flow", func() {
		authHandler := NewGithubDeviceFlowHandler(client)

		var cliOutput bytes.Buffer
		result, err := authHandler(context.Background(), &cliOutput)

		Expect(err).NotTo(HaveOccurred())
		Expect(result).To(Equal(token))
		// We need to ensure the user code and verification url are in the CLI ouput.
		// Check for the prescense of substrings to avoid failing tests on trivial output changes.
		Expect(cliOutput.String()).To(ContainSubstring(userCode))
		Expect(cliOutput.String()).To(ContainSubstring(verificationUri))
	})

	Describe("pollAuthStatus", func() {
		var rt *mockAuthRoundTripper
		var s *sleeper

		// pollTimes is a convenience function to convert from a series of polling intervals
		// to their respective polling timestamps, relative to the sleeper type's starting time
		pollTimes := func(intervals []time.Duration) []time.Time {
			zero := time.Time{}
			times := make([]time.Time, len(intervals))
			for index, interval := range intervals {
				switch index {
				case 0:
					times[index] = zero.Add(interval)
				default:
					times[index] = times[index-1].Add(interval)
				}
			}
			return times
		}

		drainPollTimes := func(pollChan <-chan time.Time) (result []time.Time) {
			for pollTime := range pollChan {
				result = append(result, pollTime)
			}
			return
		}

		Context("after a slow_down response from GitHub", func() {
			BeforeEach(func() {
				s = &sleeper{}
				rt = newMockRoundTripper(1, token, s.now)
				client.Transport = &testServerTransport{testServeUrl: ts.URL, roundTripper: rt}
			})

			It("retries with a longer interval", func() {
				interval := 5 * time.Second

				_, _ = pollAuthStatus(s.sleep, interval, client, "somedevicecode")

				expectedPollTimes := pollTimes([]time.Duration{
					interval,
					interval + 5*time.Second,
				})
				Expect(drainPollTimes(rt.callChan)).To(Equal(expectedPollTimes))
			})

			It("returns a token", func() {
				interval := 5 * time.Second

				resultToken, err := pollAuthStatus(s.sleep, interval, client, "somedevicecode")

				Expect(resultToken).To(Equal(token))
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("after several slow_down responses from GitHub", func() {
			var s *sleeper

			BeforeEach(func() {
				s = &sleeper{}
				rt = newMockRoundTripper(3, token, s.now)
				client.Transport = &testServerTransport{testServeUrl: ts.URL, roundTripper: rt}
			})

			It("keeps slowing down", func() {
				interval := 5 * time.Second

				_, _ = pollAuthStatus(s.sleep, interval, client, "somedevicecode")

				expectedPollTimes := pollTimes([]time.Duration{
					interval,
					interval + 5*time.Second,
					interval + 10*time.Second,
					interval + 15*time.Second,
				})
				Expect(drainPollTimes(rt.callChan)).To(Equal(expectedPollTimes))
			})
		})

	})
})

var _ = Describe("ValidateToken", func() {
	It("returns unauthenticated on an invalid token", func() {
		rt := &fakehttp.FakeRoundTripper{}
		gh := NewGithubAuthClient(&http.Client{Transport: rt})

		rt.RoundTripReturns(&http.Response{StatusCode: http.StatusUnauthorized}, nil)

		Expect(gh.ValidateToken(context.Background(), "sometoken")).To(HaveOccurred())
	})
	It("does not return an error when a token is valid", func() {
		rt := &fakehttp.FakeRoundTripper{}
		gh := NewGithubAuthClient(&http.Client{Transport: rt})
		rt.RoundTripReturns(&http.Response{StatusCode: http.StatusOK}, nil)

		Expect(gh.ValidateToken(context.Background(), "sometoken")).NotTo(HaveOccurred())
	})
})

type mockAuthRoundTripper struct {
	fn       func(r *http.Request) (*http.Response, error)
	calls    int
	callChan chan time.Time
}

func (rt *mockAuthRoundTripper) MockRoundTrip(fn func(r *http.Request) (*http.Response, error)) {
	rt.fn = fn
}

func (rt *mockAuthRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt.fn(r)
}

func newMockRoundTripper(pollCount int, token string, now func() time.Time) *mockAuthRoundTripper {
	rt := &mockAuthRoundTripper{calls: 0, callChan: make(chan time.Time, pollCount+1)}

	rt.MockRoundTrip(func(r *http.Request) (*http.Response, error) {
		b := bytes.NewBuffer(nil)

		var data githubAuthResponse

		switch {
		case rt.calls > pollCount:
			panic("mock API called after successful request")
		case rt.calls == pollCount:
			data = githubAuthResponse{AccessToken: token}
			rt.callChan <- now()
			close(rt.callChan)
		default:
			data = githubAuthResponse{Error: "slow_down"}
			rt.callChan <- now()
		}

		if err := json.NewEncoder(b).Encode(data); err != nil {
			return nil, err
		}

		res := &http.Response{
			Body: io.NopCloser(b),
		}

		res.StatusCode = http.StatusOK

		rt.calls = rt.calls + 1
		return res, nil
	})

	return rt
}
