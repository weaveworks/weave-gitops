package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/internal"
	"github.com/weaveworks/weave-gitops/pkg/vendorfakes/fakehttp"
)

var _ = Describe("GitlabAuthClient", func() {
	var rt fakehttp.RoundTripper

	BeforeEach(func() {
		rt = fakehttp.RoundTripper{}
	})

	It("AuthURL", func() {
		rt.RoundTripReturns(&http.Response{}, nil)
		c := auth.NewGitlabAuthClient(&http.Client{Transport: &rt})

		u, err := c.AuthURL(context.Background(), "http://example.com:9999/oauth/callback")
		Expect(err).NotTo(HaveOccurred())
		Expect(u.Hostname()).To(Equal("gitlab.com"))
		Expect(u.Scheme).To(Equal("https"))
	})

	It("ExchangeCode", func() {
		res := &http.Response{StatusCode: http.StatusOK}

		rs := &internal.GitlabTokenResponse{
			AccessToken: "this-is-a-secret",
			ExpiresIn:   1600,
		}
		b, err := json.Marshal(rs)
		Expect(err).NotTo(HaveOccurred())

		res.Body = ioutil.NopCloser(bytes.NewReader(b))

		rt.RoundTripReturns(res, nil)

		c := auth.NewGitlabAuthClient(&http.Client{Transport: &rt})

		tokenState, err := c.ExchangeCode(context.Background(), "http://example.com/oauth/callback", "abc123def456")
		Expect(err).NotTo(HaveOccurred())

		Expect(tokenState.AccessToken).To(Equal(rs.AccessToken))
		Expect(tokenState.ExpiresInSeconds).To(Equal(time.Duration(rs.ExpiresIn) * time.Second))
	})

	Describe("ValidateToken", func() {
		It("returns an error when a 401 is returned", func() {
			rt.RoundTripReturns(&http.Response{StatusCode: http.StatusUnauthorized}, nil)
			c := auth.NewGitlabAuthClient(&http.Client{Transport: &rt})

			Expect(c.ValidateToken(context.Background(), "sometoken")).To(HaveOccurred())
		})

		It("does not return an error when a token is valid", func() {
			rt.RoundTripReturns(&http.Response{StatusCode: http.StatusOK}, nil)
			c := auth.NewGitlabAuthClient(&http.Client{Transport: &rt})

			Expect(c.ValidateToken(context.Background(), "sometoken")).To(Succeed())
		})
	})
})
