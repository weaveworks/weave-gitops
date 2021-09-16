package internal

import (
	"fmt"
	"net/url"
	"strings"
)

const (
	gitlabScheme       = "https"
	gitlabHost         = "gitlab.com"
	gitlabClientId     = "451df5d954a3ebb371bba5e6b7d1468ead1a0ee6d88b0791b001566b7bbc10cd"
	gitlabClientSecret = "b402c7601b71904fffec85d3cc8aa7e953405680aa9b1fc4fb8603e9bb7e208a"

	GitlabVerifierMin    = 43
	GitlabVerifierMax    = 128
	GitlabRedirectUriCLI = "http://127.0.0.1:9999/oauth/gitlab/callback"
	GitlabCallbackPath   = "/oauth/gitlab/callback"
	GitlabTempServerPort = ":9999"
)

// GitlabTokenResponse is the expected struct when going through OAuth authorization code grant process
type GitlabTokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	CreatedAt    int64  `json:"created_at"`
}

// GitlabAuthorizeUrl returns a URL that can be used for a Gitlab OAuth authorize request
func GitlabAuthorizeUrl(redirectUri string, scopes []string, verifier CodeVerifier) (url.URL, error) {
	u := url.URL{}
	u.Scheme = gitlabScheme
	u.Host = gitlabHost
	u.Path = "/oauth/authorize"

	params := u.Query()
	params.Set("client_id", gitlabClientId)
	params.Set("redirect_uri", redirectUri)
	params.Set("response_type", "code")

	codeChallenge, err := verifier.CodeChallenge()
	if err != nil {
		return url.URL{}, fmt.Errorf("gitlab authorize url generate code challenge: %w", err)
	}

	params.Set("code_challenge", codeChallenge)
	params.Set("code_challenge_method", "S256")
	params.Set("scope", strings.Join(scopes, " "))
	u.RawQuery = params.Encode()

	return u, nil
}

// GitlabTokenUrl returns a URL that can be used for a Gitlab OAuth token request
func GitlabTokenUrl(redirectUri, authorizationCode string, verifier CodeVerifier) url.URL {
	u := url.URL{}
	u.Scheme = gitlabScheme
	u.Host = gitlabHost
	u.Path = "/oauth/token"

	params := u.Query()
	params.Set("client_id", gitlabClientId)
	params.Set("redirect_uri", redirectUri)
	params.Set("code", authorizationCode)
	params.Set("grant_type", "authorization_code")
	params.Set("code_verifier", verifier.RawValue())
	params.Set("client_secret", gitlabClientSecret)
	u.RawQuery = params.Encode()

	return u
}
