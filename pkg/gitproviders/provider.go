package gitproviders

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/weaveworks/weave-gitops/pkg/utils"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitlab"
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
		if errors.Is(err, gitprovider.ErrNotFound) || strings.Contains(err.Error(), gitprovider.ErrGroupNotFound.Error()) {
			return AccountTypeUser, nil
		}

		return "", fmt.Errorf("could not get account type %s", err)
	}

	return AccountTypeOrg, nil
}

func isEmptyRepoError(err error) bool {
	return strings.Contains(err.Error(), "409 Git Repository is empty")
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
func detectGitProviderFromUrl(raw string) (GitProviderName, error) {
	if strings.HasPrefix(raw, "git@") {
		raw = "ssh://" + raw
		raw = strings.Replace(raw, ".com:", ".com/", 1)
	}

	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("could not parse git repo url %q", raw)
	}

	switch u.Hostname() {
	case github.DefaultDomain:
		return GitProviderGitHub, nil
	case gitlab.DefaultDomain:
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

func NewNormalizedRepoURL(uri string) (NormalizedRepoURL, error) {
	providerName, err := detectGitProviderFromUrl(uri)
	if err != nil {
		return NormalizedRepoURL{}, fmt.Errorf("could get provider name from URL %s: %w", uri, err)
	}

	normalized := normalizeRepoURLString(uri, providerName)

	u, err := url.Parse(normalized)
	if err != nil {
		return NormalizedRepoURL{}, fmt.Errorf("could not create normalized repo URL %s: %w", uri, err)
	}

	owner, err := getOwnerFromUrl(*u, providerName)
	if err != nil {
		return NormalizedRepoURL{}, fmt.Errorf("could get owner name from URL %s: %w", uri, err)
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
