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
	isRequired       bool
	isBoolean        bool
}

type BootstrapCmdOption struct {
	FlagName  string
	FlagValue string
}

type BootstrapWizardCmd struct {
	Provider GitProvider
	Options  []*BootstrapCmdOption
}

type BootstrapWizard struct {
	remoteURL   string
	gitProvider GitProvider
	tasks       []*BootstrapWizardTask
	cmdOptions  []*BootstrapCmdOption
}

const (
	BranchOptionKey         = "branch"
	HostnameOptionKey       = "hostname"
	OwnerOptionKey          = "owner"
	PasswordOptionKey       = "password"
	PathOptionKey           = "path"
	PersonalOptionKey       = "personal"
	PrivateKeyFileOptionKey = "private-key-file"
	PrivateOptionKey        = "private"
	RepositoryOptionKey     = "repository"
	SSHHostnameOptionKey    = "ssh-hostname"
	TeamOptionKey           = "team"
	TokenAuthOptionKey      = "token-auth"
	URLOptionKey            = "url"
	UsernameOptionKey       = "username"
)

type DefaultValueGetter func(*gogit.Repository) string

func constantDefault(value string) DefaultValueGetter {
	return func(_ *gogit.Repository) string {
		return value
	}
}

func ownerGetter(repo *gogit.Repository) string {
	remoteURL, err := ParseRemoteURL(repo)
	if err != nil || remoteURL == "" {
		return ""
	}

	urlParts := GetURLParts(remoteURL)

	numParts := len(urlParts)

	if numParts > 2 {
		return urlParts[numParts-2]
	}

	return ""
}

