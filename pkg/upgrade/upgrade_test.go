package upgrade

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/fluxcd/go-git-providers/gitprovider"
	sourcev1 "github.com/fluxcd/source-controller/api/v1beta2"
	"github.com/stretchr/testify/assert"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"github.com/weaveworks/weave-gitops/pkg/logger/loggerfakes"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const testNamespace = "some-namespace"

func TestUpgradeDryRun(t *testing.T) {
	gitClient := &gitfakes.FakeGit{}
	kubeClient := makeClient(t, createSecret(), createGitRepository("my-app"))
	gitProvider := &gitprovidersfakes.FakeGitProvider{}
	gitProvider.CreatePullRequestStub = func(context.Context, gitproviders.RepoURL, gitproviders.PullRequestInfo) (gitprovider.PullRequest, error) {
		return &mockPullRequest{}, nil
	}
	logger := &loggerfakes.FakeLogger{}

	var output bytes.Buffer

	// Run upgrade!
	err := upgrade(context.TODO(), UpgradeValues{
		ConfigRepo:    "ssh://git@github.com/my-org/my-management-cluster.git",
		HeadBranch:    "upgrade-to-wge",
		BaseBranch:    "main",
		CommitMessage: "Upgrade to wge",
		Namespace:     testNamespace,
		DryRun:        true,
	}, kubeClient, gitClient, gitProvider, logger, &output)

	assert.NoError(t, err)
	assert.Contains(t, output.String(), "kind: HelmRelease")
	assert.Contains(t, output.String(), "kind: HelmRepository")
	assert.Contains(t, output.String(), "kind: Kustomization")
}

type mockPullRequest struct{}

func (m *mockPullRequest) APIObject() interface{} {
	return nil
}
func (m *mockPullRequest) Get() gitprovider.PullRequestInfo {
	return gitprovider.PullRequestInfo{
		Merged: false,
		Number: 1,
		WebURL: "https://github.com/org/repo/pull/1",
	}
}

func TestUpgrade(t *testing.T) {
	gitClient := &gitfakes.FakeGit{}
	kubeClient := makeClient(t, createSecret(), createGitRepository("my-app"))
	gitProvider := &gitprovidersfakes.FakeGitProvider{}
	gitProvider.CreatePullRequestStub = func(context.Context, gitproviders.RepoURL, gitproviders.PullRequestInfo) (gitprovider.PullRequest, error) {
		return &mockPullRequest{}, nil
	}
	logger := &loggerfakes.FakeLogger{}

	var output bytes.Buffer

	// Run upgrade!
	err := upgrade(context.TODO(), UpgradeValues{
		ConfigRepo:    "https://github.com/test/example.git",
		HeadBranch:    "upgrade-to-wge",
		BaseBranch:    "main",
		CommitMessage: "Upgrade to wge",
		Namespace:     testNamespace,
	}, kubeClient, gitClient, gitProvider, logger, &output)

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
			err := getBasicAuth(context.Background(), kubeClient, testNamespace)
			assert.Equal(t, tt.expectedErr, err)
		})
	}
}

func TestMakeHelmResources(t *testing.T) {
	tests := []struct {
		name             string
		values           []string
		expected         string
		expectedErrorStr string
	}{
		{
			name:             "error returned",
			values:           []string{"key"},
			expectedErrorStr: "failed parsing --set data: key \"key\" has no value",
		},
		{
			name:     "resources created",
			expected: "name: foo",
		},
		{
			name:     "resources created with given values",
			values:   []string{"key=val"},
			expected: "key: val",
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
			if err != nil {
				assert.Equal(t, tt.expectedErrorStr, err.Error())
			}
			assert.Contains(t, string(out), tt.expected)
		})
	}
}

//
// helpers
//

func makeClient(t *testing.T, clusterState ...runtime.Object) *kube.KubeHTTP {
	scheme := runtime.NewScheme()
	schemeBuilder := runtime.SchemeBuilder{
		corev1.AddToScheme,
		sourcev1.AddToScheme,
	}
	err := schemeBuilder.AddToScheme(scheme)
	assert.NoError(t, err)

	return &kube.KubeHTTP{
		Client: fake.NewClientBuilder().
			WithScheme(scheme).
			WithRuntimeObjects(clusterState...).
			Build(),
		ClusterName: "foo",
	}
}

func createSecret(opts ...func(*corev1.Secret)) *corev1.Secret {
	s := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weave-gitops-enterprise-credentials",
			Namespace: testNamespace,
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
			Namespace: testNamespace,
		},
	}
}
