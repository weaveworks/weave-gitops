package bootstrap

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	gogit "github.com/go-git/go-git/v5"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

type BootstrapWizardTask struct {
	flagName         string
	flagValue        string
	defaultFlagValue DefaultValueGetter
	flagDescription  string
	isBoolean        bool
	isPassword       bool
}

type BootstrapCmdOptions map[string]string

type BootstrapWizardCmd struct {
	Provider GitProvider
	Options  BootstrapCmdOptions
}

type BootstrapWizard struct {
	gitProvider GitProvider
	tasks       []*BootstrapWizardTask
	cmdOptions  BootstrapCmdOptions
}

const (
	BranchOptionKey         = "branch"
	HostnameOptionKey       = "hostname"
	OwnerOptionKey          = "owner"
	PasswordOptionKey       = "password"
	PersonalOptionKey       = "personal"
	PrivateKeyFileOptionKey = "private-key-file"
	PrivateOptionKey        = "private"
	RepositoryOptionKey     = "repository"
	PATOptionKey            = "pat"
	URLOptionKey            = "url"
	UsernameOptionKey       = "username"
)

type DefaultValueGetter func(*gogit.Repository) string

func GetHost(repo *gogit.Repository) string {
	return hostGetter(repo)
}

func constantDefault(value string) DefaultValueGetter {
	return func(_ *gogit.Repository) string {
		return value
	}
}

func ownerGetter(repo *gogit.Repository) string {
	remoteURL := ParseRemoteURL(repo)
	if remoteURL == "" {
		return ""
	}

	urlParts := GetURLParts(remoteURL)

	return urlParts.owner
}

func hostGetter(repo *gogit.Repository) string {
	remoteURL := ParseRemoteURL(repo)
	if remoteURL == "" {
		return ""
	}

	urlParts := GetURLParts(remoteURL)

	return urlParts.host
}

func repositoryGetter(repo *gogit.Repository) string {
	remoteURL := ParseRemoteURL(repo)
	if remoteURL == "" {
		return ""
	}

	urlParts := GetURLParts(remoteURL)

	return urlParts.repository
}

func fallbackGetter(getters ...DefaultValueGetter) DefaultValueGetter {
	return func(repo *gogit.Repository) string {
		for _, getter := range getters {
			resp := getter(repo)
			if resp != "" {
				return resp
			}
		}

		return ""
	}
}

func branchGetter(repo *gogit.Repository) string {
	head, err := repo.Head()
	if err != nil {
		return ""
	}

	branch := head.Name().String()

	if !strings.Contains(branch, branchPrefix) {
		return ""
	}

	return strings.Replace(branch, branchPrefix, "", 1)
}

func envGetter(key string) DefaultValueGetter {
	return func(_ *gogit.Repository) string {
		return os.Getenv(key)
	}
}

var boostrapGitHubTasks = []*BootstrapWizardTask{
	{
		flagName:         OwnerOptionKey,
		flagValue:        "",
		defaultFlagValue: ownerGetter,
		flagDescription:  "GitHub user or organization name",
	},
	{
		flagName:         RepositoryOptionKey,
		flagValue:        "",
		defaultFlagValue: repositoryGetter,
		flagDescription:  "GitHub repository name",
	},
	{
		flagName:         BranchOptionKey,
		flagValue:        "",
		defaultFlagValue: fallbackGetter(branchGetter, constantDefault("main")),
		flagDescription:  "Git branch",
	},
	{
		flagName:         PersonalOptionKey,
		flagValue:        "",
		defaultFlagValue: constantDefault("true"),
		flagDescription:  "if true, the owner is assumed to be a GitHub user; otherwise an org",
		isBoolean:        true,
	},
	{
		flagName:         PrivateOptionKey,
		flagValue:        "",
		defaultFlagValue: constantDefault("true"),
		flagDescription:  "If true, the repository is setup or configured as private",
		isBoolean:        true,
	},
	{
		flagName:         HostnameOptionKey,
		flagValue:        "",
		defaultFlagValue: hostGetter,
		flagDescription:  "GitHub hostname",
	},
	{
		flagName:         PATOptionKey,
		flagValue:        "",
		defaultFlagValue: envGetter("GITHUB_TOKEN"),
		flagDescription:  "GitHub Personal Access Token",
		isPassword:       true,
	},
}

