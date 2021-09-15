package gitproviders

import (
	"fmt"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/fluxcd/go-git-providers/gitprovider"
)

// GitProviderName holds a Git provider definition.
type GitProviderName string

const (
	GitProviderGitHub GitProviderName = "github"
	GitProviderGitLab GitProviderName = "gitlab"
	tokenTypeOauth    string          = "oauth2"
)

// Config defines the configuration for connecting to a GitProvider.
type Config struct {
	// Provider defines the GitProvider.
	Provider GitProviderName

	// Hostname is the HTTP/S hostname of the Provider,
	// e.g. github.example.com.
	Hostname string

	// Token contains the token used to authenticate with the
	// Provider.
	Token string
}

func buildGitProvider(config Config) (gitprovider.Client, string, error) {
	if config.Token == "" {
		return nil, "", fmt.Errorf("no git provider token present")
	}

	switch config.Provider {
	case GitProviderGitHub:
		opts := []gitprovider.ClientOption{
			gitprovider.WithOAuth2Token(config.Token),
		}
		if config.Hostname != "" {
			opts = append(opts, gitprovider.WithDomain(config.Hostname))
		}

		if client, err := github.NewClient(opts...); err != nil {
			return nil, "", err
		} else {
			return client, github.DefaultDomain, nil
		}
	case GitProviderGitLab:
		opts := []gitprovider.ClientOption{
			gitprovider.WithConditionalRequests(true),
		}
		if config.Hostname != "" {
			opts = append(opts, gitprovider.WithDomain(config.Hostname))
		}

		if client, err := gitlab.NewClient(config.Token, tokenTypeOauth, opts...); err != nil {
			return nil, "", err
		} else {
			return client, gitlab.DefaultDomain, nil
		}
	default:
		return nil, "", fmt.Errorf("unsupported Git provider '%s'", config.Provider)
	}
}
