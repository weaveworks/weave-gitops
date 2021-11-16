package upgrade

import (
	"bytes"
	"context"
	"errors"
	"path/filepath"
	"testing"

	sourcev1 "github.com/fluxcd/source-controller/api/v1beta1"
	"github.com/stretchr/testify/assert"
	wego "github.com/weaveworks/weave-gitops/api/v1alpha1"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube/kubefakes"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestUpgradeDryRun(t *testing.T) {
	gitClient := &gitfakes.FakeGit{}
	gitClient.GetRemoteUrlStub = func(dir, remote string) (string, error) {
		return "git@github.com:org/repo.git", nil
	}
	kubeClient := makeClient(t, createSecret(), createGitRepository("my-app"))
	gitProvider := &gitprovidersfakes.FakeGitProvider{}
	logger := &loggerfakes.FakeLogger{}
	k := &kubefakes.FakeKube{}

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
	}, k, gitClient, kubeClient, gitProvider, logger, &output)

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
	k := &kubefakes.FakeKube{}

	var output bytes.Buffer

	// Run upgrade!
	err := upgrade(context.TODO(), UpgradeValues{
		Remote:        "origin",
		AppConfigURL:  "https://github.com/test/example.git",
		HeadBranch:    "upgrade-to-wge",
		ProfileBranch: "main",
		BaseBranch:    "main",
		CommitMessage: "Upgrade to wge",
		Namespace:     wego.DefaultNamespace,
		GitRepository: filepath.Join(wego.DefaultNamespace, "my-app"),
	}, k, gitClient, kubeClient, gitProvider, logger, &output)

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

func TestMakeHelmResources(t *testing.T) {
	tests := []struct {
		name        string
		values      []string
		expected    string
		expectedErr error
	}{
		{
			name:        "error returned",
			values:      []string{"key1"},
			expectedErr: errors.New("failed parsing --set data: key \"key1\" has no value"),
		},
		{
			name:     "resources created",
			values:   []string{"key1=val2"},
			expected: "key1: val2",
		},
		{
			name:     "override default values",
			values:   []string{"key1=val2", "config.cluster.name=bar"},
			expected: "name: bar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			objs, err := makeHelmResources("wego-system", "v0.0.14", "foo", "example.com", tt.values)
			out, _ := marshalToYamlStream(objs)
			assert.Equal(t, tt.expectedErr, err)
			assert.Contains(t, string(out), tt.expected)
		})
	}
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
