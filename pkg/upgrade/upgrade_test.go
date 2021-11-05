package upgrade

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/jenkins-x/go-scm/scm/factory"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/pctl/pkg/catalog"
	pctl_git "github.com/weaveworks/pctl/pkg/git"
	pctl_git_fake "github.com/weaveworks/pctl/pkg/git/fakes"
	"github.com/weaveworks/pctl/pkg/install"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/git"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/git/wrapper"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

// Test key that can be successfully loaded into gitssh.NewPublicKeys
const testKey string = `
-----BEGIN OPENSSH PRIVATE KEY-----
b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
QyNTUxOQAAACA+VhTH1eCMXtv4da5ikqz4TWMx3gzZfc2AqPUoFwnA6wAAAJA0gcXzNIHF
8wAAAAtzc2gtZWQyNTUxOQAAACA+VhTH1eCMXtv4da5ikqz4TWMx3gzZfc2AqPUoFwnA6w
AAAEAPlKRw3KV5eAE8tNELDZrA0ViK7/L4rEaVjVKUlHJv/T5WFMfV4Ixe2/h1rmKSrPhN
YzHeDNl9zYCo9SgXCcDrAAAAB3Rlc3RrZXkBAgMEBQY=
-----END OPENSSH PRIVATE KEY-----
`

func TestBuildUpgradeConfigs(t *testing.T) {
	tests := []struct {
		name            string
		localGitRemote  string
		clusterState    []runtime.Object
		upgradeValues   UpgradeValues
		expectedRepo    string
		expectedGitRepo string
		expectedErr     error
	}{
		{
			name:            "Good repo form and GitRepository present",
			localGitRemote:  "git@github.com:org/repo.git",
			clusterState:    []runtime.Object{createGitRepository("repo")},
			upgradeValues:   UpgradeValues{Namespace: wego.DefaultNamespace},
			expectedRepo:    "org/repo",
			expectedGitRepo: filepath.Join(wego.DefaultNamespace, "repo"),
		},
		{
			name:           "Good repo form but GitRepository missing",
			localGitRemote: "git@github.com:org/repo.git",
			upgradeValues:  UpgradeValues{Namespace: wego.DefaultNamespace},
			expectedErr:    fmt.Errorf("couldn't find GitRepository resource \"%s/repo\" in the cluster, please specify", wego.DefaultNamespace),
		},
		{
			name:            "Specify an alterative gitRepo",
			localGitRemote:  "git@github.com:org/repo.git",
			clusterState:    []runtime.Object{createGitRepository("foo")},
			upgradeValues:   UpgradeValues{Namespace: wego.DefaultNamespace, GitRepository: filepath.Join(wego.DefaultNamespace, "foo")},
			expectedRepo:    "org/repo",
			expectedGitRepo: filepath.Join(wego.DefaultNamespace, "foo"),
		},
		{
			name:            "Specify an alterative repo",
			localGitRemote:  "git@github.com:org/repo.git",
			clusterState:    []runtime.Object{createGitRepository("repo")},
			upgradeValues:   UpgradeValues{Namespace: wego.DefaultNamespace, RepoOrgAndName: "foo/bar"},
			expectedRepo:    "foo/bar",
			expectedGitRepo: filepath.Join(wego.DefaultNamespace, "repo"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeClient := makeClient(t, tt.clusterState...)
			gitClient := &gitfakes.FakeGit{}
			gitClient.GetRemoteUrlStub = func(dir, remote string) (string, error) {
				return tt.localGitRemote, nil
			}
			uc, err := buildUpgradeConfigs(context.TODO(), tt.upgradeValues, kubeClient, gitClient, os.Stdout)
			assert.Equal(t, tt.expectedErr, err)
			if uc != nil {
				assert.Equal(t, tt.expectedRepo, uc.SCMConfig.Repo)
				assert.Equal(t, tt.expectedGitRepo, uc.InstallConfig.GitRepoNamespace+"/"+uc.InstallConfig.GitRepoName)
			}
		})
	}
}

