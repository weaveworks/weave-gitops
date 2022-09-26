package run

import (
	"fmt"
	"os"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	gogit "github.com/go-git/go-git/v5"
	"github.com/weaveworks/weave-gitops/pkg/logger"
)

type BootstrapWizardTask struct {
	flagName         string
	flagValue        string
	defaultFlagValue string
	flagDescription  string
	isRequired       bool
	isBoolean        bool
}

type BootstrapWizardResult struct {
	flagName  string
	flagValue string
}

type BootstrapWizard struct {
	remoteURL   string
	gitProvider GitProvider
	tasks       []*BootstrapWizardTask
	results     []*BootstrapWizardResult
}

// UI styling
var (
	// table
	baseTableStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	// text inputs style
	focusedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	blurredStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))

	cursorStyle         = focusedStyle.Copy()
	noStyle             = lipgloss.NewStyle()
	helpStyle           = blurredStyle.Copy()
	cursorModeHelpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))

	focusedButton = focusedStyle.Copy().Render("[ Submit ]")
	blurredButton = fmt.Sprintf("[ %s ]", blurredStyle.Render("Submit"))
)

var boostrapGitHubTasks = []*BootstrapWizardTask{
	{
		flagName:        "owner",
		flagValue:       "",
		flagDescription: "GitHub user or organization name",
		isRequired:      true,
	},
	{
		flagName:        "repository",
		flagValue:       "",
		flagDescription: "GitHub repository name",
		isRequired:      true,
	},
	{
		flagName:         "branch",
		flagValue:        "",
		defaultFlagValue: "main",
		flagDescription:  "Git branch (default \"main\")",
	},
	{
		flagName:        "path",
		flagValue:       "",
		flagDescription: "path relative to the repository root, when specified the cluster sync will be scoped to this path",
	},
	{
		flagName:        "personal",
		flagValue:       "",
		flagDescription: "if true, the owner is assumed to be a GitHub user; otherwise an org",
		isRequired:      true,
		isBoolean:       true,
	},
	{
		flagName:         "private",
		flagValue:        "",
		defaultFlagValue: "true",
		flagDescription:  "if true, the repository is setup or configured as private (default true)",
		isRequired:       true,
		isBoolean:        true,
	},
	{
		flagName:        "team",
		flagValue:       "",
		flagDescription: "GitHub team and the access to be given to it(team:maintain)",
	},
	{
		flagName:         "hostname",
		flagValue:        "",
		defaultFlagValue: "github.com",
		flagDescription:  "GitHub hostname (default \"github.com\")",
	},
	{
		flagName:        "ssh-hostname",
		flagValue:       "",
		flagDescription: "SSH hostname, to be used when the SSH host differs from the HTTPS one",
	},
	{
		flagName:        "token-auth",
		flagValue:       "",
		flagDescription: "when enabled, the personal access token will be used instead of SSH deploy key",
		isBoolean:       true,
	},
}

var boostrapGitLabTasks = []*BootstrapWizardTask{
	{
		flagName:        "owner",
		flagValue:       "",
		flagDescription: "GitLab user or group name",
		isRequired:      true,
	},
	{
		flagName:        "repository",
		flagValue:       "",
		flagDescription: "GitLab repository name",
		isRequired:      true,
	},
	{
		flagName:        "branch",
		flagValue:       "",
		flagDescription: "Git branch (default \"main\")",
	},
	{
		flagName:        "path",
		flagValue:       "",
		flagDescription: "path relative to the repository root, when specified the cluster sync will be scoped to this path",
	},
	{
		flagName:        "personal",
		flagValue:       "",
		flagDescription: "if true, the owner is assumed to be a GitLab user; otherwise a group",
		isRequired:      true,
		isBoolean:       true,
	},
	{
		flagName:         "private",
		flagValue:        "",
		defaultFlagValue: "true",
		flagDescription:  "if true, the repository is setup or configured as private (default true)",
		isRequired:       true,
		isBoolean:        true,
	},
	{
		flagName:        "team",
		flagValue:       "",
		flagDescription: "GitLab teams to be given maintainer access (also accepts comma-separated values)",
	},
	{
		flagName:        "hostname",
		flagValue:       "",
		flagDescription: "GitLab hostname (default \"gitlab.com\")",
	},
	{
		flagName:         "token-auth",
		flagValue:        "",
		defaultFlagValue: "false",
		flagDescription:  "when enabled, the personal access token will be used instead of SSH deploy key",
		isRequired:       false,
		isBoolean:        true,
	},
}

var boostrapGitTasks = []*BootstrapWizardTask{
	{
		flagName:        "url",
		flagValue:       "",
		flagDescription: "Git repository URL",
		isRequired:      true,
	},
	{
		flagName:        "password",
		flagValue:       "",
		flagDescription: "basic authentication password",
	},
	{
		flagName:        "private-key-file",
		flagValue:       "",
		flagDescription: "path to a private key file used for authenticating to the Git SSH server",
	},
}

