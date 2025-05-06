package fetcher_test

import (
	"testing"

	. "github.com/onsi/gomega"
	"k8s.io/client-go/rest"

	"github.com/weaveworks/weave-gitops/core/clustersmngr/cluster"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/fetcher"
	"github.com/weaveworks/weave-gitops/pkg/kube"
)

func TestSingleFetcher(t *testing.T) {
	config := &rest.Config{
		Host:        "my-host",
		BearerToken: "my-token",
	}

	g := NewGomegaWithT(t)

	cluster, err := cluster.NewSingleCluster("Default", config, nil, kube.UserPrefixes{})
	g.Expect(err).To(BeNil())

	fetcher := fetcher.NewSingleClusterFetcher(cluster)

	clusters, err := fetcher.Fetch(t.Context())
	g.Expect(err).To(BeNil())

	g.Expect(clusters[0].GetName()).To(Equal("Default"))
	g.Expect(clusters[0].GetHost()).To(Equal(config.Host))
}
