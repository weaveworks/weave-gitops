package clustersmngr

import (
	"golang.org/x/net/context"
)

type ClusterHelmIndexerTracker struct {
	ClustersWatcher *ClustersWatcher
	Clusters        []Cluster
}

func NewClusterHelmIndexerTracker(cf ClustersManager) *ClusterHelmIndexerTracker {
	return &ClusterHelmIndexerTracker{
		ClustersWatcher: cf.Subscribe(),
	}
}

// Start the indexer and wait for cluster updates notifications.
func (i *ClusterHelmIndexerTracker) Start(ctx context.Context) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case updates := <-i.ClustersWatcher.Updates:
				for _, added := range updates.Added {
					if !i.contains(added) {
						i.Clusters = append(i.Clusters, added)
					}
				}

				for index, removed := range updates.Removed {
					if i.contains(removed) {
						i.Clusters = append(i.Clusters[:index], i.Clusters[index+1:]...)
					}
				}
			}
		}
	}()
}

// contains returns true if the given cluster is in the list of cluster names.
func (i *ClusterHelmIndexerTracker) contains(cluster Cluster) bool {
	for _, n := range i.Clusters {
		if cluster.Name == n.Name && cluster.Server == n.Server {
			return true
		}
	}

	return false
}
