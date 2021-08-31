package auth

import (
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JWT tokens", func() {

	It("Verify should success", func() {

		provider := "github"
		token := "token"

		jwtToken, err := Generate(SecretKey, time.Millisecond, provider, token)
		Expect(err).NotTo(HaveOccurred())

		claims, err := Verify(SecretKey, jwtToken)
		Expect(err).NotTo(HaveOccurred())

		Expect(claims.Provider).To(Equal(provider))
		Expect(claims.ProviderToken).To(Equal(token))

		time.Sleep(time.Second)
		claims, err = Verify(SecretKey, jwtToken)
		Expect(err.Error()).To(MatchRegexp("invalid token: token is expired by.*"))
		Expect(claims).To(BeNil())

	})

})
