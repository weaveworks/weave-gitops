package internal

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Gitlab authorize url", func() {
	It("returns url with all required parameters", func() {
		codeVerifier, err := NewCodeVerifier(5, 10)
		Expect(err).NotTo(HaveOccurred())

		scopes := []string{"api", "read_user", "profile"}
		u, err := GitlabAuthorizeURL("https://weave.works/test", scopes, codeVerifier)
		Expect(err).NotTo(HaveOccurred())

		Expect(u.Scheme).To(Equal("https"))
		Expect(u.Host).To(Equal("gitlab.com"))
		Expect(u.Path).To(Equal("/oauth/authorize"))

		params := u.Query()
		Expect(len(params)).To(Equal(6))
		Expect(params.Get("client_id")).To(Equal(gitlabClientID))
		Expect(params.Get("response_type")).To(Equal("code"))
		Expect(params.Get("scope")).To(Equal("api read_user profile"))

		codeChallenge, err := codeVerifier.CodeChallenge()
		Expect(err).NotTo(HaveOccurred())

		Expect(params.Get("code_challenge")).To(Equal(codeChallenge))
		Expect(params.Get("code_challenge_method")).To(Equal("S256"))
		Expect(params.Get("redirect_uri")).To(Equal("https://weave.works/test"))
	})
})

var _ = Describe("Gitlab token", func() {
	It("returns url with all required parameters", func() {
		codeVerifier, err := NewCodeVerifier(5, 10)
		Expect(err).NotTo(HaveOccurred())

		u := GitlabTokenURL("https://weave.works/test", "12345", codeVerifier)
		Expect(u.Scheme).To(Equal("https"))
		Expect(u.Host).To(Equal("gitlab.com"))
		Expect(u.Path).To(Equal("/oauth/token"))

		params := u.Query()
		Expect(len(params)).To(Equal(6))
		Expect(params.Get("client_id")).To(Equal(gitlabClientID))
		Expect(params.Get("redirect_uri")).To(Equal("https://weave.works/test"))
		Expect(params.Get("grant_type")).To(Equal("authorization_code"))
		Expect(params.Get("code_verifier")).To(Equal(codeVerifier.RawValue()))
		Expect(params.Get("code")).To(Equal("12345"))
		Expect(params.Get("client_secret")).To(Equal(gitlabClientSecret))
	})
})
