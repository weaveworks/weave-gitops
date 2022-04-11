package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/weaveworks/weave-gitops/pkg/services/auth/internal"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/types"
)

var gitlabScopes = []string{"api", "read_user", "profile"}

//counterfeiter:generate . GitlabAuthClient
type GitlabAuthClient interface {
	AuthURL(ctx context.Context, redirectUri string) (url.URL, error)
	ExchangeCode(ctx context.Context, redirectUri, code string) (*types.TokenResponseState, error)
	ValidateToken(ctx context.Context, token string) error
}

type glAuth struct {
	http     *http.Client
	verifier internal.CodeVerifier
}

func NewGitlabAuthClient(client *http.Client) GitlabAuthClient {
	cv, err := internal.NewCodeVerifier(internal.GitlabVerifierMin, internal.GitlabVerifierMax)
	if err != nil {
		panic(err)
	}

	return glAuth{
		http:     client,
		verifier: cv,
	}
}

func (g glAuth) AuthURL(ctx context.Context, redirectUri string) (url.URL, error) {
	return internal.GitlabAuthorizeUrl(redirectUri, gitlabScopes, g.verifier)
}

func (g glAuth) ExchangeCode(ctx context.Context, redirectUri, code string) (*types.TokenResponseState, error) {
	tUrl := internal.GitlabTokenUrl(redirectUri, code, g.verifier)

	return doCodeExchangeRequest(ctx, tUrl, g.http)
}

func (g glAuth) ValidateToken(ctx context.Context, token string) error {
	u := internal.GitlabUserUrl()

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
	if err != nil {
		return err
	}

	res, err := g.http.Do(req)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("invalid token: %s", res.Status)
	}

	return nil
}

func doCodeExchangeRequest(ctx context.Context, tUrl url.URL, c *http.Client) (*types.TokenResponseState, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tUrl.String(), strings.NewReader(""))
	if err != nil {
		return nil, fmt.Errorf("could not create gitlab code request: %w", err)
	}

	res, err := c.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error exchanging gitlab code: %w", err)
	}

	if res.StatusCode != http.StatusOK {
		errRes := struct {
			Error       string `json:"error"`
			Description string `json:"error_description"`
		}{}

		if err := json.NewDecoder(res.Body).Decode(&errRes); err != nil {
			return nil, fmt.Errorf("could not parse error response: %w", err)
		}

		return nil, fmt.Errorf("code=%v, error=%s, description=%s", res.StatusCode, errRes.Error, errRes.Description)
	}

	r, err := parseTokenResponseBody(res.Body)
	if err != nil {
		return nil, err
	}

	token := &types.TokenResponseState{}

	token.SetGitlabTokenResponse(r)

	return token, nil
}

func parseTokenResponseBody(body io.ReadCloser) (internal.GitlabTokenResponse, error) {
	defer func() {
		_ = body.Close()
	}()

	var tokenResponse internal.GitlabTokenResponse
	err := json.NewDecoder(body).Decode(&tokenResponse)

	if err != nil {
		return internal.GitlabTokenResponse{}, err
	}

	return tokenResponse, nil
}
