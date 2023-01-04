package internal

import (
	"fmt"
	"net/url"
	"os"
	"strings"
)

const (
	gitlabScheme = "https"
	gitlabHost   = "gitlab.com"
	// Default values that can be used for OAuth with the `wego-dev` GitLab Application
	gitlabClientID       = "451df5d954a3ebb371bba5e6b7d1468ead1a0ee6d88b0791b001566b7bbc10cd"
	gitlabClientSecret   = "b402c7601b71904fffec85d3cc8aa7e953405680aa9b1fc4fb8603e9bb7e208a"
	GitlabVerifierMin    = 43
	GitlabVerifierMax    = 128
	GitlabRedirectURICLI = "http://127.0.0.1:9999/oauth/gitlab/callback"
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

func getEnvDefault(key, defaultValue string) string {
	val := os.Getenv(key)
	if val == "" {
		val = defaultValue
	}

	return val
}

func getGitlabHost() string {
	return getEnvDefault("GITLAB_HOSTNAME", gitlabHost)
}

func getGitlabClientID() string {
	return getEnvDefault("GITLAB_CLIENT_ID", gitlabClientID)
}

func getGitlabClientSecret() string {
	return getEnvDefault("GITLAB_CLIENT_SECRET", gitlabClientSecret)
}

// GitlabAuthorizeURL returns a URL that can be used for a Gitlab OAuth authorize request
func GitlabAuthorizeURL(redirectURI string, scopes []string, verifier CodeVerifier) (url.URL, error) {
	u := buildGitlabURL()
	u.Path = "/oauth/authorize"

	params := u.Query()
	params.Set("client_id", getGitlabClientID())
	params.Set("redirect_uri", redirectURI)
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

// GitlabTokenURL returns a URL that can be used for a Gitlab OAuth token request
func GitlabTokenURL(redirectURI, authorizationCode string, verifier CodeVerifier) url.URL {
	u := buildGitlabURL()
	u.Path = "/oauth/token"

	params := u.Query()
	params.Set("client_id", getGitlabClientID())
	params.Set("redirect_uri", redirectURI)
	params.Set("code", authorizationCode)
	params.Set("grant_type", "authorization_code")
	params.Set("code_verifier", verifier.RawValue())
	params.Set("client_secret", getGitlabClientSecret())
	u.RawQuery = params.Encode()

	return u
}

// GitlabUserURL returns the url to request data about the currently logged in user
func GitlabUserURL() url.URL {
	u := buildGitlabURL()
	u.Path = "/user"

	return u
}

func buildGitlabURL() url.URL {
	u := url.URL{}
	u.Scheme = gitlabScheme
	u.Host = getGitlabHost()

	return u
}
