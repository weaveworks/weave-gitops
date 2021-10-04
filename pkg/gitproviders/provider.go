package gitproviders

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/utils"

	"github.com/fluxcd/go-git-providers/gitprovider"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 -generate

type ProviderAccountType string

const (
	AccountTypeUser ProviderAccountType = "user"
	AccountTypeOrg  ProviderAccountType = "organization"
	deployKeyName                       = "wego-deploy-key"

	defaultTimeout = time.Second * 30
)

// GitProvider Handler
//counterfeiter:generate . GitProvider
type GitProvider interface {
	RepositoryExists(name string, owner string) (bool, error)
	DeployKeyExists(owner, repoName string) (bool, error)
	GetDefaultBranch(url string) (string, error)
	GetRepoVisibility(url string) (*gitprovider.RepositoryVisibility, error)
	UploadDeployKey(owner, repoName string, deployKey []byte) error
	CreatePullRequest(owner string, repoName string, targetBranch string, newBranch string, files []gitprovider.CommitFile, commitMsg string, prTitle string, prDescription string) (gitprovider.PullRequest, error)
	GetCommits(owner string, repoName, targetBranch string, pageSize int, pageToken int) ([]gitprovider.Commit, error)
	GetProviderDomain() string
}

type defaultGitProvider struct {
	domain   string
	provider gitprovider.Client
}

type AccountTypeGetter func(provider gitprovider.Client, domain string, owner string) (ProviderAccountType, error)

func New(config Config, owner string, getAccountType AccountTypeGetter) (GitProvider, error) {
	provider, domain, err := buildGitProvider(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build git provider: %w", err)
	}

	accountType, err := getAccountType(provider, domain, owner)
	if err != nil {
		return nil, err
	}

	if accountType == AccountTypeOrg {
		return orgGitProvider{
			domain:   domain,
			provider: provider,
		}, nil
	}

	return userGitProvider{
		domain:   domain,
		provider: provider,
	}, nil
}

func GetAccountType(provider gitprovider.Client, domain string, owner string) (ProviderAccountType, error) {
	_, err := provider.Organizations().Get(context.Background(), gitprovider.OrganizationRef{
		Domain:       domain,
		Organization: owner,
	})
	if err != nil {
		if errors.Is(err, gitprovider.ErrNotFound) {
			return AccountTypeUser, nil
		}

		return "", fmt.Errorf("could not get account type %s", err)
	}

	return AccountTypeOrg, nil
}

func isEmptyRepoError(err error) bool {
	return strings.Contains(err.Error(), "409 Git Repository is empty")
}

func (p defaultGitProvider) GetProviderDomain() string {
	return string(GitProviderName(p.provider.ProviderID())) + ".com"
}

func NewRepositoryInfo(description string, visibility gitprovider.RepositoryVisibility) gitprovider.RepositoryInfo {
	return gitprovider.RepositoryInfo{
		Description: &description,
		Visibility:  &visibility,
	}
}

func NewOrgRepositoryRef(domain, org, repoName string) gitprovider.OrgRepositoryRef {
	return gitprovider.OrgRepositoryRef{
		RepositoryName: repoName,
		OrganizationRef: gitprovider.OrganizationRef{
			Domain:       domain,
			Organization: org,
		},
	}
}

func NewUserRepositoryRef(domain, user, repoName string) gitprovider.UserRepositoryRef {
	return gitprovider.UserRepositoryRef{
		RepositoryName: repoName,
		UserRef: gitprovider.UserRef{
			Domain:    domain,
			UserLogin: user,
		},
	}
}

// DetectGitProviderFromUrl accepts a url related to a git repo and
// returns the name of the provider associated.
// The raw URL is assumed to be something like ssh://git@github.com/myorg/myrepo.git.
// The common `git clone` variant of `git@github.com:myorg/myrepo.git` is not supported.
func DetectGitProviderFromUrl(raw string) (GitProviderName, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("could not parse git repo url %q", raw)
	}

	switch u.Hostname() {
	case "github.com":
		return GitProviderGitHub, nil
	case "gitlab.com":
		return GitProviderGitLab, nil
	}

	return "", fmt.Errorf("no git providers found for \"%s\"", raw)
}

type RepositoryURLProtocol string

const RepositoryURLProtocolHTTPS RepositoryURLProtocol = "https"
const RepositoryURLProtocolSSH RepositoryURLProtocol = "ssh"

type NormalizedRepoURL struct {
	repoName   string
	owner      string
	url        *url.URL
	normalized string
	provider   GitProviderName
	protocol   RepositoryURLProtocol
}

var sshPrefixRe = regexp.MustCompile(`git@(.*):(.*)/(.*)`)

func normalizeRepoURLString(url string) string {
	if !strings.HasSuffix(url, ".git") {
		url = url + ".git"
	}

	captured := sshPrefixRe.FindAllStringSubmatch(url, 1)

	if len(captured) > 0 {
		captured := sshPrefixRe.FindAllStringSubmatch(url, 1)
		matches := captured[0]

		if len(matches) >= 3 {
			provider := matches[1]
			org := matches[2]
			repo := matches[3]
			n := fmt.Sprintf("ssh://git@%s/%s/%s", provider, org, repo)

			return n
		}
	}

	return url
}

func NewNormalizedRepoURL(uri string) (NormalizedRepoURL, error) {
	normalized := normalizeRepoURLString(uri)

	u, err := url.Parse(normalized)
	if err != nil {
		return NormalizedRepoURL{}, fmt.Errorf("could not create normalized repo URL %s: %w", uri, err)
	}

	owner, err := utils.GetOwnerFromUrl(normalized)
	if err != nil {
		return NormalizedRepoURL{}, fmt.Errorf("could get owner name from URL %s: %w", uri, err)
	}

	providerName, err := DetectGitProviderFromUrl(normalized)
	if err != nil {
		return NormalizedRepoURL{}, fmt.Errorf("could get provider name from URL %s: %w", uri, err)
	}

	protocol := RepositoryURLProtocolSSH
	if u.Scheme == "https" {
		protocol = RepositoryURLProtocolHTTPS
	}

	return NormalizedRepoURL{
		repoName:   utils.UrlToRepoName(uri),
		owner:      owner,
		url:        u,
		normalized: normalized,
		provider:   providerName,
		protocol:   protocol,
	}, nil
}

func (n NormalizedRepoURL) String() string {
	return n.normalized
}

func (n NormalizedRepoURL) URL() *url.URL {
	return n.url
}

func (n NormalizedRepoURL) Owner() string {
	return n.owner
}

func (n NormalizedRepoURL) RepositoryName() string {
	return n.repoName
}

func (n NormalizedRepoURL) Provider() GitProviderName {
	return n.provider
}

func (n NormalizedRepoURL) Protocol() RepositoryURLProtocol {
	return n.protocol
}
