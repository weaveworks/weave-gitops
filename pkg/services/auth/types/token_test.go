package types

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/internal"
)

var _ = Describe("TokenResponseState", func() {

	var token *TokenResponseState
	var tokenResponse internal.GitlabTokenResponse

	accessToken := "kEq6PWZ8x37CNNmk"
	tokenType := "test-token-type"
	var seconds int64 = 600
	refreshToken := "H2q4xABSMT"
	var createdAt int64 = 32425434

	_ = BeforeEach(func() {
		token = &TokenResponseState{}
		tokenResponse = internal.GitlabTokenResponse{
			AccessToken:  accessToken,
			TokenType:    tokenType,
			ExpiresIn:    seconds,
			RefreshToken: refreshToken,
			CreatedAt:    createdAt,
		}
	})

	It("populates fields properly", func() {
		token.SetGitlabTokenResponse(tokenResponse)

		Expect(token.AccessToken).To(Equal(accessToken))
		Expect(token.TokenType).To(Equal(tokenType))
		Expect(token.ExpiresIn).To(Equal(time.Duration(seconds) * time.Second))
		Expect(token.RefreshToken).To(Equal(refreshToken))
		Expect(token.CreatedAt).To(Equal(createdAt))

	})
})
