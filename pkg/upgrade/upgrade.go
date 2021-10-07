package upgrade

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/go-git/go-git/v5/plumbing/transport"
	gitssh "github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/weaveworks/pctl/pkg/catalog"
	pctl_git "github.com/weaveworks/pctl/pkg/git"
	"github.com/weaveworks/pctl/pkg/install"
	"github.com/weaveworks/pctl/pkg/runner"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type UpgradeValues struct {
	RepoOrgAndName string
	Remote         string
	HeadBranch     string
	BaseBranch     string
	CommitMessage  string
	Namespace      string
	ProfileBranch  string
	ConfigMap      string
	Out            string
	GitRepository  string
}

type UpgradeConfigs struct {
	CLIGitConfig  pctl_git.CLIGitConfig
	SCMConfig     pctl_git.SCMConfig
	InstallConfig install.Config
	Profile       catalog.Profile
}

const EnterpriseProfileURL string = "git@github.com:weaveworks/weave-gitops-enterprise-profiles.git"

// Upgrade installs the private weave-gitops-enterprise profile into the working directory:
// 1. Private deploy key is read from a secret in the cluster
// 2. Private profiles repo is cloned locally using the deploy key
// 3. pctl is used to install the profile from the local clone into the current working directory
// 4. pctl is used to add, commit, push and create a PR.
//
func Upgrade(ctx context.Context, upgradeValues UpgradeValues, w io.Writer) error {

	//
	// Clients needed for building the config
	//

	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{sourcev1.AddToScheme}
	err := schemeBuilder.AddToScheme(scheme)
	if err != nil {
		return fmt.Errorf("error adding sourcev1 to kube client scheme %v", err)
	}
	kubeClientConfig := config.GetConfigOrDie()
	kubeClient, err := client.New(kubeClientConfig, client.Options{Scheme: scheme})
	if err != nil {
		return fmt.Errorf("error creating client for cluster %v", err)
	}

	gitClient := git.New(nil, wrapper.NewGoGit())
	uc, err := buildUpgradeConfigs(ctx, upgradeValues, kubeClient, gitClient, w)
	if err != nil {
		return fmt.Errorf("failed to build upgrade configs: %v", err)
	}
	auth, err := getGitAuthFromDeployKey(ctx, kubeClient, upgradeValues.Namespace)
	if err != nil {
		return fmt.Errorf("failed to load deploy key for profiles repos from cluster: %v", err)
	}

	//
	// Clients needed for installing the profile
	//

	// New client with auth from the cluster
	gitClientWithAuth := git.New(auth, wrapper.NewGoGit())
	pctlGitClient := pctl_git.NewCLIGit(uc.CLIGitConfig, &runner.CLIRunner{})
	pctlSCMClient, err := pctl_git.NewClient(uc.SCMConfig)
	if err != nil {
		return fmt.Errorf("failed to create scm client: %w", err)
	}

	return upgrade(ctx, EnterpriseProfileURL, *uc, gitClientWithAuth, pctlGitClient, pctlSCMClient)
}

func buildUpgradeConfigs(ctx context.Context, uv UpgradeValues, kubeClient client.Client, gitClient git.Git, w io.Writer) (*UpgradeConfigs, error) {

	repoUrlString, err := gitClient.GetRemoteUrl(".", uv.Remote)
	if err != nil {
		return nil, err
	}

	// Calculate defaults from current working directory
	if uv.RepoOrgAndName == "" {
		githubRepoPath, err := getRepoOrgAndName(repoUrlString)
		if err != nil {
			return nil, err
		}

		fmt.Fprintf(w, "Deriving org/repo for PR as %v\n", githubRepoPath)
		uv.RepoOrgAndName = githubRepoPath
	}

	if uv.GitRepository == "" {
		gitRepositoryNameNamespace := fmt.Sprintf("%s/%s", uv.Namespace, strings.TrimSuffix(filepath.Base(repoUrlString), ".git"))
		fmt.Fprintf(w, "Deriving name of GitRepository Resource as %v\n", gitRepositoryNameNamespace)
		uv.GitRepository = gitRepositoryNameNamespace
	}

	err = ensureGitRepositoryResource(context.Background(), kubeClient, uv.GitRepository)
	if err != nil {
		return nil, err
	}

	return toUpgradeConfigs(uv)
}

