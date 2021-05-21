package gitproviders

import (
	"fmt"
	"os"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
)

const GITHUB_DOMAIN string = "github.com"

func GithubProvider() (gitprovider.Client, error) {
	return githubProviderHandler.GithubProvider()
}

// GithubProvider shim
type GithubProviderHandler interface {
	GithubProvider() (gitprovider.Client, error)
}

type defaultGithubProviderHandler struct{}

var githubProviderHandler GithubProviderHandler = defaultGithubProviderHandler{}

func (h defaultGithubProviderHandler) GithubProvider() (gitprovider.Client, error) {
	token, found := os.LookupEnv("GITHUB_TOKEN")
	if !found {
		return nil, fmt.Errorf("GITHUB_TOKEN not set in environment")
	}
	return github.NewClient(
		github.WithOAuth2Token(token),
		github.WithDestructiveAPICalls(true),
	)
}

// func (h defaultGithubProviderHandler) GetClusterGithubProvider() ClusterGithubProvider {
//  return getGithubProvider()
// }

func WithGithubProviderHandler(handler GithubProviderHandler, fun func() error) error {
	originalHandler := githubProviderHandler
	githubProviderHandler = handler
	defer func() {
		githubProviderHandler = originalHandler
	}()
	return fun()
}