func repositoryGetter(repo *gogit.Repository) string {
	remoteURL, err := ParseRemoteURL(repo)
	if err != nil || remoteURL == "" {
		return ""
	}

	urlParts := GetURLParts(remoteURL)

	numParts := len(urlParts)

	if numParts > 2 {
		return urlParts[numParts-1]
	}

	return ""
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

var boostrapGitHubTasks = []*BootstrapWizardTask{
	{
		flagName:         OwnerOptionKey,
		flagValue:        "",
		defaultFlagValue: ownerGetter,
		flagDescription:  "GitHub user or organization name",
		isRequired:       true,
	},
	{
		flagName:         RepositoryOptionKey,
		flagValue:        "",
		defaultFlagValue: repositoryGetter,
		flagDescription:  "GitHub repository name",
		isRequired:       true,
	},
	{
		flagName:         BranchOptionKey,
		flagValue:        "",
		defaultFlagValue: branchGetter,
		flagDescription:  "Git branch (default \"main\")",
	},
	{
		flagName:        PathOptionKey,
		flagValue:       "",
		flagDescription: "path relative to the repository root, when specified the cluster sync will be scoped to this path",
	},
	{
		flagName:        PersonalOptionKey,
		flagValue:       "",
		flagDescription: "if true, the owner is assumed to be a GitHub user; otherwise an org",
		isRequired:      true,
		isBoolean:       true,
	},
	{
		flagName:         PrivateOptionKey,
		flagValue:        "",
		defaultFlagValue: constantDefault("true"),
		flagDescription:  "if true, the repository is setup or configured as private (default true)",
		isRequired:       true,
		isBoolean:        true,
	},
	{
		flagName:        TeamOptionKey,
		flagValue:       "",
		flagDescription: "GitHub team and the access to be given to it (team:maintain)",
	},
	{
		flagName:         HostnameOptionKey,
		flagValue:        "",
		defaultFlagValue: constantDefault("github.com"),
		flagDescription:  "GitHub hostname (default \"github.com\")",
	},
	{
		flagName:        SSHHostnameOptionKey,
		flagValue:       "",
		flagDescription: "SSH hostname, to be used when the SSH host differs from the HTTPS one",
	},
	{
		flagName:        TokenAuthOptionKey,
		flagValue:       "",
		flagDescription: "when enabled, the personal access token will be used instead of SSH deploy key",
		isBoolean:       true,
	},
}

var boostrapGitLabTasks = []*BootstrapWizardTask{
	{
		flagName:         OwnerOptionKey,
		flagValue:        "",
		defaultFlagValue: ownerGetter,
		flagDescription:  "GitLab user or group name",
		isRequired:       true,
	},
	{
		flagName:         RepositoryOptionKey,
		flagValue:        "",
		defaultFlagValue: repositoryGetter,
		flagDescription:  "GitLab repository name",
		isRequired:       true,
	},
	{
		flagName:         BranchOptionKey,
		flagValue:        "",
		defaultFlagValue: branchGetter,
		flagDescription:  "Git branch (default \"main\")",
	},
	{
		flagName:        PathOptionKey,
		flagValue:       "",
		flagDescription: "path relative to the repository root, when specified the cluster sync will be scoped to this path",
	},
	{
		flagName:        PersonalOptionKey,
		flagValue:       "",
		flagDescription: "if true, the owner is assumed to be a GitLab user; otherwise a group",
		isRequired:      true,
		isBoolean:       true,
	},
	{
		flagName:         PrivateOptionKey,
		flagValue:        "",
		defaultFlagValue: constantDefault("true"),
		flagDescription:  "if true, the repository is setup or configured as private (default true)",
		isRequired:       true,
		isBoolean:        true,
	},
	{
		flagName:        TeamOptionKey,
		flagValue:       "",
		flagDescription: "GitLab teams to be given maintainer access (also accepts comma-separated values)",
	},
	{
		flagName:        HostnameOptionKey,
		flagValue:       "",
		flagDescription: "GitLab hostname (default \"gitlab.com\")",
	},
	{
		flagName:         TokenAuthOptionKey,
		flagValue:        "",
		defaultFlagValue: constantDefault("false"),
		flagDescription:  "when enabled, the personal access token will be used instead of SSH deploy key",
		isRequired:       false,
		isBoolean:        true,
	},
}

var boostrapGitTasks = []*BootstrapWizardTask{
	{
		flagName:        URLOptionKey,
		flagValue:       "",
		flagDescription: "Git repository URL",
		isRequired:      true,
	},
	{
		flagName:        PasswordOptionKey,
		flagValue:       "",
		flagDescription: "basic authentication password",
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
		isRequired:       true,
	},
	{
		flagName:         UsernameOptionKey,
		flagValue:        "",
		defaultFlagValue: constantDefault("git"),
		flagDescription:  "authentication username (default \"git\")",
	},
	{
		flagName:         RepositoryOptionKey,
		flagValue:        "",
		defaultFlagValue: repositoryGetter,
		flagDescription:  "Bitbucket Server repository name",
		isRequired:       true,
	},
	{
		flagName:        HostnameOptionKey,
		flagValue:       "",
		flagDescription: "Bitbucket Server hostname",
		isRequired:      true,
	},
	{
		flagName:         BranchOptionKey,
		flagValue:        "",
		defaultFlagValue: branchGetter,
		flagDescription:  "Git branch (default \"main\")",
	},
	{
		flagName:        PathOptionKey,
		flagValue:       "",
		flagDescription: "path relative to the repository root, when specified the cluster sync will be scoped to this path",
	},
	{
		flagName:        PersonalOptionKey,
		flagValue:       "",
		flagDescription: "if true, the owner is assumed to be a Bitbucket Server user; otherwise a group",
		isRequired:      true,
		isBoolean:       true,
	},
	{
		flagName:         PrivateOptionKey,
		flagValue:        "",
		flagDescription:  "if true, the repository is setup or configured as private (default true)",
		defaultFlagValue: constantDefault("true"),
		isRequired:       true,
		isBoolean:        true,
	},
	{
		flagName:        TokenAuthOptionKey,
		flagValue:       "",
		flagDescription: "when enabled, the personal access token will be used instead of SSH deploy key",
		isBoolean:       true,
	},
}

const (
	providerGitHub = "github.com"
	providerGitLab = "gitlab.com"
	branchPrefix   = "refs/heads/"
)

type (
	GitProvider int32
	errMsg      error
)

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
	gitProviderBitbucketServerName,
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
func ParseRemoteURL(repo *gogit.Repository) (string, error) {
	remoteURLs := []string{}

	if repo != nil {
		remotes, _ := repo.Remotes()

		for _, remote := range remotes {
			remoteURLs = append(remoteURLs, remote.Config().URLs...)
		}
	}

	if len(remoteURLs) == 1 {
		return remoteURLs[0], nil
	}

	return "", fmt.Errorf("multiple remotes detected")
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

// GetURLParts splits URL to URL parts.
func GetURLParts(remoteURL string) []string {
	replacer := strings.NewReplacer("git@", "", "https://", "", ".git", "")

	sanitizedURL := replacer.Replace(remoteURL)
	sanitizedURL = strings.Replace(sanitizedURL, ":", "/", 1)

	urlParts := strings.Split(sanitizedURL, "/")

	return urlParts
}

// NewBootstrapWizard creates a wizard to gather
// all bootstrap config options before running flux bootstrap.
func NewBootstrapWizard(log logger.Logger, remoteURL string, gitProvider GitProvider, repo *gogit.Repository, path string) (*BootstrapWizard, error) {
	if gitProvider == GitProviderUnknown {
		return nil, fmt.Errorf("unknown git provider: %d", gitProvider)
	}

	wizard := &BootstrapWizard{
		remoteURL:   remoteURL,
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

		if task.flagName == PathOptionKey {
			task.flagValue = path
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

	err := tea.NewProgram(m).Start()
	if err != nil {
		return provider, fmt.Errorf("could not start tea program: %v", err.Error())
	}

	provider = <-m.msgChan

	return provider, nil
}

// Run displays text inputs to enter or edit all command flag values.
func (wizard *BootstrapWizard) Run(log logger.Logger) error {
	log.Actionf("Please enter or edit command values...")

	m := initialWizardModel(wizard.tasks, wizard.remoteURL, make(chan []*BootstrapCmdOption))

	err := tea.NewProgram(m).Start()
	if err != nil {
		return fmt.Errorf("could not start tea program: %v", err.Error())
	}

	wizard.cmdOptions = <-m.msgChan

	return nil
}

// BuildCmd builds flux bootstrap command options as key/values pairs.
func (wizard *BootstrapWizard) BuildCmd(log logger.Logger) BootstrapWizardCmd {
	log.Actionf("Building flux bootstrap command options...")

	return BootstrapWizardCmd{
		Provider: wizard.gitProvider,
		Options:  wizard.cmdOptions,
	}
}
