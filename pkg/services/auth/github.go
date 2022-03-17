package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
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

	rb, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		// err is falsey even on 4XX or 5XX
		return nil, fmt.Errorf("request failed with status code (%s): %v", rb, res.StatusCode)
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

const accessTokenUrl = "https://github.com/login/oauth/access_token?%s"
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
	url := fmt.Sprintf(accessTokenUrl, query)

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

const ghBackOffIncrement = (5 * time.Second)

func pollAuthStatus(sleep func(d time.Duration), interval time.Duration, client *http.Client, deviceCode string) (string, error) {
	retryInterval := interval

	for {
		sleep(retryInterval)

		authToken, err := doGithubDeviceAuthRequest(client, deviceCode)
		if err != nil {
			if err == ErrAuthPending {
				// This is expected while the user goes to the webpage.
				continue
			}

			if err == ErrSlowDown {
				// Github wants us to add an additional to our interval of 5 seconds if we hit a `slow_down`
				// https://docs.github.com/en/developers/apps/building-oauth-apps/authorizing-oauth-apps#error-codes-for-the-device-flow
				retryInterval = retryInterval + ghBackOffIncrement

				continue
			}

			return "", fmt.Errorf("error fetching auth status: %w", err)
		}

		return authToken, nil
	}
}

// NewGithubDeviceFlowHandler returns a function which will initiate the Github Device Flow for the CLI.
func NewGithubDeviceFlowHandler(client *http.Client) BlockingCLIAuthHandler {
	return func(ctx context.Context, w io.Writer) (string, error) {
		codeRes, err := doGithubCodeRequest(client, GithubOAuthScope)
		if err != nil {
			return "", fmt.Errorf("could not do code request: %w", err)
		}

		fmt.Fprintln(w)
		fmt.Fprintf(w, "Visit this URL to authenticate with Github:\n\n")
		fmt.Fprintf(w, "%s\n\n", codeRes.VerificationURI)
		fmt.Fprintf(w, "Type the following code into the page at the URL above: %s\n\n", codeRes.UserCode)
		fmt.Fprintf(w, "Waiting for authentication flow completion...\n\n")

		// GH complains if you retry RIGHT at the given interval.
		// We will get a `slow_down` error from the backend without the one second padding.
		retryInterval := time.Duration(codeRes.Interval) * time.Second

		return pollAuthStatus(time.Sleep, retryInterval, client, codeRes.DeviceCode)
	}
}
