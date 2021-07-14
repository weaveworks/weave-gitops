package gitproviders

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	gh_client "github.com/google/go-github/v32/github"
	"golang.org/x/oauth2"
)

var githubProvider gitprovider.Client

func GithubProvider() (gitprovider.Client, error) {
	return githubProviderHandler.GithubProvider()
}

func SetGithubProvider(githubProviderClient gitprovider.Client) {
	githubProviderHandler.SetGithubProvider(githubProviderClient)
}

// GithubProvider shim
type GithubProviderHandler interface {
	GithubProvider() (gitprovider.Client, error)
	SetGithubProvider(githubProviderClient gitprovider.Client)
}

type defaultGithubProviderHandler struct{}

var githubProviderHandler GithubProviderHandler = defaultGithubProviderHandler{}

func (h defaultGithubProviderHandler) SetGithubProvider(githubProviderClient gitprovider.Client) {
	githubProvider = githubProviderClient
}

func (h defaultGithubProviderHandler) GithubProvider() (gitprovider.Client, error) {
	var err error

	if githubProvider == nil {
		token, found := os.LookupEnv("GITHUB_TOKEN")
		if !found {
			return nil, fmt.Errorf("GITHUB_TOKEN not set in environment")
		}

		githubProvider, err = github.NewClient(
			github.WithOAuth2Token(token),
			github.WithDestructiveAPICalls(true),
		)
		if err != nil {
			return nil, fmt.Errorf("error getting github provider %s", err)
		}
	}

	return githubProvider, nil
}

func WithGithubProviderHandler(handler GithubProviderHandler, fun func() error) error {
	originalHandler := githubProviderHandler
	githubProviderHandler = handler
	defer func() {
		githubProviderHandler = originalHandler
	}()
	return fun()
}

func (h defaultGithubProviderHandler) OauthConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_OAUTH_CLIENT_SECRET"),
		Scopes:       []string{"SCOPE1", "SCOPE2"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}
}

func (gh defaultGithubProviderHandler) GetUser(ctx context.Context, token *oauth2.Token) (*User, error) {
	githubProvider, err := github.NewClient(
		github.WithOAuth2Token(token.AccessToken),
		github.WithDestructiveAPICalls(true),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting github provider %w", err)
	}
	rawGh := githubProvider.Raw()

	// Convert to a standard GH rest API client.
	// The flux go-git-providers client does not access individual user data.
	client, ok := rawGh.(*gh_client.Client)
	if !ok {
		return nil, errors.New("could not convert Raw providers client to GH client")
	}

	user, _, err := client.Users.Get(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("could not get github user: %w", err)
	}

	return &User{Email: *user.Email}, nil
}
