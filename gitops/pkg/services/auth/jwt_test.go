package auth

import (
	"time"

	"k8s.io/apimachinery/pkg/util/rand"

	"github.com/weaveworks/weave-gitops/gitops/pkg/gitproviders"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("JWT tokens", func() {

	It("Verify should fail after waiting longer than the expiration time", func() {
		rand.Seed(time.Now().UnixNano())
		secretKey := rand.String(20)
		token := "token"
		cli := NewJwtClient(secretKey)

		jwtToken, err := cli.GenerateJWT(time.Millisecond, gitproviders.GitProviderGitHub, token)
		Expect(err).NotTo(HaveOccurred())

		claims, err := cli.VerifyJWT(jwtToken)
		Expect(err).NotTo(HaveOccurred())

		Expect(claims.Provider).To(Equal(gitproviders.GitProviderGitHub))
		Expect(claims.ProviderToken).To(Equal(token))

		time.Sleep(time.Second)
		claims, err = cli.VerifyJWT(jwtToken)
		Expect(err).To(MatchError(ErrUnauthorizedToken))
		Expect(claims).To(BeNil())
	})
	It("works with a gitlab token", func() {
		rand.Seed(time.Now().UnixNano())
		secretKey := rand.String(20)
		token := "token"
		cli := NewJwtClient(secretKey)

		jwtToken, err := cli.GenerateJWT(time.Millisecond, gitproviders.GitProviderGitLab, token)
		Expect(err).NotTo(HaveOccurred())

		claims, err := cli.VerifyJWT(jwtToken)
		Expect(err).NotTo(HaveOccurred())

		Expect(claims.Provider).To(Equal(gitproviders.GitProviderGitLab))
		Expect(claims.ProviderToken).To(Equal(token))
	})

})