var boostrapBitbucketServerTasks = []*BootstrapWizardTask{
	{
		flagName:        "owner",
		flagValue:       "",
		flagDescription: "Bitbucket Server user or project name",
		isRequired:      true,
	},
	{
		flagName:         "username",
		flagValue:        "",
		defaultFlagValue: "git",
		flagDescription:  "authentication username (default \"git\")",
	},
	{
		flagName:        "repository",
		flagValue:       "",
		flagDescription: "Bitbucket Server repository name",
		isRequired:      true,
	},
	{
		flagName:        "hostname",
		flagValue:       "",
		flagDescription: "Bitbucket Server hostname",
		isRequired:      true,
	},
	{
		flagName:         "branch",
		flagValue:        "",
		defaultFlagValue: "main",
		flagDescription:  "Git branch (default \"main\")",
	},
	{
		flagName:        "path",
		flagValue:       "",
		flagDescription: "path relative to the repository root, when specified the cluster sync will be scoped to this path",
	},
	{
		flagName:        "personal",
		flagValue:       "",
		flagDescription: "if true, the owner is assumed to be a Bitbucket Server user; otherwise a group",
		isRequired:      true,
		isBoolean:       true,
	},
	{
		flagName:         "private",
		flagValue:        "",
		flagDescription:  "if true, the repository is setup or configured as private (default true)",
		defaultFlagValue: "true",
		isRequired:       true,
		isBoolean:        true,
	},
	{
		flagName:        "token-auth",
		flagValue:       "",
		flagDescription: "when enabled, the personal access token will be used instead of SSH deploy key",
		isBoolean:       true,
	},
}

const (
	providerGitHub = "github.com"
	providerGitLab = "gitlab.com"
	gitSuffix      = ".git"
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

// ParseRemoteURLs extracts remote URLs from the repository
func ParseRemoteURLs(log logger.Logger, repo *gogit.Repository) ([]string, error) {
	remoteURLs := []string{}

	if repo != nil {
		remotes, _ := repo.Remotes()

		for _, remote := range remotes {
			remoteURLs = append(remoteURLs, remote.Config().URLs...)
		}
	}

	return remoteURLs, nil
}

// ParseGitProvider extracts git provider from the remote URL, if possible
func ParseGitProvider(log logger.Logger, remoteURL string) GitProvider {
	provider := GitProviderUnknown

	if remoteURL == "" {
		return provider
	}

	if strings.Contains(remoteURL, providerGitHub) {
		return GitProviderGitHub
	}

	if strings.Contains(remoteURL, providerGitLab) {
		return GitProviderGitLab
	}

	return provider
}

// parseOwner extracts owner from the remote URL
func parseOwner(log logger.Logger, remoteURL string) string {
	start := strings.Index(remoteURL, providerGitHub+":") + 1
	if start == -1 {
		start = strings.Index(remoteURL, providerGitHub+"/")
	}

	if start != -1 {
		start += len(providerGitHub)
	}

	end := strings.Index(remoteURL, "/")

	if start == -1 || end == -1 {
		return ""
	}

	return remoteURL[start:end]
}

// parseRepo extracts repository from the remote URL
func parseRepo(log logger.Logger, remoteURL string) string {
	start := strings.LastIndex(remoteURL, "/") + 1
	end := strings.Index(remoteURL, gitSuffix)

	if start == -1 || end == -1 {
		return ""
	}

	return remoteURL[start:end]
}

// NewBootstrapWizard creates a wizard to gather
// all bootstrap config options before running flux bootstrap.
func NewBootstrapWizard(log logger.Logger, remoteURL string, gitProvider GitProvider) (*BootstrapWizard, error) {
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

	if remoteURL != "" {
		// if possible, parse owner and repository from the remote URL
		owner := ""
		repo := ""

		if gitProvider == GitProviderGitHub {
			owner = parseOwner(log, remoteURL)
		}

		repo = parseRepo(log, remoteURL)

		for _, task := range wizard.tasks {
			if task.flagName == "owner" {
				task.flagValue = owner
			}

			if task.flagName == "repository" {
				task.flagValue = repo
			}
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

	m := initialWizardModel(wizard.tasks, wizard.remoteURL, make(chan []*BootstrapWizardResult))

	err := tea.NewProgram(m).Start()
	if err != nil {
		return fmt.Errorf("could not start tea program: %v", err.Error())
	}

	wizard.results = <-m.msgChan

	return nil
}

// BuildCommand builds flux bootstrap command as a text string.
func (wizard *BootstrapWizard) BuildCommand(log logger.Logger) string {
	log.Actionf("Building flux bootstrap command...")

	cmd := "flux bootstrap "

	switch wizard.gitProvider {
	case GitProviderGitHub:
		cmd += "github"
	case GitProviderGitLab:
		cmd += "gitlab"
	case GitProviderGit:
		cmd += "git"
	case GitProviderBitbucketServer:
		cmd += "bitbucket-server"
	}

	cmd += " \\\n"

	for i, result := range wizard.results {
		cmd += "  --" + result.flagName + "=" + result.flagValue

		if i < len(wizard.results)-1 {
			cmd += " \\\n"
		}
	}

	return cmd
}
