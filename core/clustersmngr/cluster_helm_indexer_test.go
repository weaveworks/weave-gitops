package clustersmngr_test

import (
	"log"
	"testing"

	"github.com/go-logr/logr"
	. "github.com/onsi/gomega"
	"github.com/weaveworks/weave-gitops/core/clustersmngr"
	"github.com/weaveworks/weave-gitops/core/clustersmngr/clustersmngrfakes"
	"github.com/weaveworks/weave-gitops/core/nsaccess/nsaccessfakes"
	"github.com/weaveworks/weave-gitops/pkg/kube"
	"golang.org/x/net/context"
)

func TestClusterHelmIndexerTracker(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logr.Discard()
	ctx := context.Background()

	nsChecker := &nsaccessfakes.FakeChecker{}

	clustersFetcher := new(clustersmngrfakes.FakeClusterFetcher)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	clientsFactory := clustersmngr.NewClustersManager(
		clustersFetcher, nsChecker, logger, scheme, clustersmngr.NewClustersClientsPool, clustersmngr.DefaultKubeConfigOptions)
	err = clientsFactory.UpdateClusters(ctx)
	g.Expect(err).To(BeNil())

	clusterName1 := "bar"
	clusterName2 := "foo"

	c1 := makeLeafCluster(t, clusterName1)
	c2 := makeLeafCluster(t, clusterName2)

	watcher := clientsFactory.Subscribe()
	g.Expect(watcher).ToNot(BeNil())

	indexer := clustersmngr.NewClusterHelmIndexerTracker(clientsFactory)
	g.Expect(indexer.ClustersWatcher).ToNot(BeNil())

	indexer.Start(ctx)

	clusterNames := func(c []clustersmngr.Cluster) []string {
		names := []string{}
		for _, v := range c {
			names = append(names, v.Name)
		}

		return names
	}

	t.Run("indexer should be notified with two clusters added", func(t *testing.T) {
		g := NewGomegaWithT(t)
		clustersFetcher.FetchReturns([]clustersmngr.Cluster{c1, c2}, nil)

		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())
		log.Printf("indexer.Added: %+v", indexer)

		<-watcher.Updates
		g.Expect(clusterNames(indexer.Clusters)).To(Equal(clusterNames([]clustersmngr.Cluster{c1, c2})))
	})
}
