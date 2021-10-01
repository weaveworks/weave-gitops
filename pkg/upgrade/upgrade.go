package upgrade

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/weaveworks/pctl/pkg/bootstrap"
	"github.com/weaveworks/pctl/pkg/catalog"
	"github.com/weaveworks/pctl/pkg/git"
	"github.com/weaveworks/pctl/pkg/install"
	"github.com/weaveworks/pctl/pkg/runner"
	go_git "gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing"
	"gopkg.in/src-d/go-git.v4/plumbing/transport"
	gitssh "gopkg.in/src-d/go-git.v4/plumbing/transport/ssh"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type UpgradeValues struct {
	RepoOrgAndName string
	Remote         string
	HeadBranch     string
	BaseBranch     string
	CommitMessage  string
	Name           string
	Namespace      string
	ProfileBranch  string
	ConfigMap      string
	Out            string
	ProfileRepoURL string
	ProfilePath    string
	GitRepository  string
	Version        string
}

func Upgrade(upgradeValues UpgradeValues, w io.Writer) error {
	config := config.GetConfigOrDie()

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "Checking if entitlement exists...\n")

	entitlement, err := getEntitlement(clientset)
	if err != nil {
		return err
	}

	upgradeValues.Name = "weave-gitops-enterprise"
	upgradeValues.ProfileBranch = "main"
	upgradeValues.ConfigMap = ""
	upgradeValues.Out = "."
	upgradeValues.ProfileRepoURL = "git@github.com:weaveworks/weave-gitops-enterprise-profiles.git"
	upgradeValues.ProfilePath = "weave-gitops-enterprise"

	repoURL, err := getRepoURL(upgradeValues.Remote)
	if err != nil {
		return err
	}

	if upgradeValues.RepoOrgAndName == "" {
		githubRepoPath, err := getRepoOrgAndName(repoURL)
		if err != nil {
			return err
		}

		fmt.Fprintf(w, "Deriving org/repo for PR as %v\n", githubRepoPath)
		upgradeValues.RepoOrgAndName = githubRepoPath
	}

	if upgradeValues.GitRepository == "" {
		gitRepositoryNameNamespace := "wego-system/" + strings.TrimSuffix(filepath.Base(repoURL), ".git")
		fmt.Fprintf(w, "Deriving name of GitRepository Resource as %v\n", gitRepositoryNameNamespace)
		upgradeValues.GitRepository = gitRepositoryNameNamespace
	}

	key := entitlement.Data["deploy-key"]

	localRepo, err := cloneToTempDir("", upgradeValues.ProfileRepoURL, upgradeValues.ProfileBranch, key, w)
	if err != nil {
		return err
	}

	upgradeValues.ProfileRepoURL = localRepo.WorktreeDir()

	installationDirectory, err := addProfile(upgradeValues)
	if err != nil {
		return err
	}

	if err := createPullRequest(upgradeValues, installationDirectory); err != nil {
		return err
	}

	fmt.Fprintf(w, "Upgrade pull request created\n")

	return nil
}

func getRepoOrgAndName(url string) (string, error) {
	repoEndpoint, err := transport.NewEndpoint(url)
	if err != nil {
		return "", err
	}

	return strings.Trim(strings.TrimSuffix(strings.TrimSpace(repoEndpoint.Path), ".git"), "/"), nil
}

func cloneToTempDir(parentDir, gitURL, branch string, privKey []byte, w io.Writer) (*GitRepo, error) {
	fmt.Fprintf(w, "Creating a temp directory...")

	gitDir, err := ioutil.TempDir(parentDir, "git-")
	if err != nil {
		return nil, fmt.Errorf("failed to create tem directory: %v", err)
	}

	fmt.Fprintf(w, "Temp directory %q created.", gitDir)

	fmt.Fprintf(w, "Cloning the Git repository %q to %q...", gitURL, gitDir)

	auth, err := gitssh.NewPublicKeys("git", privKey, "")
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate: %v", err)
	}

	repo, err := go_git.PlainClone(gitDir, false, &go_git.CloneOptions{
		URL:           gitURL,
		Auth:          auth,
		ReferenceName: plumbing.NewBranchReferenceName(branch),

		SingleBranch: true,
		Tags:         go_git.NoTags,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to clone git repo: %v", err)
	}

	fmt.Fprintf(w, "Cloned repo: %s", gitURL)

	return &GitRepo{
		worktreeDir: gitDir,
		repo:        repo,
		auth:        auth,
	}, nil
}

