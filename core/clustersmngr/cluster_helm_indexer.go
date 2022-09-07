package clustersmngr

import (
	"context"

	"github.com/weaveworks/weave-gitops/pkg/helm"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/cache"
	"github.com/weaveworks/weave-gitops/pkg/helm/watcher/controller"
	"k8s.io/apimachinery/pkg/types"
)

type ClusterHelmIndexerTracker struct {
	Cache           cache.Cache
	RepoManager     helm.RepoManager
	Namespace       string
	ClusterWatchers map[Cluster]*controller.HelmWatcherReconciler
}

func NewClusterHelmIndexerTracker(c cache.Cache, rm helm.RepoManager, ns string) *ClusterHelmIndexerTracker {
	return &ClusterHelmIndexerTracker{
		Cache:           c,
		RepoManager:     rm,
		Namespace:       ns,
		ClusterWatchers: make(map[Cluster]*controller.HelmWatcherReconciler),
	}
}

func (c *ClusterHelmIndexerTracker) newIndexer(c Cluster) error {
	clientForCluster := c.clientPool.Cluster(c.Name)
	if err = (&controller.HelmWatcherReconciler{
		Client:                cl,
		Cache:                 c.Cache,
		RepoManager:           c.repoManager,
		Scheme:                scheme,
		ExternalEventRecorder: eventRecorder,
	}).SetupWithManager(mgr); err != nil {
		ctrl.Log.Error(err, "unable to create controller", "controller", "HelmWatcherReconciler")
		return err
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		return err
	}
}

// Start the indexer and wait for cluster updates notifications.
func (i *ClusterHelmIndexerTracker) Start(ctx context.Context, cw *ClusterWatcher) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case updates := <-cw.Updates:
				for _, added := range updates.Added {
					_, ok := i.Clusters[added]
					if !ok {
						indexer := newIndexer(i.Cache, i.clientsPool.Client(added.Name))
						i.Clusters[added] = indexer
						indexer.Start()
					}
				}

				for index, removed := range updates.Removed {
					watcher, ok := i.Clusters[removed]
					if ok {
						watcher.Stop()
						cache.DeleteCluster(types.NamespacedName{Name: removed.Name, Namespace: i.Namespace})
					}
				}
			}
		}
	}()
}
