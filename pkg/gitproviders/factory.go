package gitproviders

import (
	"fmt"
	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/fluxcd/go-git-providers/stash"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
)

// GitProviderName holds a Git provider definition.
type GitProviderName string

const (
	GitProviderGitHub          GitProviderName = "github"
	GitProviderGitLab          GitProviderName = "gitlab"
	GitProviderBitBucketServer GitProviderName = "bitbucket-server"
	GitProviderAzureDevOps     GitProviderName = "azure-devops"
	tokenTypeOauth             string          = "oauth2"
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

	// Username contains the username needed for git operations
	// in the BitBucket Server git provider.
	Username string
}

func buildGitProvider(config Config) (gitprovider.Client, string, error) {
	if config.Token == "" {
		return nil, "", fmt.Errorf("no git provider token present")
	}

	logger := zapr.NewLogger(zap.L())
	switch config.Provider {
	case GitProviderGitHub:
		opts := []gitprovider.ClientOption{
			gitprovider.WithOAuth2Token(config.Token),
			gitprovider.WithLogger(&logger),
		}

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
		opts := []gitprovider.ClientOption{
			gitprovider.WithOAuth2Token(config.Token),
			gitprovider.WithConditionalRequests(true),
			gitprovider.WithLogger(&logger),
		}

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
	case GitProviderBitBucketServer:
		if config.Username == "" {
			return nil, "", fmt.Errorf("the BitBucket Server git provider requires a username to be set")
		}

		opts := []gitprovider.ClientOption{
			gitprovider.WithOAuth2Token(config.Token),
			gitprovider.WithConditionalRequests(true),
			gitprovider.WithLogger(&logger),
		}

		hostname := "https://" + config.Hostname
		opts = append(opts, gitprovider.WithDomain(hostname))

		if client, err := stash.NewStashClient(config.Username, config.Token, opts...); err != nil {
			return nil, "", fmt.Errorf("failed to create BitBucket Server client: %w", err)
		} else {
			return client, hostname, nil
		}
	default:
		return nil, "", fmt.Errorf("unsupported Git provider '%s'", config.Provider)
	}
}