var boostrapGitLabTasks = []*BootstrapWizardTask{
	{
		flagName:         OwnerOptionKey,
		flagValue:        "",
		defaultFlagValue: ownerGetter,
		flagDescription:  "GitLab user or group name",
	},
	{
		flagName:         RepositoryOptionKey,
		flagValue:        "",
		defaultFlagValue: repositoryGetter,
		flagDescription:  "GitLab repository name",
	},
	{
		flagName:         BranchOptionKey,
		flagValue:        "",
		defaultFlagValue: branchGetter,
		flagDescription:  "Git branch (default \"main\")",
	},
	{
		flagName:         PersonalOptionKey,
		flagValue:        "",
		defaultFlagValue: constantDefault("true"),
		flagDescription:  "if true, the owner is assumed to be a GitLab user; otherwise a group",
		isBoolean:        true,
	},
	{
		flagName:         PrivateOptionKey,
		flagValue:        "",
		defaultFlagValue: constantDefault("true"),
		flagDescription:  "if true, the repository is setup or configured as private (default true)",
		isBoolean:        true,
	},
	{
		flagName:         HostnameOptionKey,
		flagValue:        "",
		defaultFlagValue: hostGetter,
		flagDescription:  "GitLab hostname (default \"gitlab.com\")",
	},
	{
		flagName:         PATOptionKey,
		flagValue:        "",
		defaultFlagValue: envGetter("GITLAB_TOKEN"),
		flagDescription:  "GitLab Personal Access Token",
		isPassword:       true,
	},
}

var boostrapGitTasks = []*BootstrapWizardTask{
	{
		flagName:        URLOptionKey,
		flagValue:       "",
		flagDescription: "Git repository URL",
	},
	{
		flagName:        PasswordOptionKey,
		flagValue:       "",
		flagDescription: "basic authentication password",
		isPassword:      true,
	},
	{
		flagName:        PrivateKeyFileOptionKey,
		flagValue:       "",
		flagDescription: "path to a private key file used for authenticating to the Git SSH server",
	},
}

var boostrapBitbucketServerTasks = []*BootstrapWizardTask{
	{
		flagName:         OwnerOptionKey,
		flagValue:        "",
		defaultFlagValue: ownerGetter,
		flagDescription:  "Bitbucket Server user or project name",
	},
	{
		flagName:         UsernameOptionKey,
		flagValue:        "",
		defaultFlagValue: constantDefault("git"),
		flagDescription:  "authentication username",
	},
	{
		flagName:         RepositoryOptionKey,
		flagValue:        "",
		defaultFlagValue: repositoryGetter,
		flagDescription:  "Bitbucket Server repository name",
	},
	{
		flagName:         HostnameOptionKey,
		flagValue:        "",
		defaultFlagValue: hostGetter,
		flagDescription:  "Bitbucket Server hostname",
	},
	{
		flagName:         BranchOptionKey,
		flagValue:        "",
		defaultFlagValue: branchGetter,
		flagDescription:  "Git branch",
	},
	{
		flagName:         PersonalOptionKey,
		flagValue:        "",
		defaultFlagValue: constantDefault("true"),
		flagDescription:  "if true, the owner is assumed to be a Bitbucket Server user; otherwise a group",
		isBoolean:        true,
	},
	{
		flagName:         PrivateOptionKey,
		flagValue:        "",
		defaultFlagValue: constantDefault("true"),
		flagDescription:  "if true, the repository is setup or configured as private",
		isBoolean:        true,
	},
	{
		flagName:         PATOptionKey,
		flagValue:        "",
		defaultFlagValue: envGetter("BITBUCKET_TOKEN"),
		flagDescription:  "BitBucket Personal Access Token",
		isPassword:       true,
	},
}

const (
	providerGitHub = "github.com"
	providerGitLab = "gitlab.com"
	branchPrefix   = "refs/heads/"
)

type GitProvider int32

const (
	GitProviderUnknown         GitProvider = 0
	GitProviderGitHub          GitProvider = 1
	GitProviderGitLab          GitProvider = 2
	GitProviderGit             GitProvider = 3
	GitProviderBitbucketServer GitProvider = 4
)

const (
	gitProviderGitHubName          = "GitHub"
	gitProviderGitLabName          = "GitLab"
	gitProviderGitName             = "Git"
	gitProviderBitbucketServerName = "BitbucketServer"
)

var allGitProviderNames = []string{
	gitProviderGitHubName,
	gitProviderGitLabName,
	gitProviderGitName,
	// gitProviderBitbucketServerName,
}

var allGitProviders = map[string]GitProvider{
	gitProviderGitHubName:          GitProviderGitHub,
	gitProviderGitLabName:          GitProviderGitLab,
	gitProviderGitName:             GitProviderGit,
	gitProviderBitbucketServerName: GitProviderBitbucketServer,
}

var gitProvidersWithTasks = map[GitProvider][]*BootstrapWizardTask{
	GitProviderGitHub:          boostrapGitHubTasks,
	GitProviderGitLab:          boostrapGitLabTasks,
	GitProviderGit:             boostrapGitTasks,
	GitProviderBitbucketServer: boostrapBitbucketServerTasks,
}

