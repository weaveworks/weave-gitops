package gitproviders

import (
	"fmt"
	"net/http"

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

	// RoundTripper allows for the internal http transport to customized
	RoundTripper http.RoundTripper
}

func buildGitProvider(config Config) (gitprovider.Client, string, error) {
	if config.Token == "" {
		return nil, "", fmt.Errorf("no git provider token present")
	}

	opts := []gitprovider.ClientOption{
		gitprovider.WithOAuth2Token(config.Token),
	}
	if config.RoundTripper != nil {
		opts = append(opts, gitprovider.WithPreChainTransportHook(func(in http.RoundTripper) (out http.RoundTripper) {
			return config.RoundTripper
		}))
	}

	switch config.Provider {
	case GitProviderGitHub:
		// Quirk of ggp, if using github.com or gitlab.com and you prepend
		// that with https:// you end up with https://https//github.com !!!
		hostname := github.DefaultDomain
		if config.Hostname != "" && config.Hostname != github.DefaultDomain {
			// Quirk of ggp, have to specify scheme with custom domain
			hostname = "https://" + config.Hostname
			opts = append(opts, gitprovider.WithDomain(hostname))
		}

		if client, err := github.NewClient(opts...); err != nil {
			return nil, "", err
		} else {
			return client, hostname, nil
		}
	case GitProviderGitLab:
		opts = append(opts, gitprovider.WithConditionalRequests(true))

		// Quirk, see above
		hostname := gitlab.DefaultDomain
		if config.Hostname != "" && config.Hostname != gitlab.DefaultDomain {
			// Quirk, see above
			hostname = "https://" + config.Hostname
			opts = append(opts, gitprovider.WithDomain(hostname))
		}

		if client, err := gitlab.NewClient(config.Token, tokenTypeOauth, opts...); err != nil {
			return nil, "", err
		} else {
			return client, hostname, nil
		}
	default:
		return nil, "", fmt.Errorf("unsupported Git provider '%s'", config.Provider)
	}
}
