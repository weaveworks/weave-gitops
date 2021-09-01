package auth

import (
	"time"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JWT tokens", func() {

	It("Verify should success", func() {

		token := "token"

		jwtToken, err := GenerateJWT(SecretKey, time.Millisecond, gitproviders.GitProviderGitHub, token)
		Expect(err).NotTo(HaveOccurred())

		claims, err := VerifyJWT(SecretKey, jwtToken)
		Expect(err).NotTo(HaveOccurred())

		Expect(claims.Provider).To(Equal(gitproviders.GitProviderGitHub))
		Expect(claims.ProviderToken).To(Equal(token))

		time.Sleep(time.Second)
		claims, err = VerifyJWT(SecretKey, jwtToken)
		Expect(err).To(Equal(ErrUnauthorizedToken))
		Expect(claims).To(BeNil())

	})

})
