package types

import (
	"time"

	"github.com/weaveworks/weave-gitops/pkg/services/auth/internal"
)

// TokenResponseState is used for passing state through HTTP middleware
type TokenResponseState struct {
	AccessToken    string
	TokenType      string
	ExpiresIn      time.Duration
	RefreshToken   string
	CreatedAt      int64
	HTTPStatusCode int
	Err            error
}

// SetGitlabTokenResponse will modify the TokenResponseState and populate the relevant fields from
// a GitlabTokenResponse
func (t *TokenResponseState) SetGitlabTokenResponse(token internal.GitlabTokenResponse) {
	t.AccessToken = token.AccessToken
	t.RefreshToken = token.RefreshToken
	t.ExpiresIn = time.Duration(token.ExpiresIn) * time.Second
	t.CreatedAt = token.CreatedAt
	t.TokenType = token.TokenType
}
