package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// GithubDeviceCodeResponse represents response body from the Github API
type GithubDeviceCodeResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	Interval        int    `json:"interval"`
}

// Uniquely identifies us as a GitHub app.
// This does not need to be obfuscated because it is publicly available
// to anyone who does an OAuth request via wego.
// See the auth ADR for more details:
// https://github.com/weaveworks/weave-gitops/blob/main/doc/adr/0005-wego-core-auth-strategy.md#design
const WeGOGithubClientID = "edcb13588d46f254052c"

//counterfeiter:generate . GithubAuthClient
type GithubAuthClient interface {
	GetDeviceCode() (*GithubDeviceCodeResponse, error)
	GetDeviceCodeAuthStatus(deviceCode string) (string, error)
	ValidateToken(ctx context.Context, token string) error
}

type ghAuth struct {
	http *http.Client
}

func NewGithubAuthClient(client *http.Client) GithubAuthClient {
	return ghAuth{http: client}
}

func (g ghAuth) GetDeviceCode() (*GithubDeviceCodeResponse, error) {
	return doGithubCodeRequest(g.http, GithubOAuthScope)
}

func (g ghAuth) GetDeviceCodeAuthStatus(deviceCode string) (string, error) {
	return doGithubDeviceAuthRequest(g.http, deviceCode)
}

func (g ghAuth) ValidateToken(ctx context.Context, token string) error {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", token))

	_, err = doRequest(req, g.http)

	return err
}

// Encapsulate shared logic between doCodeRequest and doAuthRequest
func doRequest(req *http.Request, client *http.Client) ([]byte, error) {
	req.Header.Set("Accept", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	rb, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		// err is falsey even on 4XX or 5XX
		return nil, ParseGitHubError(rb, res.StatusCode)
	}

	return rb, nil
}

const codeRequestURL = "https://github.com/login/device/code?%s"

// doGithubCodeRequest does the initial request of the Device Flow
func doGithubCodeRequest(client *http.Client, scope string) (*GithubDeviceCodeResponse, error) {
	query := url.Values.Encode(map[string][]string{
		"client_id": {WeGOGithubClientID},
		"scope":     {scope},
	})

	req, err := http.NewRequest("POST", fmt.Sprintf(codeRequestURL, query), nil)
	if err != nil {
		return nil, err
	}

	b, err := doRequest(req, client)
	if err != nil {
		return nil, fmt.Errorf("error doing code request: %w", err)
	}

	d := &GithubDeviceCodeResponse{}

	if err := json.Unmarshal(b, d); err != nil {
		return nil, fmt.Errorf("could not unmarshal code response: %w", err)
	}

	return d, nil
}

var ErrAuthPending = errors.New("auth pending")
var ErrSlowDown = errors.New("slow down")

const accessTokenURL = "https://github.com/login/oauth/access_token?%s"
const githubRequiredGrantType = "urn:ietf:params:oauth:grant-type:device_code"

// It appears we need `repo` scope, which is VERY permissive.
// We need to be able to push a deploy key and merge commits. No other scopes matched.
// Available scopes: https://docs.github.com/en/developers/apps/building-oauth-apps/scopes-for-oauth-apps
const GithubOAuthScope = "repo"

type githubAuthResponse struct {
	AccessToken string `json:"access_token"`
	Error       string `json:"error"`
}

// doGithubDeviceAuthRequest is used to poll for the status of the device flow.
func doGithubDeviceAuthRequest(client *http.Client, deviceCode string) (string, error) {
	query := url.Values.Encode(map[string][]string{
		"client_id":   {WeGOGithubClientID},
		"device_code": {deviceCode},
		"grant_type":  {githubRequiredGrantType},
	})
	url := fmt.Sprintf(accessTokenURL, query)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return "", fmt.Errorf("could not create auth request: %w", err)
	}

	b, err := doRequest(req, client)
	if err != nil {
		return "", fmt.Errorf("error doing auth request: %w", err)
	}

	p := githubAuthResponse{}

	if err := json.Unmarshal(b, &p); err != nil {
		return "", fmt.Errorf("err marshaling request body: %w", err)
	}

	if p.Error == "authorization_pending" {
		// This is expected until the user completes the auth flow.
		return "", ErrAuthPending
	}

	if p.Error == "slow_down" {
		return "", ErrSlowDown
	}

	if p.AccessToken != "" {
		return p.AccessToken, nil
	}

	// Note p.Error is a string here
	return "", fmt.Errorf("error doing auth request: %s", p.Error)
}

func ParseGitHubError(b []byte, statusCode int) error {
	var gerr GitHubError
	if err := json.Unmarshal(b, &gerr); err != nil {
		return fmt.Errorf("failed to unmarshal GitHub error: %w", err)
	}

	gerr.StatusCode = statusCode

	return gerr
}

// GitHubError indicates a failure response from GitHub.
type GitHubError struct {
	Type        string `json:"error"`
	Description string `json:"error_description"`
	URI         string `json:"error_uri"`
	StatusCode  int
}

func (e GitHubError) Error() string {
	return fmt.Sprintf("GitHub %d - %s (%q) more information at %s", e.StatusCode, e.Type, e.Description, e.URI)
}
