package git_providers

import (
	"os"

	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/fluxcd/go-git-providers/gitprovider"
)

const GITLAB_DOMAIN string = "gitlab.com"

func GetGitlabProvider() (gitprovider.Client, error) {
	return gitlab.NewClient(os.Getenv("GITLAB_TOKEN"), "oauth2")
}
