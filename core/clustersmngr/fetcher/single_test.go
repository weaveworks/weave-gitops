package fetcher_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clusters"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/fetcher"
	"k8s.io/client-go/rest"
)

func TestSingleFetcher(t *testing.T) {
	config := &rest.Config{
		Host:        "my-host",
		BearerToken: "my-token",
	}

	g := NewGomegaWithT(t)

	cluster, err := clusters.NewSingleCluster("Default", config, nil)
	g.Expect(err).To(BeNil())

	fetcher, err := fetcher.NewSingleClusterFetcher(cluster)
	g.Expect(err).To(BeNil())

	clusters, err := fetcher.Fetch(context.TODO())
	g.Expect(err).To(BeNil())

	g.Expect(clusters[0].GetName()).To(Equal("Default"))
	g.Expect(clusters[0].GetHost()).To(Equal(config.Host))
}
