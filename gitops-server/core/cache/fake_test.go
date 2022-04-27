package cache_test

import (
	"context"

	"github.com/go-logr/logr"
	"github.com/weaveworks/weave-gitops/gitops-server/core/cache"
	mngr "github.com/weaveworks/weave-gitops/gitops-server/core/clustersmngr"
	"k8s.io/client-go/rest"
)

func withFakeClusterCache(cfg *rest.Config) cache.WithCacheFunc {
	return func(l logr.Logger, syncChan chan mngr.Cluster) (cache.StorageType, cache.Cache) {
		return "FakeClusterStorage", &fakeStore{syncChan, cfg, nil}
	}
}

type fakeStore struct {
	syncChan chan<- mngr.Cluster
	cfg      *rest.Config
	cancel   func()
}

func (f *fakeStore) List() interface{} {
	return nil
}

func (f *fakeStore) ForceRefresh() {
	// shrug_emoji
}

func (f *fakeStore) Stop() {
	// shrug_emoji
}

func (f *fakeStore) Start(ctx context.Context) {
	go func() {
		clusters := []mngr.Cluster{
			{
				Name:        "foo",
				Server:      f.cfg.Host,
				BearerToken: f.cfg.BearerToken,
				TLSConfig:   f.cfg.TLSClientConfig,
			},
			{
				Name:        "bar",
				Server:      f.cfg.Host,
				BearerToken: f.cfg.BearerToken,
				TLSConfig:   f.cfg.TLSClientConfig,
			},
		}

		if f.syncChan != nil {
			for _, cluster := range clusters {
				f.syncChan <- cluster
			}
		}

		close(f.syncChan)
	}()
}