// ParseRemoteURL extracts remote URL from the repository
func ParseRemoteURL(repo *gogit.Repository) string {
	remoteURLs := make(map[string]string)

	if repo != nil {
		remotes, _ := repo.Remotes()

		for _, remote := range remotes {
			config := remote.Config()
			remoteURLs[config.Name] = config.URLs[0]
		}
	}

	// Return origin first - that's usually the one git cloned from
	if url, ok := remoteURLs["origin"]; ok {
		return url
	}

	// No origin, return "any" - hopefully there's just one
	for _, url := range remoteURLs {
		return url
	}

	// No remotes - return nothing
	return ""
}

// ParseGitProvider extracts git provider from the remote URL, if possible
func ParseGitProvider(hostname string) GitProvider {
	provider := GitProviderUnknown

	if hostname == "" {
		return provider
	}

	if hostname == providerGitHub {
		return GitProviderGitHub
	}

	if hostname == providerGitLab {
		return GitProviderGitLab
	}

	return provider
}

type parts struct {
	host       string
	owner      string
	repository string
}

// GetURLParts splits URL to URL parts.
// This takes inspiration from
// https://github.com/fluxcd/go-git-providers/blob/cda93bf5a5fa65bd994a60d6d3ef9ad119cfb684/gitprovider/repositoryref.go#L309
// except it has to do it in reverse
func GetURLParts(remoteURL string) parts {
	sanitizedURL := remoteURL
	if strings.HasPrefix(sanitizedURL, "git@") {
		sanitizedURL = strings.Replace(sanitizedURL, ":", "/", 1)
	}

	replacer := strings.NewReplacer("git@", "", "https://", "", ".git", "", "ssh://", "")

	sanitizedURL = replacer.Replace(sanitizedURL)

	urlParts := strings.Split(sanitizedURL, "/")

	return parts{
		host:       urlParts[0],
		repository: urlParts[len(urlParts)-1],
		owner:      strings.Join(urlParts[1:len(urlParts)-1], "/"),
	}
}

// NewBootstrapWizard creates a wizard to gather
// all bootstrap config options before running flux bootstrap.
func NewBootstrapWizard(log logger.Logger, gitProvider GitProvider, repo *gogit.Repository) (*BootstrapWizard, error) {
	if gitProvider == GitProviderUnknown {
		return nil, fmt.Errorf("unknown git provider: %d", gitProvider)
	}

	wizard := &BootstrapWizard{
		gitProvider: gitProvider,
		tasks:       []*BootstrapWizardTask{},
	}

	tasks := gitProvidersWithTasks[gitProvider]

	wizard.tasks = make([]*BootstrapWizardTask, len(tasks))
	copy(wizard.tasks, tasks)

	log.Actionf("Parsing values ...")

	if repo == nil {
		return wizard, nil
	}

	for _, task := range wizard.tasks {
		if task.flagValue == "" && task.defaultFlagValue != nil {
			task.flagValue = task.defaultFlagValue(repo)
			continue
		}
	}

	return wizard, nil
}

// ParseGitRemote parses the git remote (if it exists)
// from the working directory to autofill some command options.
func ParseGitRemote(log logger.Logger, workingDir string) (*gogit.Repository, error) {
	log.Actionf("Collecting information about Git remote ...")

	if workingDir == "" {
		return nil, fmt.Errorf("unable to parse Git remote for empty workingDir")
	}

	if _, err := os.Stat(workingDir); err != nil {
		return nil, fmt.Errorf("error validating workingDir %s: %w", workingDir, err)
	}

	repo, err := gogit.PlainOpen(workingDir)
	if err != nil {
		return nil, fmt.Errorf("error parsing Git remote for workingDir %s: %w", workingDir, err)
	}

	return repo, nil
}

// SelectGitProvider displays text inputs to enter or edit all command flag values.
func SelectGitProvider(log logger.Logger) (GitProvider, error) {
	provider := GitProviderUnknown

	m := initialPreWizardModel(make(chan GitProvider))

	err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()).Start()
	if err != nil {
		return provider, fmt.Errorf("could not start tea program: %v", err.Error())
	}

	provider = <-m.msgChan

	return provider, nil
}

// Run displays text inputs to enter or edit all command flag values.
func (wizard *BootstrapWizard) Run(log logger.Logger) error {
	log.Actionf("Please enter or edit command values...")

	m := initialWizardModel(wizard.tasks, make(chan BootstrapCmdOptions))

	err := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion()).Start()
	if err != nil {
		return fmt.Errorf("could not start tea program: %v", err.Error())
	}

	wizard.cmdOptions = <-m.msgChan

	return nil
}

// BuildCmd builds flux bootstrap command options as key/values pairs.
func (wizard *BootstrapWizard) BuildCmd(log logger.Logger) BootstrapWizardCmd {
	return BootstrapWizardCmd{
		Provider: wizard.gitProvider,
		Options:  wizard.cmdOptions,
	}
}
