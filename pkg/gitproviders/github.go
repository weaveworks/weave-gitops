package gitproviders

import (
	"fmt"
	"os"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
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
