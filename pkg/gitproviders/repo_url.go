package gitproviders

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/weaveworks/weave-gitops/pkg/utils"
)

type RepositoryURLProtocol string

const RepositoryURLProtocolHTTPS RepositoryURLProtocol = "https"
const RepositoryURLProtocolSSH RepositoryURLProtocol = "ssh"

type RepoURL struct {
	repoName   string
	owner      string
	url        *url.URL
	normalized string
	provider   GitProviderName
	protocol   RepositoryURLProtocol
}

func NewRepoURL(uri string) (RepoURL, error) {
	providerName, err := detectGitProviderFromUrl(uri)
	if err != nil {
		return RepoURL{}, fmt.Errorf("could not get provider name from URL %s: %w", uri, err)
	}

	normalized := normalizeRepoURLString(uri, providerName)

	u, err := url.Parse(normalized)
	if err != nil {
		return RepoURL{}, fmt.Errorf("could not create normalized repo URL %s: %w", uri, err)
	}

	owner, err := getOwnerFromUrl(*u, providerName)
	if err != nil {
		return RepoURL{}, fmt.Errorf("could not get owner name from URL %s: %w", uri, err)
	}

	protocol := RepositoryURLProtocolSSH
	if u.Scheme == "https" {
		protocol = RepositoryURLProtocolHTTPS
	}

	return RepoURL{
		repoName:   utils.UrlToRepoName(uri),
		owner:      owner,
		url:        u,
		normalized: normalized,
		provider:   providerName,
		protocol:   protocol,
	}, nil
}

func (n RepoURL) String() string {
	return n.normalized
}

func (n RepoURL) URL() *url.URL {
	return n.url
}

func (n RepoURL) Owner() string {
	return n.owner
}

func (n RepoURL) RepositoryName() string {
	return n.repoName
}

func (n RepoURL) Provider() GitProviderName {
	return n.provider
}

func (n RepoURL) Protocol() RepositoryURLProtocol {
	return n.protocol
}

func getOwnerFromUrl(url url.URL, providerName GitProviderName) (string, error) {
	url.Path = strings.TrimPrefix(url.Path, "/")

	parts := strings.Split(url.Path, "/")
	if len(parts) < 2 {
		return "", fmt.Errorf("could not get owner from url %v", url.String())
	}

	if providerName == GitProviderGitLab {
		if len(parts) > 3 {
			return "", fmt.Errorf("a subgroup in a subgroup is not currently supported")
		}

		if len(parts) > 2 {
			return parts[0] + "/" + parts[1], nil
		}
	}

	return parts[0], nil
}

// detectGitProviderFromUrl accepts a url related to a git repo and
// returns the name of the provider associated.
// The raw URL is assumed to be something like ssh://git@github.com/myorg/myrepo.git.
// The common `git clone` variant of `git@github.com:myorg/myrepo.git` is not supported.
func detectGitProviderFromUrl(raw string) (GitProviderName, error) {
	if strings.HasPrefix(raw, "git@") {
		raw = "ssh://" + raw
		raw = strings.Replace(raw, ".com:", ".com/", 1)
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("could not parse git repo url %q: %w", raw, err)
	}

	switch u.Hostname() {
	case github.DefaultDomain:
		return GitProviderGitHub, nil
	case gitlab.DefaultDomain:
		return GitProviderGitLab, nil
	}

	return "", fmt.Errorf("no git providers found for \"%s\"", raw)
}

// normalizeRepoURLString accepts a url like git@github.com:someuser/podinfo.git and converts it into
// a string like ssh://git@github.com/someuser/podinfo.git. This helps standardize the different
// user inputs that might be provided.
func normalizeRepoURLString(url string, providerName GitProviderName) string {
	trimmed := ""

	if !strings.HasSuffix(url, ".git") {
		url = url + ".git"
	}

	sshPrefix := fmt.Sprintf("git@%s.com:", providerName)
	httpsPrefix := fmt.Sprintf("https://%s.com/", providerName)

	if strings.HasPrefix(url, sshPrefix) {
		trimmed = strings.TrimPrefix(url, sshPrefix)
	} else if strings.HasPrefix(url, httpsPrefix) {
		trimmed = strings.TrimPrefix(url, httpsPrefix)
	}

	if trimmed != "" {
		return fmt.Sprintf("ssh://git@%s.com/%s", providerName, trimmed)
	}

	return url
}