func TestToUpgradeConfigs(t *testing.T) {
	uv := UpgradeValues{
		RepoOrgAndName: "org/repo",
		Remote:         "origin",
		HeadBranch:     "upgrade-to-wge",
		BaseBranch:     "main",
		CommitMessage:  "lets upgrade!",
		Namespace:      wego.DefaultNamespace,
		ProfileBranch:  "main",
		ConfigMap:      "my-values",
		Out:            "config/repo/subdir",
		GitRepository:  filepath.Join(wego.DefaultNamespace, "repo"),
	}

	uc, err := toUpgradeConfigs(uv)
	assert.NoError(t, err)
	assert.Equal(t, &UpgradeConfigs{
		CLIGitConfig: pctl_git.CLIGitConfig{
			Directory: ".",
			Branch:    "upgrade-to-wge",
			Remote:    "origin",
			Message:   "lets upgrade!",
			Base:      "main",
		},
		SCMConfig: pctl_git.SCMConfig{
			Branch: "upgrade-to-wge",
			Base:   "main",
			Repo:   "org/repo",
		},
		InstallConfig: install.Config{
			RootDir:          "config/repo/subdir/.weave-gitops/clusters/management/user",
			GitRepoNamespace: wego.DefaultNamespace,
			GitRepoName:      "repo",
		},
		Profile: catalog.Profile{
			GitRepoConfig: catalog.GitRepoConfig{
				Name:      "repo",
				Namespace: wego.DefaultNamespace,
			},
			ProfileConfig: catalog.ProfileConfig{
				ConfigMap:     "my-values",
				Namespace:     wego.DefaultNamespace,
				SubName:       "weave-gitops-enterprise",
				ProfileBranch: "main",
				Path:          ".weave-gitops/clusters/management/user",
			},
		},
	}, uc)
}

func TestToUpgradeConfigsErrors(t *testing.T) {
	tests := []struct {
		name string
		uv   UpgradeValues
		err  error
	}{
		{"Bad repo form", UpgradeValues{GitRepository: "repo"}, errors.New("git-repository must in format <namespace>/<name>; was: repo")},
		{"Good repo form", UpgradeValues{GitRepository: "repo/org"}, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := toUpgradeConfigs(tt.uv)
			assert.Equal(t, err, tt.err)
		})
	}
}

func TestUpgrade(t *testing.T) {
	tempDir := createLocalProfileRepo(t)
	defer os.RemoveAll(tempDir)

	configDir := createLocalClusterConfigRepo(t)
	defer os.RemoveAll(configDir)

	// All the fake clients
	fakePctlGitClient := &pctl_git_fake.FakeGit{}
	fakeScm, err := factory.NewClient("fake", "", "")
	assert.NoError(t, err)
	scmClient, err := pctl_git.NewClient(pctl_git.SCMConfig{Client: fakeScm})
	assert.NoError(t, err)

	gitClient := git.New(nil, wrapper.NewGoGit())

	// Run upgrade!
	uc, err := toUpgradeConfigs(UpgradeValues{
		Remote:        "origin",
		HeadBranch:    "upgrade-to-wge",
		ProfileBranch: "main",
		BaseBranch:    "main",
		CommitMessage: "Upgrade to wge",
		Namespace:     wego.DefaultNamespace,
		Out:           configDir,
		GitRepository: filepath.Join(wego.DefaultNamespace, "my-app"),
	})
	assert.NoError(t, err)

	// dry run
	err = upgrade(context.Background(), "file://"+tempDir, *uc, gitClient, fakePctlGitClient, scmClient, true)
	assert.NoError(t, err)
	contents, err := ioutil.ReadFile(path.Join(configDir, ".weave-gitops/clusters/management/user", "profile-installation.yaml"))
	assert.NoError(t, err)
	assert.NotNil(t, contents)
	os.RemoveAll(path.Join(configDir, ".weave-gitops/clusters/management/user"))
	// fake git not called
	pushCount := fakePctlGitClient.PushCallCount()
	assert.Equal(t, 0, pushCount)

	err = upgrade(context.Background(), "file://"+tempDir, *uc, gitClient, fakePctlGitClient, scmClient, false)
	assert.NoError(t, err)
	contents, err = ioutil.ReadFile(path.Join(configDir, ".weave-gitops/clusters/management/user", "profile-installation.yaml"))
	assert.NoError(t, err)
	assert.NotNil(t, contents)
	os.RemoveAll(path.Join(configDir, ".weave-gitops/clusters/management/user"))
	// fake git would have pushed etc
	pushCount = fakePctlGitClient.PushCallCount()
	assert.Equal(t, 1, pushCount)
}

