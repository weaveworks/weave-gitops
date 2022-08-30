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

type ClusterHelmIndexerTracker struct {
	clustersWatcher *clustersmngr.ClustersWatcher
	clusters        []clustersmngr.Cluster
}

func newClusterHelmIndexerTracker(cf clustersmngr.ClientsFactory) *ClusterHelmIndexerTracker {
	return &ClusterHelmIndexerTracker{
		clustersWatcher: cf.Subscribe(),
	}
}

// Start the indexer and wait for cluster updates notifications.
func (i *ClusterHelmIndexerTracker) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case updates := <-i.clustersWatcher.Updates:
				for _, added := range updates.Added {
					if i.contains(added) == false {
						i.clusters = append(i.clusters, added)
					}
				}

				for index, removed := range updates.Removed {
					if i.contains(removed) {
						i.clusters = append(i.clusters[:index], i.clusters[index+1:]...)
					}
				}
			}
		}
	}()
}

// contains returns true if the given cluster is in the list of cluster names.
func (i *ClusterHelmIndexerTracker) contains(cluster clustersmngr.Cluster) bool {
	for _, n := range i.clusters {
		if cluster.Name == n.Name && cluster.Server == n.Server {
			return true
		}
	}

	return false
}

func TestClusterHelmIndexerTracker(t *testing.T) {
	g := NewGomegaWithT(t)
	logger := logr.Discard()
	ctx := context.Background()

	nsChecker := &nsaccessfakes.FakeChecker{}

	clustersFetcher := new(clustersmngrfakes.FakeClusterFetcher)

	scheme, err := kube.CreateScheme()
	g.Expect(err).To(BeNil())

	clientsFactory := clustersmngr.NewClientFactory(
		clustersFetcher, nsChecker, logger, scheme, clustersmngr.NewClustersClientsPool)
	err = clientsFactory.UpdateClusters(ctx)
	g.Expect(err).To(BeNil())

	clusterName1 := "bar"
	clusterName2 := "foo"

	c1 := makeLeafCluster(t, clusterName1)
	c2 := makeLeafCluster(t, clusterName2)

	watcher := clientsFactory.Subscribe()
	g.Expect(watcher).ToNot(BeNil())

	indexer := newClusterHelmIndexerTracker(clientsFactory)
	g.Expect(indexer.clustersWatcher).ToNot(BeNil())

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
		g.Expect(clusterNames(indexer.clusters)).To(Equal(clusterNames([]clustersmngr.Cluster{c1, c2})))
	})
}
