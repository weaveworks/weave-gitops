package internal

import (
	"fmt"

	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
)

type gitProviderClient struct {
	token string
}

func NewGitProviderClient(token string) gitproviders.Client {
	return &gitProviderClient{
		token: token,
	}
}

// GetProvider returns a GitProvider passing the auth token into the implementation
func (c *gitProviderClient) GetProvider(repoUrl gitproviders.RepoURL, getAccountType gitproviders.AccountTypeGetter) (gitproviders.GitProvider, error) {
	provider, err := gitproviders.New(gitproviders.Config{
		Provider: repoUrl.Provider(),
		Token:    c.token,
		Hostname: repoUrl.URL().Host,
	}, repoUrl.Owner(), getAccountType)
	if err != nil {
		return nil, fmt.Errorf("error creating git provider client: %w", err)
	}

	return provider, nil
}
