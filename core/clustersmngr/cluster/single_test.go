package cluster

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
	"github.com/weaveworks/weave-gitops/pkg/testutils"
	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/rest"
)

func TestSingleCluster(t *testing.T) {
	config := &rest.Config{
		Host:        "my-host",
		BearerToken: "my-token",
	}

	g := NewGomegaWithT(t)

	cluster, err := NewSingleCluster("Default", config, nil)
	g.Expect(err).To(BeNil())

	g.Expect(cluster.GetName()).To(Equal("Default"))
	g.Expect(cluster.GetHost()).To(Equal(config.Host))
	g.Expect(cluster.(*singleCluster).restConfig.BearerToken).To(Equal(config.BearerToken))
}

func TestClientConfigWithUser(t *testing.T) {
	var k8sEnv *testutils.K8sTestEnv

	g := NewGomegaWithT(t)

	k8sEnv, err := testutils.StartK8sTestEnvironment([]string{
		"../../manifests/crds",
		"../../tools/testcrds",
	})
	if err != nil {
		panic(err)
	}

	defer k8sEnv.Stop()

	tests := []struct {
		name                  string
		principal             *auth.UserPrincipal
		expectedErr           error
		expectedToken         string
		expectedImpersonation *rest.ImpersonationConfig
	}{
		{
			name:                  "good user and cluster in: good config out",
			principal:             &auth.UserPrincipal{ID: "Juan", Groups: []string{"team-a"}},
			expectedErr:           nil,
			expectedToken:         "",
			expectedImpersonation: &rest.ImpersonationConfig{UserName: "Juan", Groups: []string{"team-a"}},
		},
		{
			name:                  "good token and cluster in: good config out",
			principal:             auth.NewUserPrincipal(auth.Token("some-token-i-guess")),
			expectedErr:           nil,
			expectedToken:         "some-token-i-guess",
			expectedImpersonation: nil,
		},
		{
			name:                  "No token or user Id should error",
			principal:             &auth.UserPrincipal{},
			expectedErr:           fmt.Errorf("no user ID or Token found in UserPrincipal"),
			expectedToken:         "",
			expectedImpersonation: nil,
		},
	}

	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up
			clusterName := fmt.Sprintf("clustersmngr-test-%d-%s", idx, rand.String(5))

			cluster, err := NewSingleCluster(clusterName, k8sEnv.Rest, nil)
			g.Expect(err).NotTo(HaveOccurred())
			res, err := getImpersonatedConfig(cluster.(*singleCluster).restConfig, tt.principal)

			// Test
			if tt.expectedErr != nil {
				g.Expect(err).To(Equal(tt.expectedErr))
				g.Expect(res).To(BeNil())

			} else {
				g.Expect(err).To(BeNil())

				if tt.expectedImpersonation != nil {
					g.Expect(res.Impersonate.UserName).To(Equal(tt.expectedImpersonation.UserName))
					g.Expect(res.Impersonate.Groups).To(Equal(tt.expectedImpersonation.Groups))
					g.Expect(res.BearerToken).To(Equal(k8sEnv.Rest.BearerToken))

				} else if tt.expectedToken != "" {
					g.Expect(res.BearerToken).To(Equal(tt.expectedToken))
					g.Expect(rest.ImpersonationConfig{}).To(Equal(res.Impersonate))
				}
			}
		})
	}
}
