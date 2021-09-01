package auth

import (
	"time"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("JWT tokens", func() {

	It("Verify should fail after waiting longer than the expiration time", func() {

		token := "token"

		cli := NewJwtClient()

		jwtToken, err := cli.GenerateJWT(SecretKey, time.Millisecond, gitproviders.GitProviderGitHub, token)
		Expect(err).NotTo(HaveOccurred())

		claims, err := cli.VerifyJWT(SecretKey, jwtToken)
		Expect(err).NotTo(HaveOccurred())

		Expect(claims.Provider).To(Equal(gitproviders.GitProviderGitHub))
		Expect(claims.ProviderToken).To(Equal(token))

		time.Sleep(time.Second)
		claims, err = cli.VerifyJWT(SecretKey, jwtToken)
		Expect(err).To(Equal(ErrUnauthorizedToken))
		Expect(claims).To(BeNil())

	})

})
