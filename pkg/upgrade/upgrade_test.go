package upgrade

import (
	"bytes"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/fake"
)

func TestPreflightCheck(t *testing.T) {
	tests := []struct {
		name             string
		result           string
		clusterState     []runtime.Object
		err              error
		expected         string
		expectedErrorStr string
	}{
		{
			name:             "error returned",
			err:              errors.New("something went wrong"),
			expectedErrorStr: "failed to get entitlement: secrets \"weave-gitops-enterprise-credentials\" not found",
		},
		{
			name:         "preflight check pass",
			err:          nil,
			clusterState: []runtime.Object{createSecret()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset(tt.clusterState...)
			w := new(bytes.Buffer)
			_, err := getEntitlement(clientset)
			assert.Equal(t, tt.expected, w.String())
			if err != nil {
				assert.EqualError(t, err, tt.expectedErrorStr)
			}
		})
	}
}

func TestGetGithubRepoPath(t *testing.T) {
	repoPath, err := getRepoOrgAndName("git@github.com:ww/repo.git")
	assert.NoError(t, err)
	assert.Equal(t, "ww/repo", repoPath)

	repoPath, err = getRepoOrgAndName("https://github.com/ww/repo.git")
	assert.NoError(t, err)
	assert.Equal(t, "ww/repo", repoPath)

}

func createSecret() *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "weave-gitops-enterprise-credentials",
			Namespace: "wego-system",
		},
		Type: "Opaque",
		Data: map[string][]byte{"entitlement": []byte("foo")},
	}
}
