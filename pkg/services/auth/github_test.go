package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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
})