func TestGetGitAuthFromDeployKey(t *testing.T) {
	tests := []struct {
		name         string
		clusterState []runtime.Object
		expected     string
		expectedErr  error
	}{
		{
			name:        "error returned",
			expectedErr: errors.New("failed to get entitlement: secrets \"weave-gitops-enterprise-credentials\" not found"),
		},
		{
			name:         "preflight check pass",
			clusterState: []runtime.Object{createSecret()},
			expected:     "ssh-public-keys",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeClient := makeClient(t, tt.clusterState...)
			key, err := getGitAuthFromDeployKey(context.Background(), kubeClient, wego.DefaultNamespace)
			assert.Equal(t, err, tt.expectedErr)
			if key != nil {
				assert.Equal(t, tt.expected, key.Name())
			}
		})
	}
}

func TestGetRepoOrgAndName(t *testing.T) {
	repoPath, err := getRepoOrgAndName("git@github.com:ww/repo.git")
	assert.NoError(t, err)
	assert.Equal(t, "ww/repo", repoPath)

	repoPath, err = getRepoOrgAndName("https://github.com/ww/repo.git")
	assert.NoError(t, err)
	assert.Equal(t, "ww/repo", repoPath)
}

//
// helpers
//

func makeClient(t *testing.T, clusterState ...runtime.Object) client.Client {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		sourcev1.AddToScheme,
	}
	err := schemeBuilder.AddToScheme(scheme)
	assert.NoError(t, err)

	return fake.NewClientBuilder().
		WithScheme(scheme).
		WithRuntimeObjects(clusterState...).
		Build()
}

func createSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weave-gitops-enterprise-credentials",
			Namespace: wego.DefaultNamespace,
		},
		Type: "Opaque",
		Data: map[string][]byte{"deploy-key": []byte(testKey)},
	}
}

func createGitRepository(name string) *sourcev1.GitRepository {
	return &sourcev1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: wego.DefaultNamespace,
		},
	}
}

func createLocalClusterConfigRepo(t *testing.T) string {
	configGitClient := git.New(nil, wrapper.NewGoGit())
	configDir, err := ioutil.TempDir("", "git-config-")
	assert.NoError(t, err)
	_, err = configGitClient.Init(configDir, "https://github.com/github/gitignore", "main")
	assert.NoError(t, err)

	return configDir
}

func createLocalProfileRepo(t *testing.T) string {
	tempDir, err := ioutil.TempDir("", "local-profile-")
	assert.NoError(t, err)

	gitClient := git.New(nil, wrapper.NewGoGit())
	_, err = gitClient.Init(tempDir, "https://github.com/github/gitignore", "main")
	assert.NoError(t, err)

	content, err := ioutil.ReadFile("testdata/profile.yaml")
	assert.NoError(t, err)
	err = gitClient.Write("/.weave-gitops/clusters/management/user/profile.yaml", content)
	assert.NoError(t, err)
	err = gitClient.Write("/.weave-gitops/clusters/management/user/nginx/deployment/deployment.yaml", []byte("test deployment"))
	assert.NoError(t, err)

	_, err = gitClient.Commit(git.Commit{
		Author:  git.Author{Name: "test", Email: "test@example.com"},
		Message: "test commit",
	})
	assert.NoError(t, err)

	return tempDir
}
