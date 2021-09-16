package types

import "github.com/weaveworks/weave-gitops/pkg/services/auth/internal"

// TokenResponseState is used for passing state through HTTP middleware
type TokenResponseState struct {
	AccessToken    string
	TokenType      string
	ExpiresIn      int64
	RefreshToken   string
	CreatedAt      int64
	HttpStatusCode int
	Err            error
}

// SetGitlabTokenResponse will modify the TokenResponseState and populate the relevant fields from
// a GitlabTokenResponse
func (t *TokenResponseState) SetGitlabTokenResponse(token internal.GitlabTokenResponse) {
	t.AccessToken = token.AccessToken
	t.RefreshToken = token.RefreshToken
	t.ExpiresIn = token.ExpiresIn
	t.CreatedAt = token.CreatedAt
	t.TokenType = token.TokenType
}
