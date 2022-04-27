package cache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/gitops-server/core/clustersmngr"
	mngr "github.com/weaveworks/weave-gitops/gitops-server/core/clustersmngr"
	v1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/rest"
)

const NamespaceStorage StorageType = "namespace"

type namespaceStore struct {
	logger      logr.Logger
	restCfg     *rest.Config
	clusterChan <-chan mngr.Cluster

	namespaces map[string][]v1.Namespace

	lock     sync.Mutex
	cancel   func()
	interval time.Duration
}

func WithNamespaceCache(cfg *rest.Config) WithCacheFunc {
	return func(l logr.Logger, syncChan chan mngr.Cluster) (StorageType, Cache) {
		return NamespaceStorage, newNamespaceStore(l, cfg, syncChan)
	}
}

func newNamespaceStore(logger logr.Logger, restCfg *rest.Config, clusterChan <-chan mngr.Cluster) *namespaceStore {
	return &namespaceStore{
		logger:      logger.WithName("namespace-cache"),
		restCfg:     restCfg,
		clusterChan: clusterChan,
		namespaces:  map[string][]v1.Namespace{},
		cancel:      nil,
		lock:        sync.Mutex{},
		interval:    DefaultPollIntervalSeconds,
	}
}

func (n *namespaceStore) List() interface{} {
	return n.namespaces
}

func (n *namespaceStore) ForceRefresh() {
	n.update(context.Background())
}

func (n *namespaceStore) Stop() {
	if n.cancel != nil {
		n.logger.Info("stopping namespace cache")
		n.cancel()
	}
}

func (n *namespaceStore) Start(ctx context.Context) {
	var newCtx context.Context

	newCtx, n.cancel = context.WithCancel(ctx)

	n.logger.Info("starting namespace cache")

	// Force load namespaces on startup.
	n.update(newCtx)

	go func() {
		ticker := time.NewTicker(n.interval * time.Second)

		defer ticker.Stop()

		for {
			select {
			case <-newCtx.Done():
				break
			case <-ticker.C:
				continue
			}

			n.update(newCtx)
		}
	}()
}

func (n *namespaceStore) update(ctx context.Context) {
	n.lock.Lock()
	defer n.lock.Unlock()

	clientsPool := clustersmngr.NewClustersClientsPool()

	if err := clientsPool.Add(clientConfig(n.restCfg), addSelf(n.restCfg)); err != nil {
		n.logger.Error(err, "unable to create client for home cluster")
	}

	if n.clusterChan != nil {
		for cluster := range n.clusterChan {
			if err := clientsPool.Add(clientConfig(n.restCfg), cluster); err != nil {
				n.logger.Error(err, "unable to create client for cluster", "cluster", cluster.Name)

				continue
			}
		}
	}

	for name, c := range clientsPool.Clients() {
		list := &v1.NamespaceList{}

		if err := c.List(ctx, list); err != nil {
			if !apierrors.IsForbidden(err) && !errors.Is(err, context.Canceled) {
				n.logger.Error(err, "unable to fetch namespaces", "cluster", name)
			}

			continue
		}

		newList := []v1.Namespace{}
		newList = append(newList, list.Items...)

		n.namespaces[name] = newList
	}
}

func clientConfig(restCfg *rest.Config) clustersmngr.ClusterClientConfig {
	return func(cluster clustersmngr.Cluster) *rest.Config {
		return restCfg
	}
}

func addSelf(restCfg *rest.Config) mngr.Cluster {
	return mngr.Cluster{
		Name:        mngr.DefaultCluster,
		Server:      restCfg.Host,
		BearerToken: restCfg.BearerToken,
		TLSConfig:   restCfg.TLSClientConfig,
	}
}
