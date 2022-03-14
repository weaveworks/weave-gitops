package multicluster_test

import (
	"context"
	"testing"

	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/multicluster"
	"k8s.io/client-go/rest"
)

func TestSingleFetcher(t *testing.T) {
	config := &rest.Config{
		Host:        "my-host",
		BearerToken: "my-token",
	}

	g := NewGomegaWithT(t)

	fetcher, err := multicluster.NewSingleClustersFetcher(config, "default")
	g.Expect(err).To(BeNil())

	clusters, err := fetcher.Fetch(context.TODO())
	g.Expect(err).To(BeNil())

	g.Expect(clusters[0].Name).To(Equal("Default"))
	g.Expect(clusters[0].Server).To(Equal(config.Host))
	g.Expect(clusters[0].BearerToken).To(Equal(config.BearerToken))
}
