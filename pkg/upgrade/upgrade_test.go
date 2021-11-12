package upgrade

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/stretchr/testify/assert"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

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
			uv, err := buildUpgradeConfigs(context.TODO(), tt.upgradeValues, kubeClient, gitClient, os.Stdout)
			assert.Equal(t, tt.expectedErr, err)
			if uv != nil {
				assert.Equal(t, tt.expectedRepo, uv.RepoOrgAndName, "RepoOrgAndName")
				assert.Equal(t, tt.expectedGitRepo, uv.GitRepository, "GitRepository")
			}
		})
	}
}

func TestUpgradeDryRun(t *testing.T) {
	gitClient := &gitfakes.FakeGit{}
	gitClient.GetRemoteUrlStub = func(dir, remote string) (string, error) {
		return "git@github.com:org/repo.git", nil
	}
	kubeClient := makeClient(t, createSecret(), createGitRepository("my-app"))
	gitProvider := &gitprovidersfakes.FakeGitProvider{}
	logger := &loggerfakes.FakeLogger{}

	var output bytes.Buffer

	// Run upgrade!
	err := upgrade(context.TODO(), UpgradeValues{
		Remote:        "origin",
		HeadBranch:    "upgrade-to-wge",
		ProfileBranch: "main",
		BaseBranch:    "main",
		CommitMessage: "Upgrade to wge",
		Namespace:     wego.DefaultNamespace,
		DryRun:        true,
		GitRepository: filepath.Join(wego.DefaultNamespace, "my-app"),
	}, gitClient, kubeClient, gitProvider, logger, &output)

	assert.Contains(t, output.String(), "kind: HelmRelease")
	assert.Contains(t, output.String(), "kind: HelmRepository")
	assert.NoError(t, err)
}

func TestUpgrade(t *testing.T) {
	gitClient := &gitfakes.FakeGit{}
	gitClient.GetRemoteUrlStub = func(dir, remote string) (string, error) {
		return "git@github.com:org/repo.git", nil
	}
	kubeClient := makeClient(t, createSecret(), createGitRepository("my-app"))
	gitProvider := &gitprovidersfakes.FakeGitProvider{}
	logger := &loggerfakes.FakeLogger{}

	var output bytes.Buffer

	// Run upgrade!
	err := upgrade(context.TODO(), UpgradeValues{
		Remote:        "origin",
		HeadBranch:    "upgrade-to-wge",
		ProfileBranch: "main",
		BaseBranch:    "main",
		CommitMessage: "Upgrade to wge",
		Namespace:     wego.DefaultNamespace,
		GitRepository: filepath.Join(wego.DefaultNamespace, "my-app"),
	}, gitClient, kubeClient, gitProvider, logger, &output)

	assert.NoError(t, err)
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
			name: "user pass missing",
			clusterState: []runtime.Object{createSecret(func(s *corev1.Secret) {
				s.Data["username"] = []byte("")
			})},
			expectedErr: errors.New("username or password missing in entitlement secret, may be an old entitlement file"),
		},
		{
			name:         "preflight check pass",
			clusterState: []runtime.Object{createSecret()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			kubeClient := makeClient(t, tt.clusterState...)
			err := getBasicAuth(context.Background(), kubeClient, wego.DefaultNamespace)
			assert.Equal(t, tt.expectedErr, err)
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

func createSecret(opts ...func(*corev1.Secret)) *corev1.Secret {
	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weave-gitops-enterprise-credentials",
			Namespace: wego.DefaultNamespace,
		},
		Type: "Opaque",
		Data: map[string][]byte{
			"entitlement": []byte("hi"),
			"username":    []byte("hi"),
			"password":    []byte("hi"),
		},
	}

	for _, fn := range opts {
		fn(s)
	}

	return s
}

func createGitRepository(name string) *sourcev1.GitRepository {
	return &sourcev1.GitRepository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: wego.DefaultNamespace,
		},
	}
}
