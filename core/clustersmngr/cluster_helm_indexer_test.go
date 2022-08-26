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
	clusterNames    []string
}

func newClusterHelmIndexerTracker(cf clustersmngr.ClientsFactory) *ClusterHelmIndexerTracker {
	return &ClusterHelmIndexerTracker{
		clustersWatcher: cf.Subscribe(),
	}
}

func (i *ClusterHelmIndexerTracker) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case updates := <-i.clustersWatcher.Updates:
				for _, added := range updates.Added {
					if i.contains(added) == false {
						i.clusterNames = append(i.clusterNames, added.Name)
					}
				}

				for index, removed := range updates.Removed {
					if i.contains(removed) {
						i.clusterNames = append(i.clusterNames[:index], i.clusterNames[index+1:]...)
					}
				}
			}
		}
	}()
}

// contains returns true if the given cluster is in the list of cluster names.
func (i *ClusterHelmIndexerTracker) contains(cluster clustersmngr.Cluster) bool {
	for _, n := range i.clusterNames {
		if cluster.Name == n {
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

	clientsFactory := clustersmngr.NewClientFactory(clustersFetcher, nsChecker, logger, scheme, clustersmngr.NewClustersClientsPool)
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

	t.Run("indexer should be notified with two clusters added", func(t *testing.T) {
		g := NewGomegaWithT(t)
		clustersFetcher.FetchReturns([]clustersmngr.Cluster{c1, c2}, nil)

		g.Expect(clientsFactory.UpdateClusters(ctx)).To(Succeed())
		log.Printf("indexer.Added: %+v", indexer.clusterNames)

		<-watcher.Updates
		g.Expect(indexer.clusterNames).To(Equal([]string{"bar", "foo"}))
	})
}
