package clustersmngr_test

import (
	"fmt"
	"testing"

	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/util/rand"
	"k8s.io/client-go/rest"

	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/pkg/server/auth"
)

func TestClientConfigWithUser(t *testing.T) {
	g := NewGomegaWithT(t)

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
			expectedErr:           fmt.Errorf("No user ID or Token found in UserPrincipal."),
			expectedToken:         "",
			expectedImpersonation: nil,
		},
	}

	for idx, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up
			clusterName := fmt.Sprintf("clustersmngr-test-%d-%s", idx, rand.String(5))

			clusterCfgFunc := clustersmngr.ClientConfigWithUser(tt.principal)

			cluster := makeLeafCluster(t, clusterName)

			// Run
			res, err := clusterCfgFunc(cluster)

			// Test
			if tt.expectedErr != nil {
				g.Expect(err).To(Equal(tt.expectedErr))
				g.Expect(res).To(BeNil())

			} else {
				g.Expect(err).To(BeNil())

				if tt.expectedImpersonation != nil {
					g.Expect(res.Impersonate.UserName).To(Equal(tt.expectedImpersonation.UserName))
					g.Expect(res.Impersonate.Groups).To(Equal(tt.expectedImpersonation.Groups))
					g.Expect(res.BearerToken).To(Equal(cluster.BearerToken))

				} else if tt.expectedToken != "" {
					g.Expect(res.BearerToken).To(Equal(tt.expectedToken))
					g.Expect(rest.ImpersonationConfig{}).To(Equal(res.Impersonate))
				}

				// Expect flowcontrol to be active, so no explicitly set rate limit
				g.Expect(res.QPS).To(BeEquivalentTo(-1))
				g.Expect(res.Burst).To(Equal(-1))
			}
		})
	}
}
