package gitproviders

import (
	"os"

	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/fluxcd/go-git-providers/gitprovider"
)

func GetGitlabProvider() (gitprovider.Client, error) {
	return gitlab.NewClient(
		os.Getenv("GITLAB_TOKEN"),
		"oauth2",
		gitlab.WithDestructiveAPICalls(true),
	)
}
