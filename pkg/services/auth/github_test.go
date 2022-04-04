package auth

import (
	"context"
	"net/http"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/vendorfakes/fakehttp"
)

var _ = Describe("GitHub error parsing", func() {
	It("parses from a response", func() {
		err := parseGitHubError([]byte(`{"error":"device_flow_disabled","error_description":"Device Flow must be explicitly enabled for this App","error_uri":"https://docs.github.com"}`), http.StatusBadRequest)

		Expect(err).To(MatchError(`GitHub 400 - device_flow_disabled ("Device Flow must be explicitly enabled for this App") more information at https://docs.github.com`))
	})
})

var _ = Describe("ValidateToken", func() {
	It("returns unauthenticated on an invalid token", func() {
		rt := &fakehttp.RoundTripper{}
		gh := NewGithubAuthClient(&http.Client{Transport: rt})

		rt.RoundTripReturns(&http.Response{StatusCode: http.StatusUnauthorized}, nil)

		Expect(gh.ValidateToken(context.Background(), "sometoken")).To(HaveOccurred())
	})
	It("does not return an error when a token is valid", func() {
		rt := &fakehttp.RoundTripper{}
		gh := NewGithubAuthClient(&http.Client{Transport: rt})
		rt.RoundTripReturns(&http.Response{StatusCode: http.StatusOK}, nil)

		Expect(gh.ValidateToken(context.Background(), "sometoken")).NotTo(HaveOccurred())
	})
})
