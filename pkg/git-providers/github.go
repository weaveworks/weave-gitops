package git_providers

import (
	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitprovider"
	"os"
)

const GITHUB_DOMAIN string = "github.com"

func GithubProvider() (gitprovider.Client,error) {
	return github.NewClient(github.WithOAuth2Token(os.Getenv("GITHUB_TOKEN")))
}
