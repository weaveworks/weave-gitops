package auth

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/services/auth/internal"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/types"
)

const (
	serverShutdownSuccessMessage = "The authorization flow was completed, you may close your browser and go back to the CLI.\n"
	serverShutdownErrorMessage   = "\nThere was an issue with the authorization flow, you may close your browser and go back to the CLI for more information.\n"
)

var gitlabScopes = []string{"api", "read_user", "profile"}

type gitlabAuthFlow struct {
	client       *http.Client
	codeVerifier internal.CodeVerifier
	redirectUri  string
	scopes       []string
}

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

func NewGitlabAuthFlow(redirectUri string, client *http.Client) (types.AuthFlow, error) {
	cv, err := internal.NewCodeVerifier(internal.GitlabVerifierMin, internal.GitlabVerifierMax)
	if err != nil {
		return nil, err
	}

	return &gitlabAuthFlow{
		client:       client,
		codeVerifier: cv,
		redirectUri:  redirectUri,
		scopes:       gitlabScopes,
	}, nil
}

func (gaf *gitlabAuthFlow) Authorize(ctx context.Context) (*http.Request, error) {
	authUrl, err := internal.GitlabAuthorizeUrl(gaf.redirectUri, gaf.scopes, gaf.codeVerifier)
	if err != nil {
		return nil, fmt.Errorf("gitlab auth flow create authorize url: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, authUrl.String(), strings.NewReader(""))
	if err != nil {
		return nil, fmt.Errorf("gitlab authorize endpoint new request: %w", err)
	}

	return req, nil
}

func (gaf *gitlabAuthFlow) CallbackHandler(tokenState *types.TokenResponseState, next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer next.ServeHTTP(w, r)
		params := r.URL.Query()
		code := params.Get("code")
		tokenUrl := internal.GitlabTokenUrl(gaf.redirectUri, code, gaf.codeVerifier)

		ctx := context.Background()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenUrl.String(), strings.NewReader(""))
		req.Header.Set("Content-Type", "application/json")

		if err != nil {
			reqErr := fmt.Errorf("unable to generate request to fetch gitlab token, please try again later: %w", err)
			tokenState.HttpStatusCode = http.StatusInternalServerError
			tokenState.Err = reqErr

			w.WriteHeader(http.StatusInternalServerError)

			_, _ = w.Write([]byte(reqErr.Error()))

			return
		}

		res, err := gaf.client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
			tokenState.HttpStatusCode = http.StatusInternalServerError
			tokenState.Err = fmt.Errorf("gitlab token requeset client issue: %w", err)

			return
		}

		if res.StatusCode == http.StatusOK {
			payload, err := parseTokenResponseBody(res.Body)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_, _ = w.Write([]byte(http.StatusText(http.StatusInternalServerError)))
				tokenState.HttpStatusCode = http.StatusInternalServerError
				tokenState.Err = fmt.Errorf("gitlab token response json decode: %w", err)

				return
			}

			w.WriteHeader(http.StatusOK)
			tokenState.SetGitlabTokenResponse(payload)
			tokenState.HttpStatusCode = http.StatusOK
		} else {
			w.WriteHeader(res.StatusCode)
			_, _ = w.Write([]byte(http.StatusText(res.StatusCode)))
			tokenState.HttpStatusCode = res.StatusCode
			tokenState.Err = errors.New(http.StatusText(res.StatusCode))
		}
	}

	return http.HandlerFunc(fn)
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

// NewGitlabAuthFlowHandler returns a BlockingCLIAuthHandler for the Gitlab OAuth flow in a CLI.
// It will set up a temporary server as a callback mechanism.  Once the user runs through the flow
// the server will be shutdown, and we will exit the function.
func NewGitlabAuthFlowHandler(client *http.Client, flow types.AuthFlow) BlockingCLIAuthHandler {
	return func(ctx context.Context, w io.Writer) (string, error) {
		req, err := flow.Authorize(ctx)
		if err != nil {
			return "", fmt.Errorf("could not do code request: %w", err)
		}

		_, _ = fmt.Fprintf(w, "Starting authorization server:\n")
		_, _ = fmt.Fprintf(w, "Visit this URL to authenticate with Gitlab:\n\n")
		_, _ = fmt.Fprintf(w, "%s\n\n", req.URL)
		_, _ = fmt.Fprintf(w, "Waiting for authentication flow completion...\n\n")

		serverShutdown := &sync.WaitGroup{}
		serverShutdown.Add(1)

		token := types.TokenResponseState{}
		server := startAuthServerForCLI(serverShutdown, w, &token, flow)

		_, clientErr := client.Do(req)
		if clientErr != nil {
			serverShutdown.Done()

			token.Err = fmt.Errorf("gitlab auth client error: %w", clientErr)
		}

		serverShutdown.Wait()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		defer cancel()

		fmt.Fprintf(w, "Shutting the server down...\n")

		if err := server.Shutdown(shutdownCtx); err != nil {
			fmt.Fprintf(w, "An error occurred shutting down the server: %s\n", err.Error())
		}

		fmt.Fprintf(w, "Server shutdown complete!\n")

		if token.Err != nil {
			fmt.Fprintf(w, "There was an issue going through the Gitlab authentication flow:\n\n")

			if token.HttpStatusCode == 0 || token.HttpStatusCode == http.StatusOK {
				fmt.Fprintf(w, "%s", token.Err.Error())
			} else {
				fmt.Fprintf(w, "Gitlab returned status code %d with the error message: %s", token.HttpStatusCode, token.Err.Error())
			}

			return "", token.Err
		} else {
			return token.AccessToken, nil
		}
	}
}

func startAuthServerForCLI(wg *sync.WaitGroup, w io.Writer, token *types.TokenResponseState, flow types.AuthFlow) *http.Server {
	srv := &http.Server{Addr: internal.GitlabTempServerPort}

	http.Handle(internal.GitlabCallbackPath, flow.CallbackHandler(token, shutdownServerForCLI(token, wg)))

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Fprintf(w, "Error starting server: %s", err.Error())
		}
	}()

	return srv
}

func shutdownServerForCLI(token *types.TokenResponseState, wg *sync.WaitGroup) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()

		if token.HttpStatusCode == http.StatusOK {
			_, _ = w.Write([]byte(serverShutdownSuccessMessage))
		} else {
			_, _ = w.Write([]byte(serverShutdownErrorMessage))
		}
	}

	return http.HandlerFunc(fn)
}
