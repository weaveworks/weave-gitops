package gitproviders

import (
	"fmt"
	"os"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
)

const GITHUB_DOMAIN string = "github.com"

func GithubProvider() (gitprovider.Client, error) {

	token, found := os.LookupEnv("GITHUB_TOKEN")
	if !found {
		return nil, fmt.Errorf("GITHUB_TOKEN not set in environment")
	}

	return github.NewClient(
		github.WithOAuth2Token(token),
		github.WithDestructiveAPICalls(true),
	)
}