func toUpgradeConfigs(uv UpgradeValues) (*UpgradeConfigs, error) {

	gitRepoNamespace, gitRepoName, err := getGitRepositoryNamespaceAndName(uv.GitRepository)
	if err != nil {
		return nil, err
	}

	return &UpgradeConfigs{
		CLIGitConfig: pctl_git.CLIGitConfig{
			Directory: uv.Out,
			Branch:    uv.HeadBranch,
			Remote:    uv.Remote,
			Base:      uv.BaseBranch,
			Message:   uv.CommitMessage,
		},
		SCMConfig: pctl_git.SCMConfig{
			Branch: uv.HeadBranch,
			Base:   uv.BaseBranch,
			Repo:   uv.RepoOrgAndName,
		},
		InstallConfig: install.Config{
			RootDir:          filepath.Join(uv.Out, "weave-gitops-enterprise"),
			GitRepoNamespace: gitRepoNamespace,
			GitRepoName:      gitRepoName,
		},
		Profile: catalog.Profile{
			ProfileConfig: catalog.ProfileConfig{
				ConfigMap:     uv.ConfigMap,
				Namespace:     uv.Namespace,
				Path:          "weave-gitops-enterprise",
				ProfileBranch: uv.ProfileBranch,
				SubName:       "weave-gitops-enterprise",
			},
			GitRepoConfig: catalog.GitRepoConfig{
				Namespace: gitRepoNamespace,
				Name:      gitRepoName,
			},
		},
	}, nil
}

func upgrade(ctx context.Context, enterpriseProfileURL string, uc UpgradeConfigs, gitClient git.Git, pctlGitClient pctl_git.Git, pctlSCMClient pctl_git.SCMClient) error {
	tempDir, err := ioutil.TempDir("", "git-")
	if err != nil {
		return fmt.Errorf("failed to create temp folder for remote git clone of profiles repo: %w", err)
	}
	defer os.RemoveAll(tempDir)

	_, err = gitClient.Clone(
		ctx,
		tempDir,
		enterpriseProfileURL,
		uc.Profile.ProfileConfig.ProfileBranch,
	)
	if err != nil {
		return fmt.Errorf("failed to clone git repo: %v", err)
	}

	profileDefinition := uc.Profile
	profileDefinition.URL = "file://" + tempDir
	err = addProfile(uc.InstallConfig, profileDefinition, uc.CLIGitConfig)
	if err != nil {
		return err
	}

	err = catalog.CreatePullRequest(pctlSCMClient, pctlGitClient, uc.SCMConfig.Branch, uc.InstallConfig.RootDir)
	if err != nil {
		return err
	}

	return nil
}

// addProfile installs the profile into a local git repo.
func addProfile(installConfig install.Config, profile catalog.Profile, cliGitConfig pctl_git.CLIGitConfig) error {
	manager := &catalog.Manager{}

	// We're not interested in mocking this out for now (it just runs git commands on the local fs) so we make an inline client.
	gitClient := pctl_git.NewCLIGit(cliGitConfig, &runner.CLIRunner{})
	installConfig.GitClient = gitClient

	return manager.Install(catalog.InstallConfig{
		Clients: catalog.Clients{Installer: install.NewInstaller(installConfig)},
		Profile: profile,
	})
}

func getGitAuthFromDeployKey(ctx context.Context, kubeClient client.Client, ns string) (transport.AuthMethod, error) {
	key, err := getDeployKey(ctx, kubeClient, ns)
	if err != nil {
		return nil, err
	}

	return gitssh.NewPublicKeys("git", key, "")
}

func getDeployKey(ctx context.Context, kubeClient client.Client, ns string) ([]byte, error) {
	deployKeySecret := &v1.Secret{}
	err := kubeClient.Get(ctx, client.ObjectKey{
		Namespace: ns,
		Name:      "weave-gitops-enterprise-credentials",
	}, deployKeySecret)
	if err != nil {
		return nil, fmt.Errorf("failed to get entitlement: %v", err)
	}

	key := deployKeySecret.Data["deploy-key"]
	return key, nil
}

func ensureGitRepositoryResource(ctx context.Context, kubeClient client.Client, gitRepository string) error {
	gitRepoNamespace, gitRepoName, err := getGitRepositoryNamespaceAndName(gitRepository)
	if err != nil {
		return err
	}

	gitRepo := &sourcev1.GitRepository{}
	err = kubeClient.Get(ctx, client.ObjectKey{
		Namespace: gitRepoNamespace,
		Name:      gitRepoName,
	}, gitRepo)
	if errors.IsNotFound(err) {
		return fmt.Errorf("couldn't find GitRepository %v/%v to install into", gitRepoNamespace, gitRepoName)
	}
	if err != nil {
		return fmt.Errorf("failed to look up GitRepository %v/%v to install into: %v", gitRepoNamespace, gitRepoName, err)
	}
	return nil
}

func getGitRepositoryNamespaceAndName(gitRepository string) (string, string, error) {
	split := strings.Split(gitRepository, "/")
	if len(split) != 2 {
		return "", "", fmt.Errorf("git-repository must in format <namespace>/<name>; was: %s", gitRepository)
	}

	return split[0], split[1], nil
}

// getRepoOrgAndName transforms git remotes to github "paths"
// - git@github.com:org/repo.git -> org/repo
// - https://github.com/org/repo.git -> org/repo
//
func getRepoOrgAndName(url string) (string, error) {
	repoEndpoint, err := transport.NewEndpoint(url)
	if err != nil {
		return "", err
	}

	return strings.TrimPrefix(strings.TrimSuffix(repoEndpoint.Path, ".git"), "/"), nil
}
