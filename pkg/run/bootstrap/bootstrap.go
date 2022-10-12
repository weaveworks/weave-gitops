package bootstrap

import (
	"context"

	"github.com/fluxcd/go-git-providers/github"
	"github.com/fluxcd/go-git-providers/gitlab"
	"github.com/fluxcd/go-git-providers/gitprovider"
	"github.com/fluxcd/go-git-providers/stash"
	"github.com/weaveworks/weave-gitops/pkg/fluxexec"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

type Bootstrap interface {
	RunBootstrapCmd(context.Context, *fluxexec.Flux) error
	SyncResources(context.Context, logger.Logger, []gitprovider.CommitFile) error
}

type BootstrapCommon struct {
	provider    GitProvider
	clusterPath string
	branch      string
}

// TODO: this needs to be implemented using plain go-git
type BootstrapRaw struct {
	BootstrapCommon
	url            string
	password       string
	privateKeyFile string
}

type BootstrapForge struct {
	BootstrapCommon
	isPersonal bool
	isPrivate  bool
	host       string
	owner      string
	repository string
	user       string
	pat        string
}

func NewBootstrap(clusterPath string, options BootstrapCmdOptions, provider GitProvider) Bootstrap {
	if provider == GitProviderGitHub || provider == GitProviderGitLab || provider == GitProviderBitbucketServer {
		return &BootstrapForge{
			BootstrapCommon: BootstrapCommon{
				provider:    provider,
				clusterPath: clusterPath,
				branch:      options[BranchOptionKey],
			},
			isPersonal: options[PersonalOptionKey] == "true",
			isPrivate:  options[PrivateOptionKey] == "true",
			host:       options[HostnameOptionKey],
			owner:      options[OwnerOptionKey],
			repository: options[RepositoryOptionKey],
			user:       options[UsernameOptionKey],
			pat:        options[PATOptionKey],
		}
	} else if provider == GitProviderGit {
		return &BootstrapRaw{
			BootstrapCommon: BootstrapCommon{
				provider:    provider,
				clusterPath: clusterPath,
				branch:      options[BranchOptionKey],
			},
			url:            options[URLOptionKey],
			password:       options[PasswordOptionKey],
			privateKeyFile: options[PrivateKeyFileOptionKey],
		}
	} else {
		// TODO put additional manifests on disk
		return nil
	}
}

func (b *BootstrapForge) RunBootstrapCmd(ctx context.Context, flux *fluxexec.Flux) error {
	globalOptions := fluxexec.WithBootstrapOptions(
		fluxexec.Branch(b.branch),
	)

	switch b.provider {
	case GitProviderGitHub:
		flux.SetEnvVar("GITHUB_TOKEN", b.pat)

		return flux.BootstrapGitHub(ctx,
			fluxexec.Hostname(b.host),
			fluxexec.Owner(b.owner),
			fluxexec.Repository(b.repository),
			fluxexec.Path(b.clusterPath),
			fluxexec.Personal(b.isPersonal),
			fluxexec.Private(b.isPrivate),
			fluxexec.WithBootstrapOptions(fluxexec.Branch(b.branch)),
			globalOptions,
		)
	case GitProviderGitLab:
		flux.SetEnvVar("GITLAB_TOKEN", b.pat)

		return flux.BootstrapGitlab(ctx,
			fluxexec.Hostname(b.host),
			fluxexec.Owner(b.owner),
			fluxexec.Repository(b.repository),
			fluxexec.Path(b.clusterPath),
			fluxexec.Personal(b.isPersonal),
			fluxexec.Private(b.isPrivate),
			fluxexec.WithBootstrapOptions(fluxexec.Branch(b.branch)),
			globalOptions,
		)
	case GitProviderBitbucketServer:
		flux.SetEnvVar("BITBUCKET_TOKEN", b.pat)

		return flux.BootstrapBitbucketServer(ctx,
			fluxexec.Hostname(b.host),
			fluxexec.Owner(b.owner),
			fluxexec.Repository(b.repository),
			fluxexec.Path(b.clusterPath),
			fluxexec.Personal(b.isPersonal),
			fluxexec.Private(b.isPrivate),
			fluxexec.WithBootstrapOptions(fluxexec.Branch(b.branch)),
			globalOptions,
		)
	}
	// unreachable
	return nil
}

func (b *BootstrapForge) SyncResources(ctx context.Context, log logger.Logger, commitFiles []gitprovider.CommitFile) error {
	var (
		commits gitprovider.CommitClient
		client  gitprovider.Client
		err     error
	)

	switch b.provider {
	case GitProviderGitHub:
		client, err = github.NewClient(
			gitprovider.WithDomain(b.host),
			gitprovider.WithOAuth2Token(b.pat),
		)
	case GitProviderGitLab:
		client, err = gitlab.NewClient(b.pat, "",
			gitprovider.WithDomain(b.host),
			gitprovider.WithConditionalRequests(true),
		)
	case GitProviderBitbucketServer:
		client, err = stash.NewStashClient(b.user, b.pat, gitprovider.WithDomain(b.host))
	}

	if err != nil || client == nil {
		return err
	}

	var repoURL string

	if b.isPersonal {
		ref := gitprovider.UserRepositoryRef{
			UserRef: gitprovider.UserRef{
				Domain:    b.host,
				UserLogin: b.owner,
			},
			RepositoryName: b.repository,
		}

		repo, err := client.UserRepositories().Get(ctx, ref)
		if err != nil {
			return err
		}

		commits = repo.Commits()
		repoURL = ref.String()
	} else {
		ref := gitprovider.OrgRepositoryRef{
			OrganizationRef: gitprovider.OrganizationRef{
				Domain:       b.host,
				Organization: b.owner,
				// TODO: support suborganizations
			},
			RepositoryName: b.repository,
		}
		repo, err := client.OrgRepositories().Get(ctx, ref)
		if err != nil {
			return err
		}
		commits = repo.Commits()
		repoURL = ref.String()
	}

	_, err = commits.Create(ctx, b.branch, "[gitops run] Additional manifests", commitFiles)
	if err != nil {
		return err
	}

	log.Successf("Your automations have been synced to %v", repoURL)

	return nil
}

func (b *BootstrapRaw) RunBootstrapCmd(ctx context.Context, flux *fluxexec.Flux) error {
	globalOptions := fluxexec.WithBootstrapOptions(
		fluxexec.Branch(b.branch),
		fluxexec.PrivateKeyFile(b.privateKeyFile),
	)

	return flux.BootstrapGit(ctx,
		fluxexec.URL(b.url),
		fluxexec.Password(b.password),
		fluxexec.Path(b.clusterPath),
		globalOptions,
	)
}

func (b *BootstrapRaw) SyncResources(ctx context.Context, log logger.Logger, commitFiles []gitprovider.CommitFile) error {
	// TODO this isn't implemented
	return nil
}