type GitRepo struct {
	worktreeDir string
	repo        *go_git.Repository
	auth        *gitssh.PublicKeys
}

func (gr *GitRepo) WorktreeDir() string {
	return gr.worktreeDir
}

func getRepoURL(remote string) (string, error) {
	cmd := exec.Command("git", "config", "--get", "remote."+remote+".url")
	cmd.Dir = "."

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(stdout.String()), nil
}

func getEntitlement(clientset kubernetes.Interface) (*v1.Secret, error) {
	var entitlement *v1.Secret

	entitlement, err := clientset.CoreV1().Secrets("wego-system").Get(context.Background(), "weave-gitops-enterprise-credentials", metav1.GetOptions{})
	if err != nil {
		return entitlement, fmt.Errorf("failed to get entitlement: %v", err)
	}

	return entitlement, nil
}

func addProfile(values UpgradeValues) (string, error) {
	url := values.ProfileRepoURL

	r := &runner.CLIRunner{}
	g := git.NewCLIGit(git.CLIGitConfig{
		Message: values.CommitMessage,
	}, r)

	gitRepoNamespace, gitRepoName, err := getGitRepositoryNamespaceAndName(values.GitRepository)
	if err != nil {
		return "", err
	}

	installationDirectory := filepath.Join(values.Out, values.Name)
	installer := install.NewInstaller(install.Config{
		GitClient:        g,
		RootDir:          installationDirectory,
		GitRepoNamespace: gitRepoNamespace,
		GitRepoName:      gitRepoName,
	})

	cfg := catalog.InstallConfig{
		Clients: catalog.Clients{
			Installer: installer,
		},
		Profile: catalog.Profile{
			ProfileConfig: catalog.ProfileConfig{
				ConfigMap:     values.ConfigMap,
				Namespace:     values.Namespace,
				Path:          values.ProfilePath,
				ProfileBranch: values.ProfileBranch,
				ProfileName:   "",
				SubName:       values.Name,
				URL:           url,
				Version:       values.Version,
			},
			GitRepoConfig: catalog.GitRepoConfig{
				Namespace: gitRepoNamespace,
				Name:      gitRepoName,
			},
		},
	}
	manager := &catalog.Manager{}
	err = manager.Install(cfg)

	return installationDirectory, err
}

func getGitRepositoryNamespaceAndName(gitRepository string) (string, string, error) {
	if gitRepository != "" {
		split := strings.Split(gitRepository, "/")
		if len(split) != 2 {
			return "", "", fmt.Errorf("git-repository must in format <namespace>/<name>; was: %s", gitRepository)
		}

		return split[0], split[1], nil
	}

	wd, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("failed to fetch current working directory: %w", err)
	}

	config := bootstrap.GetConfig(wd)
	if err == nil && config != nil {
		return config.GitRepository.Namespace, config.GitRepository.Name, nil
	}

	return "", "", fmt.Errorf("flux git repository not provided, please provide the --git-repository flag or use the pctl bootstrap functionality")
}

func createPullRequest(values UpgradeValues, installationDirectory string) error {
	r := &runner.CLIRunner{}
	g := git.NewCLIGit(git.CLIGitConfig{
		Directory: values.Out,
		Branch:    values.HeadBranch,
		Remote:    values.Remote,
		Base:      values.BaseBranch,
		Message:   values.CommitMessage,
	}, r)

	scmClient, err := git.NewClient(git.SCMConfig{
		Branch: values.HeadBranch,
		Base:   values.BaseBranch,
		Repo:   values.RepoOrgAndName,
	})
	if err != nil {
		return fmt.Errorf("failed to create scm client: %w", err)
	}

	return catalog.CreatePullRequest(scmClient, g, values.HeadBranch, installationDirectory)
}
