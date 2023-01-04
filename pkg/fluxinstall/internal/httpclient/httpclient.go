package httpclient

import (
	"net/http"

	"github.com/hashicorp/go-cleanhttp"
)

// NewHTTPClient provides a pre-configured http.Client
// e.g. with relevant User-Agent header
func NewHTTPClient() *http.Client {
	client := cleanhttp.DefaultClient()

	cli := cleanhttp.DefaultPooledClient()
	cli.Transport = &userAgentRoundTripper{
		userAgent: "flux-install/0.1.0",
		inner:     cli.Transport,
	}

	return client
}

type userAgentRoundTripper struct {
	inner     http.RoundTripper
	userAgent string
}

func (rt *userAgentRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if _, ok := req.Header["User-Agent"]; !ok {
		req.Header.Set("User-Agent", rt.userAgent)
	}

	return rt.inner.RoundTrip(req)
}
