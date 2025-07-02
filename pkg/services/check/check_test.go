package check_test

import (
	"errors"
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/version"
	fakediscovery "k8s.io/client-go/discovery/fake"
	fakeclientset "k8s.io/client-go/kubernetes/fake"
	kubetesting "k8s.io/client-go/testing"

	"github.com/weaveworks/weave-gitops/pkg/services/check"
)

func TestKubernetesVersionWithError(t *testing.T) {
	g := NewWithT(t)

	expectedError := errors.New("an error occurred")
	fakeClient := fakeclientset.NewClientset()
	fakeClient.Discovery().(*fakediscovery.FakeDiscovery).PrependReactor("*", "*", func(action kubetesting.Action) (handled bool, ret runtime.Object, err error) {
		return true, nil, expectedError
	})

	res, err := check.KubernetesVersion(fakeClient.Discovery())

	g.Expect(err).To(MatchError(ContainSubstring(expectedError.Error())))
	g.Expect(res).To(BeEmpty())
}

func TestKubernetesVersion(t *testing.T) {
	tests := []struct {
		name          string
		serverVersion string
		expectedErr   string
		expectedRes   string
	}{
		{
			name:          "server version satisfies constraint",
			serverVersion: "v1.28.4",
			expectedRes:   `^✔ Kubernetes 1.28.4 >=1.`,
		},
		{
			name:          "server version too low",
			serverVersion: "v1.20.5",
			expectedErr:   `✗ kubernetes version v1\.20\.5 does not match >=1\.`,
		},
		{
			name:          "server version not semver compliant",
			serverVersion: "1.x",
			expectedErr:   `"1.x".*invalid semantic version`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := NewWithT(t)
			client := fakeclientset.NewClientset()
			fakeDiscovery, ok := client.Discovery().(*fakediscovery.FakeDiscovery)
			if !ok {
				t.Fatalf("couldn't convert Discovery() to *FakeDiscovery")
			}

			fakeDiscovery.FakedServerVersion = &version.Info{
				GitVersion: tt.serverVersion,
			}

			res, err := check.KubernetesVersion(client.Discovery())

			if tt.expectedErr != "" {
				g.Expect(err).To(MatchError(MatchRegexp(tt.expectedErr)))
			}

			g.Expect(res).To(MatchRegexp(tt.expectedRes))
		})
	}
}
